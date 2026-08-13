# ZEP-0026: Enhanced State Management

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [User Stories (Optional)](#user-stories-optional)
    - [Story 1](#story-1)
    - [Story 2](#story-2)
    - [Story 3](#story-3)
    - [Story 4](#story-4)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
  - [`DeployedPackage` and `DeployedComponent` changes](#deployedpackage-and-deployedcomponent-changes)
    - [Add `Events`](#add-events)
    - [Reading status from `Events`](#reading-status-from-events)
    - [`PackageEvent` retention](#packageevent-retention)
    - [`DeployConfig`, `RemoveConfig`, and `ConfigDigest`](#deployconfig-removeconfig-and-configdigest)
  - [Behavior Changes](#behavior-changes)
    - [Reusing the successful deploy config](#reusing-the-successful-deploy-config)
    - [Chart, Component, and Image Reconciliation](#chart-component-and-image-reconciliation)
      - [Chart reconciliation](#chart-reconciliation)
      - [Component reconciliation](#component-reconciliation)
      - [Image reconciliation](#image-reconciliation)
    - [Graceful Cancellation Handling](#graceful-cancellation-handling)
  - [Test Plan](#test-plan)
      - [Prerequisite testing updates](#prerequisite-testing-updates)
      - [Unit tests](#unit-tests)
      - [e2e tests](#e2e-tests)
  - [Graduation Criteria](#graduation-criteria)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
  - [Version Skew Strategy](#version-skew-strategy)
- [Implementation History](#implementation-history)
- [Drawbacks](#drawbacks)
- [Alternatives](#alternatives)
  - [Flat `PackageStatus` / `ComponentStatus` fields](#flat-packagestatus--componentstatus-fields)
  - [Split `PackagePhase` / `PackageOutcome`](#split-packagephase--packageoutcome)
- [Infrastructure Needed (Optional)](#infrastructure-needed-optional)
<!-- /toc -->

## Summary

This ZEP proposes to improve Zarf's tracking of package and component state - a history of deploy/remove events, source, and configuration - so that Zarf and its library users can make smarter decisions about a deployed package: what changed, where it came from, and what to reconcile when redeploying or removing it.

## Motivation

The following use cases aren't addressed today because of limitations with Zarf's tracking of packages/components:

1. If a package update has removed a component (or a chart) from one version to the next, Zarf will lose track of that component during an upgrade
1. There is currently no way to determine where a deployed package came from, which prevents redeploying it from its original source
1. There is no consistent way to see overall package status - especially in remove states and failure edge cases
1. There is no consistent way to compare the configuration a package was deployed with against a new configuration to determine if they differ

There are existing issues tracking against the above such as https://github.com/zarf-dev/zarf/issues/2992, https://github.com/zarf-dev/zarf/issues/4182, and https://github.com/zarf-dev/zarf/issues/4969.

### Goals

- Improve Zarf's internal reconcile consistency between package updates
- Enable the following workflows
  - Redeploy (when the source is online - e.g. https/oci)
  - Report deploy state / progress
  - Report remove state / progress
  - Resume multi-package deployment (identified by package digest + config digest)
- Match Kubernetes/Helm conventions where we can

### Non-Goals

- Usurp Helm's lifecycle management of charts/objects deployed by Zarf
- Rollback to a previous Zarf package on failure (would require previous full package source)
- Record full history of previously installed versions of a given Zarf package
- Continuously reconcile state to ensure correctness at all times (e.g. recovery after a hard stop)

## Proposal

To meet these goals, this proposal adds an event history and the last successful `DeployConfig` to `DeployedPackage`, plus a single last-event snapshot to each `DeployedComponent`. It also changes deploy/remove behavior so both stay accurate over time, and reorganizes `packager.DeployOptions`/`packager.RemoveOptions` so a subset of each can be stored and hashed for comparison. See [Design Details](#design-details) for the specifics of each change.

- `DeployedPackage` gains an append-only `Events` history recording deploy/remove lifecycle transitions - each transition's type (`Deploy`/`Remove`), outcome (`InProgress`/`Succeeded`/`Failed`/`Cancelled`), timestamp, package version, flavor and digest, and config digest. Deploy events additionally carry the source used.
- `DeployedPackage` stores the normalized `DeployConfig` from its latest successful deployment. Failed and cancelled attempts remain visible while retained in `Events` but do not replace it, and full configs are not copied into event history.
- `DeployedComponent` gains a `LastEvent` field recording only the most recent deploy/remove attempt against that specific component, since components mostly mirror the package's own timeline and rarely need a history of their own.
- Package status is read via `LatestEvent()`, component status by reading `LastEvent` directly - neither is a separately stored field, so neither can drift from what actually happened. See [Reading status from `Events`](#reading-status-from-events).
- On a graceful stop, Zarf appends a `Cancelled` package event and sets the component `LastEvent` to `Cancelled`, rather than leaving either stuck at `InProgress`. See [Graceful Cancellation Handling](#graceful-cancellation-handling).
- `zarf package deploy` and `zarf package remove` gain an opt-in `--prune[=<scopes>]` flag. Scopes are selected with a comma-separated list, and a bare `--prune` defaults to all scopes applicable to the command.
- `DeployOptions`' config-affecting fields move into an embedded `DeployConfig`; `RemoveOptions`' equivalent fields move into an embedded `RemoveConfig`. Every deploy or remove `PackageEvent` records a standard `sha256:<hex>` config digest. `zarf package deploy --reuse-config` seeds the base config with the stored successful `DeployConfig`, then merges any inputs from the new invocation over it.

### User Stories (Optional)

#### Story 1

**As** a package deployer, when I deploy a new version of a package that has removed a component or chart from a previous version, **I want** Zarf to detect that the component is no longer present and clean up the orphaned resources during `zarf package deploy --prune`, **so that** they are not left behind, untracked, in the cluster.

#### Story 2

**As** a package deployer, **I want** to know where a currently deployed package came from (e.g. the OCI reference or URL it was sourced from) and/or reuse its last successful configuration **so that** I can redeploy it later, such as to recover from an incident, without having to keep separate records of the source and deployment inputs myself.

#### Story 3

**As** a package deployer or operator, **I want** a single, reliable way to check the overall status of a package - including while it is being removed or after a removal has failed - **so that** I don't have to infer status by piecing together logs or inspecting individual cluster resources.

#### Story 4

**As** a package deployer or SDK integrator adopting an existing package, **I want** to read its last successful deployment configuration and compare it with a candidate configuration **so that** I can hydrate imported state and tell whether a new deployment would actually change anything.

### Risks and Mitigations

- **`--prune` incorrectly removes resources still in use.** If the orphan-detection logic incorrectly identifies a component as removed, or the cross-package image reference counting has a bug, `--prune` could delete a chart or image that's still needed - including one shared with an unrelated package. This is the only irreversible/destructive behavior this proposal adds.
  - *Mitigation:* `--prune` is opt-in and off by default on both `zarf package deploy` and `zarf package remove`. Users can limit pruning to explicit scopes, and anyone who hits a correctness issue can stop passing the flag to get today's behavior.
- **The latest `PackageEvent`/`LastEvent` can be left at `InProgress`.** If a deploy or remove is interrupted, the most recent package `PackageEvent` or component `LastEvent` can be left indefinitely showing an operation that's no longer actually running.
  - *Mitigation:* see [Graceful Cancellation Handling](#graceful-cancellation-handling) - on a controlled stop, Zarf appends a `Cancelled` package event and sets in-progress component events to `Cancelled` before exiting. An ungraceful termination can still leave `InProgress` as the latest event; that residual risk is accepted because Zarf is not a continuously reconciling controller.
- **Unbounded `Events` growth.** `DeployedPackage` is stored in a Kubernetes Secret; an ever-growing `Events` list risks approaching the Secret size limit on a long-lived, frequently-redeployed package. `DeployedComponent.LastEvent` is a single value, not a list, so it doesn't carry this risk.
  - *Mitigation:* `Events` is capped at a fixed number of most-recent entries (see [`PackageEvent` retention](#packageevent-retention)), evicting the oldest entry once the cap is exceeded. Also, only one successful `DeployConfig` is stored at the package level rather than one copy per event.
- **Persisting the successful `DeployConfig` creates another copy of sensitive deployment inputs.** Helm release Secrets already retain values passed to charts, but package-level variables and values can also feed Zarf actions, manifests, cluster creation, or host-level components and may never reach Helm. Persisting the normalized override layer in the `DeployedPackage` Secret centralizes those inputs for anyone who can read that Secret or its backups. Importing them into another provider (such as OpenTofu) would also copy them into that provider's state under a different access boundary, and `--reuse-config` can replay credentials or action inputs long after their original use.
  - *Mitigation:* redact the snapshot from default human-readable inspect, status, confirmation, and state-related log output and document that operators must protect Zarf state Secrets with appropriate RBAC and encryption at rest.  Any other providers using Zarf as a library must also mark corresponding state attributes as sensitive. Zarf cannot reliably filter only values used by cluster resources because values are package-scoped and can be consumed indirectly, so the residual exposure of action and host-level values must be accepted and documented.
- **Config digest false change reports from numeric formatting.** `ConfigDigest` is computed by JSON-marshaling user-supplied values. Semantically-identical values that are formatted differently (e.g. `5` vs `5.0` in a values file) could produce different digests, causing Zarf to report a configuration change when there isn't a meaningful one.
  - *Mitigation:* document this as a known limitation of the initial implementation; if it proves disruptive in practice, add a canonicalization pass (e.g. normalize numeric types before hashing) in a follow-up.
- **Config digest generation can change between CLI versions.** A future change to config normalization or serialization could make the same logical config produce a different digest. Because retained events can have been written by several CLI versions, `DeployedPackage.CLIVersion` cannot reliably identify the rules used for each digest.
  - *Mitigation:* config digests are comparison fingerprints, not permanent content identifiers. A mismatch means "changed or unknown," so a conservative mismatch across versions is safe. For deploy comparisons, the retained `DeployConfig` can be re-hashed using the current rules; historical event-only comparisons remain best-effort.
- **A `Source` could persist embedded credentials.** A source string like `https://user:pass@host/...` recorded on a deploy `PackageEvent` would be stored as-is in the `DeployedPackage` secret and could resurface in `zarf package inspect` output or logs. In practice this should be rare - most sources are `oci://` or local tarballs, and Zarf already supports `.netrc` for credential storage, so embedding credentials in the source URI is uncommon.
  - *Mitigation:* document that credentials should be provided via `.netrc` rather than embedded in the source string; no code-level scrubbing planned given how rare this path is.

## Design Details

### `DeployedPackage` and `DeployedComponent` changes

`DeployedPackage.CLIVersion` is existing provenance identifying the Zarf CLI version recorded with the deployed package; it is not a schema version and is not necessarily refreshed by every state update. This proposal does not add a schema version. In particular, `CLIVersion` does not version individual `PackageEvent` entries: a retained event history can contain entries from several CLI versions, and a newer unsuccessful attempt can update package state while preserving an older successful `DeployConfig`. Readers therefore must not use the top-level `CLIVersion` to infer how a particular config digest was generated and must accept the "changed or unknown" mitigation mentioned above.

#### Add `Events`

`DeployedPackage` gains an `Events` list and `DeployedComponent` gains a single `LastEvent`. The existing `ComponentStatus` field is deprecated in favor of `LastEvent`, but is not dropped immediately - see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy) for the deprecation window:

```go
// EventType is the kind of lifecycle operation a PackageEvent or ComponentEvent records.
type EventType string

const (
	EventTypeDeploy EventType = "Deploy"
	EventTypeRemove EventType = "Remove"
)

// EventOutcome is the result of a PackageEvent or ComponentEvent.
type EventOutcome string

const (
	EventOutcomeUnknown    EventOutcome = "Unknown" // never stored; returned by helpers when there's no history
	EventOutcomeInProgress EventOutcome = "InProgress"
	EventOutcomeSucceeded  EventOutcome = "Succeeded"
	EventOutcomeFailed     EventOutcome = "Failed"
	EventOutcomeCancelled  EventOutcome = "Cancelled"
)

// PackageEvent records one lifecycle transition in a deploy or remove attempt.
type PackageEvent struct {
	Type      EventType    `json:"type"`
	Outcome   EventOutcome `json:"outcome"`
	Timestamp time.Time    `json:"timestamp"`
	// Version is the package's Metadata.Version at the time of this event.
	Version string `json:"version,omitempty"`
	// Flavor is the package's Build.Flavor at the time of this event.
	Flavor string `json:"flavor,omitempty"`
	// Digest is the package's content digest (as returned by pkgLayout.Digest()) at the time
	// of this event.
	Digest string `json:"digest,omitempty"`
	// Source is the source the package was deployed from (e.g. oci:// URL, tarball path).
	// Only set on Deploy events.
	Source string `json:"source,omitempty"`
	// ConfigDigest is a deterministic digest of the DeployConfig (Deploy events) or
	// RemoveConfig (Remove events) used for this operation.
	ConfigDigest digest.Digest `json:"configDigest,omitempty"`
}

// ComponentEvent records the most recent deploy or remove attempt against a single component.
// Unlike PackageEvent, it carries no Version, Flavor, Digest, Source, or ConfigDigest - those describe
// the package as a whole, not an individual component.
type ComponentEvent struct {
	Type      EventType    `json:"type"`
	Outcome   EventOutcome `json:"outcome"`
	Timestamp time.Time    `json:"timestamp"`
}
```

Each deploy/remove attempt appends an `InProgress` `PackageEvent`; completion appends a terminal event of the same type. `LatestEvent()` therefore returns current status without finding and mutating an earlier entry. `DeployedComponent.LastEvent` remains a single value and is overwritten as that component progresses.

`Version`, `Flavor`, `Digest`, and `ConfigDigest` are set on every `PackageEvent`, deploy or remove alike, since both operations act on a specific package version and both have a config-affecting subset of options (`DeployConfig`/`RemoveConfig` - see [`DeployConfig`, `RemoveConfig`, and `ConfigDigest`](#deployconfig-removeconfig-and-configdigest)). `Version`/`Flavor`/`Digest` give a history of what versions/flavors/digests were deployed and removed over time, alongside `DeployedPackage`'s own top-level `Digest` field (which still tracks only the current one, unchanged by this proposal). `Source` stays Deploy-only, since a remove doesn't have a source of its own - it operates on whatever's already deployed.

`DeployedPackage` gains `Events []PackageEvent` - a history. `DeployedComponent` gains a single `LastEvent ComponentEvent` - not a list, and not the same type as the package's events:

```go
// LastEvent is the outcome of the most recent deploy or remove attempt against this component,
// which can be ahead of, behind, or independent of the package's own latest PackageEvent - e.g.
// a package still shows Deploy/InProgress overall while component 1 is already Deploy/Succeeded,
// component 2 is still Deploy/InProgress, and component 3 hasn't been reached yet at all (no
// DeployedComponent recorded).
LastEvent ComponentEvent `json:"lastEvent,omitempty"`
```

A component's outcomes generally mirror the package's own `Events` timeline one-for-one, so a full per-component history would mostly duplicate it. A component only needs one fact from its history: what happened the last time that component was touched. `LastEvent` gives that directly. `Type` and `Outcome` reuse the same `EventType`/`EventOutcome` vocabulary as `PackageEvent`, so there's nothing new to keep in sync between the two levels; `ComponentEvent` just omits the fields (`Version`, `Flavor`, `Digest`, `Source`, `ConfigDigest`) that only make sense at the package level.

#### Reading status from `Events`

`DeployedPackage` gains an accessor method so library users don't have to hand-roll list-walking logic to answer a basic question - what happened most recently:

```go
// LatestEvent returns the most recently recorded PackageEvent, or false if there aren't any yet.
func (d *DeployedPackage) LatestEvent() (PackageEvent, bool)

// LastSuccessfulDeploy returns the most recent retained successful Deploy event, or false if
// there isn't one.
func (d *DeployedPackage) LastSuccessfulDeploy() (PackageEvent, bool)
```

`LastSuccessfulDeploy()` scans from newest to oldest, skipping later failed or cancelled deploys and remove events. This gives library users the `Source`, package metadata, and `ConfigDigest` associated with the last successful deployment without duplicating those fields at the top level. Because it operates on the retained event history, `false` means there is no matching retained event, not necessarily that the package has never deployed successfully.

`DeployedComponent.LastEvent` becomes the source of truth in place of `ComponentStatus` - `Type`+`Outcome` say whether that component is deploying, removing, or at a terminal state. `ComponentStatus` is deprecated but continues to be populated alongside `LastEvent` for the deprecation window described in [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy).

#### `PackageEvent` retention

Since `DeployedPackage` is stored inside a Kubernetes Secret, `Events` can't grow without bound. The list is capped at a fixed number of most-recent entries (e.g. `10`); once the cap is exceeded, the oldest entry is dropped. `DeployedComponent.LastEvent` is a single value, so each deploy or remove transition overwrites it.

#### `DeployConfig`, `RemoveConfig`, and `ConfigDigest`

Comparing a candidate deployment's configuration against what's already deployed - or retaining the inputs needed to repeat a successful deployment - requires distinguishing fields that affect the resulting deployment from fields that only affect how the deploy operation runs. Today `packager.DeployOptions` mixes both. This proposal splits it as follows:

- **Config** (affects the resulting deployment, and should be part of the digest): `SetVariables`, `Values` (`value.Values`), `NamespaceOverride`, `ValuesOverridesMap`, `OptionalComponents`, and `Connected`
- **Imperative** (affects only the deploy operation, and stays out of the digest): everything else, e.g. `Timeout`, `Retries`, `OCIConcurrency`, `IsInteractive`, `ForceConflicts`, `AdoptExistingResources`, `SkipVersionCheck`, `ReuseConfig`, `RemoteOptions`, and the Zarf init state used to configure a cluster (`GitServer`, `RegistryInfo`, `ArtifactServer`, `AgentTLS`, `AgentMutationPolicy`, `StorageClass`, `InjectorPort`)

The same split applies to `packager.RemoveOptions`, which today also mixes `Values` and `NamespaceOverride` in with imperative fields (`Cluster`, `Timeout`, `SkipVersionCheck`) - Zarf Values are usable in `onRemove` actions, so they affect what a remove actually does, not just how it runs. `RemoveOptions`' config subset is smaller than `DeployOptions`': it has no `SetVariables` or `ValuesOverridesMap` to begin with, since those only apply to Helm chart installs.

The config fields move into new `DeployConfig` and `RemoveConfig` types, embedded in `DeployOptions` and `RemoveOptions` respectively. `DeployConfig` is also the persisted representation used by `DeployedPackage`:

```go
// DeployConfig holds the subset of deploy options that affect the resulting deployment,
// as opposed to options that only affect how the deploy operation itself runs.
type DeployConfig struct {
	SetVariables       map[string]string `json:"setVariables,omitempty"`
	Values             value.Values      `json:"values,omitempty"`
	NamespaceOverride  string            `json:"namespaceOverride,omitempty"`
	ValuesOverridesMap ValuesOverrides   `json:"valuesOverridesMap,omitempty"`
	OptionalComponents []string          `json:"optionalComponents,omitempty"`
	Connected          bool              `json:"connected,omitempty"`
}

// Digest returns a deterministic SHA-256 digest of the deploy config.
func (c DeployConfig) Digest() (digest.Digest, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return digest.FromBytes(b), nil
}

// RemoveConfig holds the subset of remove options that affect the resulting removal (via
// onRemove actions), as opposed to options that only affect how the remove operation runs.
type RemoveConfig struct {
	Values            value.Values `json:"values,omitempty"`
	NamespaceOverride string       `json:"namespaceOverride,omitempty"`
}

// Digest returns a deterministic SHA-256 digest of the remove config.
func (c RemoveConfig) Digest() (digest.Digest, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return digest.FromBytes(b), nil
}
```

`DeployedPackage` stores the latest successful snapshot as a pointer so package state written before this proposal can be distinguished from a successful deployment whose normalized config is empty:

```go
// DeployConfig is the normalized configuration used by the latest successful deployment.
DeployConfig *DeployConfig `json:"deployConfig,omitempty"`
```

`encoding/json` sorts map keys alphabetically at every nesting level, so `json.Marshal` over either type is deterministic for the map-shaped fields they contain. `OptionalComponents` is the normalized, deduplicated, and sorted optional-component input supplied to the deploy process, rather than the full component set selected after applying package `only` filters such as local OS or cluster architecture. This keeps the persisted config directly reusable while the incoming package and current environment determine filter-based selection during each deployment. Input files and their layering are not persisted; `DeployConfig` contains the normalized explicit inputs produced after they are parsed and merged. Package and chart defaults are also not stored in `DeployConfig`; the package digest identifies that packaged content which (unless the source shifts) is re-retrievable through `Source`.

Config digests use the standard `github.com/opencontainers/go-digest` representation (`sha256:<hex>`). They deliberately carry no separate Zarf format version. Digests are compared only within the same operation type: deploy config to deploy config or remove config to remove config. Equality is authoritative only when both digests were produced under compatible normalization and serialization rules; a mismatch means the config changed or the rules did. Changing those rules should be intentional and covered by deterministic test fixtures, but will not require compatibility code unless stable cross-version comparison becomes a concrete requirement in the future.

Before changing anything, `packager.Deploy` resolves all config inputs, takes an independent snapshot of the resulting `DeployConfig`, and computes `DeployConfig.Digest()` from that snapshot. The new `PackageEvent` records that digest, along with `Source`, `Version`, `Flavor`, and `Digest`, when `InProgress` is appended. Only when the entire deployment succeeds does Zarf assign that same snapshot to `DeployedPackage.DeployConfig` while appending `Succeeded`. A failed or cancelled deployment retains the prior successful config. A remove attempt does not modify it; a successful remove deletes the package state as it does today.

`packager.Remove` similarly computes `RemoveConfig.Digest()` before changing anything and records it on the remove event, but does not persist the full `RemoveConfig`. Library users can compare a candidate deploy digest with `DeployedPackage.DeployConfig.Digest()` or the retained event returned by `LastSuccessfulDeploy()`, and compare a candidate remove digest with the relevant retained remove event, without performing either operation.

For deploys, package digest + config digest is the complete resume identity: the package digest identifies the Zarf package and all content/defaults it contains, while the config digest identifies how that package was deployed. The package digest remains the exact OCI manifest digest rather than a Zarf-wrapped value, so it can be used directly as an OCI reference. A change to how Zarf constructs or orders the manifest can legitimately change that digest even when inputs appear semantically equivalent; pinned deterministic tests guard against doing so accidentally.

**Known limitation:** values passed through YAML/JSON can round-trip as `float64`, so numerically-equal-but-differently-formatted values (e.g. `5` vs `5.0`) could theoretically produce different digests even though they represent the same configuration. This proposal accepts that limitation initially - see [Risks and Mitigations](#risks-and-mitigations).

### Behavior Changes

#### Reusing the successful deploy config

`zarf package deploy` gains an opt-in `--reuse-config` flag, with a matching `ReuseConfig bool` SDK option. When selected, Zarf loads `DeployedPackage.DeployConfig` and uses an independent copy as the lowest-precedence explicit config source for the new deployment. Inputs supplied for the new operation are resolved over that baseline through the ordinary precedence and merge path: maps merge recursively, while lists and scalar values replace the stored value. With no new config inputs, the prior successful config is reused exactly.

Only explicit `DeployConfig` is reused. Package and chart defaults are loaded from the incoming package, so this behaves like Helm's `--reset-then-reuse-values`, not exact `--reuse-values`: incoming defaults, then stored explicit config, then new explicit config. The package source and imperative options such as timeout, retries, confirmation, and pruning still come from the new operation. Zarf computes the new event's `ConfigDigest` from the merged config and, if the deployment succeeds, replaces the stored config with it.

`--reuse-config` fails validation before any actions or cluster changes if the package is not already deployed or its state has no stored `DeployConfig`. It does not fall back to empty config, and Zarf does not print the stored config as part of loading or reusing it.

Zarf also needs to reconcile differences between package upgrades to avoid orphaned charts or components, matching Helm / `kubectl` conventions.

Both commands accept an optional comma-separated value for `--prune`:

- Omitting `--prune` disables optional pruning.
- A bare `--prune` or `--prune=all` enables every prune scope applicable to the command.
- `--prune=cluster-resources`, `--prune=images`, and `--prune=cluster-resources,images` select explicit scopes. Explicit lists remain limited to those scopes if more scopes are added later, while `all` includes future scopes applicable to the command.

| Command | `cluster-resources` | `images` |
| --- | --- | --- |
| `zarf package deploy` | Removes charts and components orphaned by the new package. | Removes images orphaned by the new package and no longer referenced by any deployed package. |
| `zarf package remove` | Not applicable; the command already removes the package's cluster resources. | Removes the package's images that are no longer referenced by any remaining package. |

Unknown scopes and scopes not applicable to the selected command fail validation before the operation begins. Before pruning anything, Zarf identifies the selected scopes and prints the charts, components, and images it's about to remove during the same flow as Zarf's existing `--confirm` prompt.

#### Chart, Component, and Image Reconciliation

The selected scopes reconcile at three levels: charts within a component that's still deployed, components dropped entirely from the new package, and images no longer referenced by any deployed package. Each level compares the previously deployed package (fetched via `Cluster.GetDeployedPackage`) against the package about to be deployed (`pkgLayout.Pkg`, already filtered by OS).

##### Chart reconciliation

When `cluster-resources` is selected, the deploy path builds the full list of installed charts for a component present in both the old and new package into a single `[]state.InstalledChart` - both actual Helm charts (`installCharts`) and charts synthesized from `manifests` (`installManifests`), see `deploy.go:531-546`. Each entry is uniquely identified by `namespace/chartName`, the same key `state.MergeInstalledChartsForComponent` already uses to merge chart state across deployments. Zarf diffs the component's previously recorded `InstalledCharts` against this new list by that key: any chart present in the old list but absent from the new one is uninstalled with the same `helm.RemoveChart(ctx, chart.Namespace, chart.ChartName, opts.Timeout)` call `zarf package remove` already uses, and dropped from the stored `InstalledCharts` for that component.

##### Component reconciliation

When `cluster-resources` is selected and an entire component is missing from the new package (its name isn't in `pkgLayout.Pkg.Components`), Zarf removes it the same way `zarf package remove` would: running `Actions.OnRemove` (`Before`, then uninstalling every chart in `InstalledCharts` - Helm-defined and manifest-derived alike, since they already share one list - then `After`/`OnSuccess`/`OnFailure`), and deleting the component's `DeployedComponent` entry. `remove.go`'s per-component removal loop already implements exactly this behavior; this proposal extracts it into a shared helper so `zarf package deploy --prune=cluster-resources` and `zarf package remove` stay behaviorally identical instead of reimplementing component teardown twice.

##### Image reconciliation

When `images` is selected, Zarf diffs each changed or removed component's previously deployed `Images` against the new component's `Images` (both plain `[]string` image references on `v1alpha1.ZarfComponent`). An image present in the old set but not the new one is a pruning candidate only if no other deployed package still needs it: Zarf calls `Cluster.GetDeployedZarfPackages` to list every `DeployedPackage` secret in the cluster and checks whether the candidate image appears in any other package's component images. If nothing else references it, both tags Zarf pushes for that image are removed from the internal registry - the plain tag and the CRC-32-suffixed tag the Zarf agent uses for transparent redirection (see `images/push.go`). If another package still references it, both tags are left alone.

This is the same reference-counted removal `zarf tools registry prune` already does against the whole registry, just scoped down to the images that belonged to the package being deployed or removed instead of scanning every image in the registry.  Because this is package-scoped this operation is also more separable within a registry that is shared with Zarf since users can namespace the repository Zarf uses (`127.0.0.1:31999/zarf`) and place other repositories for other purposes next to Zarf and not have those be affected (like they would with the full catalog explosion the existing prune does).

#### Graceful Cancellation Handling

Today, if a `zarf package deploy` or `zarf package remove` is interrupted, the latest package `PackageEvent`/component `LastEvent` (once this proposal exists) can be left at `InProgress` indefinitely, with nothing to correct it afterward.

On a controlled stop - Zarf catching `SIGINT`/`SIGTERM` via a cancellable `context.Context` (e.g. `signal.NotifyContext`) - Zarf will attempt to append a `Cancelled` package event and set the in-progress component's `LastEvent` to `Cancelled` before exiting. This is best-effort: an ungraceful termination (`SIGKILL`, node crash, power loss) gives Zarf no opportunity to run any code, so the latest event can still be left at `InProgress` - see [Risks and Mitigations](#risks-and-mitigations).

### Test Plan

[x] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

The e2e suite will need to simulate a controlled stop (sending `SIGINT`/`SIGTERM` to a running `zarf package deploy`/`zarf package remove`) so [Graceful Cancellation Handling](#graceful-cancellation-handling) can be exercised end-to-end rather than only unit tested. Test fixtures will also be needed for each `--prune` scope and for `--reuse-config`: package definitions that differ by components, charts, images, variables, and values so reconciliation and config reuse can be tested against a real upgrade rather than a synthetic diff. Fixtures driving multiple sequential deploy/remove cycles against the same package will also be needed to exercise `Events` accumulation and retention eviction once the cap is exceeded.

##### Unit tests

- `DeployedPackage.Events` appends `InProgress` and terminal transitions for deploy/remove attempts; `DeployedComponent.LastEvent` is overwritten as the latest component attempt progresses.
- `LatestEvent()` returns the newest event, while `LastSuccessfulDeploy()` skips newer failed, cancelled, and remove events and returns `false` when no successful deploy remains in retained history.
- `DeployedPackage.Events` retention caps at the configured number of entries, evicting the oldest entry first without clearing the independent successful `DeployConfig` snapshot.
- A successful deploy stores an independent normalized `DeployConfig` snapshot whose digest matches the successful event; failed and cancelled deploys retain the previous snapshot, and remove attempts do not replace it.
- `--reuse-config` copies the stored config before merging explicit new inputs, preserves unspecified prior inputs, applies explicit inputs with normal precedence, and fails before side effects when no stored config exists.
- Default inspect, status, confirmation, and state-related log rendering does not expose the persisted config field, and loading it for reuse does not print it.
- `DeployConfig.Digest()` is deterministic: repeated calls with the same normalized config, map literals built in a different key order, and optional components supplied in a different order produce the same digest; the digest changes when deploy config changes and does not change when only imperative `DeployOptions` fields change.
- Config digests parse and validate as `github.com/opencontainers/go-digest.Digest` values and use SHA-256.
- `DeployConfig.Digest()` returns an error rather than panicking when given a value `encoding/json` can't marshal.
- `RemoveConfig.Digest()` has the same determinism and error-handling properties as `DeployConfig.Digest()`, scoped to its smaller `Values`/`NamespaceOverride` field set; it does not change when only imperative `RemoveOptions` fields (`Timeout`, `SkipVersionCheck`) change.
- Prune scope parsing handles an absent flag, bare `--prune`, each explicit scope, a comma-separated scope list, `all`, unknown scopes, and scopes not applicable to the command.
- `cluster-resources` and `images` independently gate their reconciliation behavior, and image pruning leaves an image alone if it's still referenced by any other deployed package, not just the one being pruned.

##### e2e tests

- `zarf package deploy` with each explicit prune scope, confirming only the selected resource class is removed, and with bare `--prune`, confirming both applicable scopes run.
- `zarf package remove` with `--prune=images` and bare `--prune`, confirming image cleanup for components being uninstalled, and invalid-scope validation before removal begins.
- Two packages sharing an image, with one pruned/removed: confirming the shared image's tags survive while the other package still references it, and are removed once nothing references it anymore.
- Deploying with variables, values, namespace overrides, and per-chart overrides, then deploying with `--reuse-config`, confirming the prior config is reused without being printed; explicitly override one input and confirm the other stored inputs remain unchanged.
- Attempting `--reuse-config` against legacy or otherwise missing config state, confirming the command fails before running actions or changing cluster resources.
- Interrupting a `zarf package deploy`/`zarf package remove` with `SIGINT`, confirming a `Cancelled` package event is appended and the in-progress component's `LastEvent` is set to `Cancelled`.
- Deploying a package twice with identical configuration produces the same `ConfigDigest` on both events; changing `SetVariables` or `Values` between deploys produces a different `ConfigDigest`.
- Removing a package twice (redeploying between removals) with identical `Values` produces the same `ConfigDigest` on both Remove events; changing `Values` between removals produces a different `ConfigDigest`.

### Graduation Criteria

`Events` and `DeployConfig` on `DeployedPackage` are purely additive, and on `DeployedComponent`, `LastEvent` is added alongside the existing `Status` (`ComponentStatus`) field rather than replacing it outright - `ComponentStatus` is deprecated and scheduled for removal after 4 release cycles, but stays populated until then (see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)). A single Zarf version generally manages a given cluster, and integrations built around it are typically tailored to that version, so these schema changes can be absorbed as part of upgrading those integrations rather than needing a coordinated migration. Once Zarf deploys again in that environment, existing package status structs are upgraded (or downgraded) automatically as part of that deploy.

`--prune` is the part of this proposal that can destroy data (see [Risks and Mitigations](#risks-and-mitigations)), so it should stay opt-in and scope-selectable to limit that risk and to match `kubectl` and `helm` conventions that already work this way.

### Upgrade / Downgrade Strategy

The state schema changes are additive, and the new prune and reuse behaviors are opt-in, so upgrades should happen automatically without breaking existing operations. Successful deploys with the newer CLI begin persisting `DeployConfig`. Downgrading remains operationally compatible because older versions ignore or strip the new fields, but `--reuse-config` is unavailable and callers must provide config again; users of `--prune` must likewise return to the older command behavior manually.

This is a breaking change for SDK/library users: moving deploy configuration into `DeployConfig`, and moving `Values`/`NamespaceOverride` into `RemoveConfig` (see [`DeployConfig`, `RemoveConfig`, and `ConfigDigest`](#deployconfig-removeconfig-and-configdigest)), will not compile against existing keyed struct literals. The fix is mechanical: callers nest those fields in `DeployConfig`/`RemoveConfig`; downgrading maps them back to the old top level. This should be documented in release notes.

Rather than dropping `DeployedComponent.Status` (`ComponentStatus`) outright in favor of `LastEvent`, this proposal deprecates it and keeps setting it for 4 release cycles. Zarf doesn't have a general deprecation-timeline policy, but an immediate break is risky here: a tool reading Zarf state and the Zarf CLI writing it aren't guaranteed to be on the same version, so either side could be caught without the new field.

During those 4 releases, `ComponentStatus` is marked deprecated but Zarf keeps setting it alongside `LastEvent`, so existing consumers reading it directly keep working. It has no equivalent for the new `Cancelled` outcome, so a cancelled component just keeps whatever `ComponentStatus` value it already had. After the window, `ComponentStatus` is removed - called out in release notes like any other breaking change - and at that point this becomes the breaking change for anyone still reading it directly.

Existing `DeployedPackage` secrets predating this proposal start with an empty `Events` list, no stored `DeployConfig`, and a zero-value `LastEvent` per component, while their old `status` values stay intact and readable via `ComponentStatus`. `LatestEvent` reports `Unknown` until the package's next deploy or remove. The next successful deploy records `DeployConfig`; until then, `--reuse-config` fails explicitly rather than guessing at prior inputs.

### Version Skew Strategy

This proposal doesn't impact how Zarf's Agent and CLI interact, so no changes are needed there.

It does introduce skew between CLI versions reading the same `DeployedPackage`/`DeployedComponent` secret. A newer CLI reading a secret written by an older, pre-`Events` CLI sees an empty package `Events` list, no `DeployConfig`, and zero-value component `LastEvent`s, and reports `Unknown` via `LatestEvent` until the package's next deploy or remove. An older CLI reading a secret written by a newer CLI ignores the unrecognized `Events`, `DeployConfig`, and `LastEvent` JSON fields, but does still see `ComponentStatus`, which the newer CLI keeps writing during the deprecation window (see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)). If the older CLI subsequently writes the package state, the stored config can be lost and `--reuse-config` remains unavailable until another successful deploy with a newer CLI.

Config digest equality across CLI versions is best-effort. If normalization or serialization changes, a digest produced with the new rules may not match a retained event digest produced with the old rules; Zarf treats that as "changed or unknown" rather than attempting to select digest rules from the top-level `CLIVersion`. The persisted successful `DeployConfig` remains reusable and can be re-hashed under the current rules. No coordinated rollout is required.

## Implementation History

2026-07-01: Initial version of this document.
2026-07-01: Replaced the flat `PackageStatus`/`ComponentStatus` fields with an `Events` list; see [Alternatives](#alternatives).
2026-07-06: `ComponentStatus` is now deprecated and kept for 4 release cycles instead of being dropped immediately.
2026-08-05: Made `--prune` scope-selectable, with bare `--prune` defaulting to all scopes applicable to the command.
2026-08-13: Stored the latest successful `DeployConfig` and added `--reuse-config` behavior.

## Drawbacks

This proposal introduces a breaking SDK change by splitting both `DeployOptions` into `DeployConfig` and `RemoveOptions` into `RemoveConfig`, each separated from their imperative fields, to enable configuration comparison, import hydration, and config reuse. For a proposal primarily about status tracking, requiring every library consumer to update their integration code in two places introduces migration cost, even though the fix itself is mechanical (see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)).

`--prune` is also a larger maintenance commitment than the rest of this proposal. Correctly reconciling orphaned charts, components, and cross-package image references across arbitrary upgrade paths is nontrivial logic to get right and keep right, and a bug here can delete something a user still needed, which most Zarf features don't risk.

`ConfigDigest` approximates whether the configuration changed rather than guaranteeing it. Numeric formatting and future normalization or serialization changes can produce a mismatch where there isn't a meaningful configuration difference (see [Risks and Mitigations](#risks-and-mitigations)). This should be documented clearly so users treat a digest mismatch as "changed or unknown," not an authoritative diff.

Persisting `DeployConfig` improves import, drift reporting, and repeat deployment, but it also makes all normalized deployment inputs durable cluster state even when some values were used only by actions or host-level components. This increases the sensitive data available through the Zarf state Secret and its backups, and consumers such as OpenTofu will generally retain another copy in their own state.

Finally, `DeployedPackage.Events` is a bigger piece of design than a flat status field would have been. Reading package status now requires calling a helper or writing custom logic instead of reading one field directly, and the list needs a retention/cap policy a scalar field never would - the tradeoff for the history and consistency `Events` provides at the package level. The accessor helpers need to stay convenient enough that library users reach for them instead of re-implementing list-walking logic inconsistently. Component-level state (`LastEvent`) avoids that cost by staying a single value rather than a list, but has no history of its own - only the package's `Events` does.

## Alternatives

### Flat `PackageStatus` / `ComponentStatus` fields

An earlier version of this proposal added a single `PackageStatus`/`ComponentStatus` string field to `DeployedPackage`/`DeployedComponent`, extending what already ships for `ComponentStatus` today (`Succeeded`/`Failed`/`Deploying`, unwired for remove). Cancellation added `Cancelled`/`RemoveCancelled` alongside the existing `Failed`/`RemoveFailed` pair:

```go
type PackageStatus string

const (
	PackageStatusUnknown         PackageStatus = "Unknown"
	PackageStatusSucceeded       PackageStatus = "Succeeded"
	PackageStatusFailed          PackageStatus = "Failed"
	PackageStatusDeploying       PackageStatus = "Deploying"
	PackageStatusRemoving        PackageStatus = "Removing"
	PackageStatusRemoveFailed    PackageStatus = "RemoveFailed"
	PackageStatusCancelled       PackageStatus = "Cancelled"
	PackageStatusRemoveCancelled PackageStatus = "RemoveCancelled"
)
```

This was replaced by `Events` for three reasons. First, every new lifecycle operation needed its own paired terminal states - `Cancelled` and `RemoveCancelled` for one new capability, and any future operation (e.g. a verify step) would need the same pairing again; `Type`+`Outcome` composes instead of multiplying. Second, a flat field can't carry a timestamp, so there was no way to tell "just started deploying" from "been stuck deploying for six hours," which conflicts with this proposal's own goal to report deploy/remove progress. Third: a single stored status field has to be actively kept in sync with reality by every code path that touches it, and it's easy to get that wrong - the partial-remove case (a successful removal of some components leaving a sibling component's earlier `Failed` status hidden behind a blanket `Succeeded`) needed dedicated derivation logic to patch over.

### Split `PackagePhase` / `PackageOutcome`

Phase (what Zarf is currently doing) and outcome (the result of the last completed operation) could be tracked as two scalar fields instead of a list, closer to Kubernetes' phase/conditions pattern:

```go
// PackagePhase is the current lifecycle operation.
type PackagePhase string

const (
	PackagePhaseIdle      PackagePhase = "Idle"
	PackagePhaseDeploying PackagePhase = "Deploying"
	PackagePhaseRemoving  PackagePhase = "Removing"
)

// PackageOutcome is the outcome of the most recent completed operation.
type PackageOutcome string

const (
	PackageOutcomeUnknown  PackageOutcome = "Unknown"
	PackageOutcomeHealthy  PackageOutcome = "Healthy"
	PackageOutcomeDegraded PackageOutcome = "Degraded"
)
```

This solves the enum-pairing problem the same way `Events` does - `Phase` and `Outcome` vary independently instead of multiplying - but it was rejected in favor of `Events` because a pair of scalars cannot retain operation history or timestamps. `Events` answers "what was the last Deploy vs. the last Remove, and when" directly.

`Phase`/`Outcome` would offer one capability `Events` doesn't: a persistent "health" value that changes independently of any operation in progress - useful for a system that reconciles continuously, which this proposal explicitly says Zarf isn't becoming (see [Non-Goals](#non-goals)). If that changes, this tradeoff should be revisited.

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
