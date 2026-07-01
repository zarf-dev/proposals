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
  - Resume multi-package deployment (based on uniqueness of package + config)
- Match Kubernetes/Helm conventions where we can

### Non-Goals

- Usurp Helm's lifecycle management of charts/objects deployed by Zarf
- Rollback to a previous Zarf package on failure (would require previous full package source)
- Record full history of previously installed versions of a given Zarf package
- Continuously reconcile state to ensure correctness at all times (e.g. recovery after a hard stop)

## Proposal

To meet these goals, this proposal adds an event history to Zarf's `DeployedPackage` struct and a single last-event snapshot to each `DeployedComponent` struct.  It also changes deploy/remove behavior so both stay accurate over time, and reorganizes `packager.DeployOptions`/`packager.RemoveOptions` so a subset of each can be hashed for comparison. See [Design Details](#design-details) for the specifics of each change.

- `DeployedPackage` gains an `Events` list recording every deploy/remove attempt against the package - its type (`Deploy`/`Remove`), outcome (`InProgress`/`Succeeded`/`Failed`/`Cancelled`), timestamp, package version, flavor and digest, and config digest. Deploy events additionally carry the source used.
- `DeployedComponent` gains a `LastEvent` field recording only the most recent deploy/remove attempt against that specific component, since components mostly mirror the package's own timeline and rarely need a history of their own.
- Package status is read via `LatestEvent()`, component status by reading `LastEvent` directly - neither is a separately stored field, so neither can drift from what actually happened. See [Reading status from `Events`](#reading-status-from-events).
- On a graceful stop, Zarf appends a `Cancelled` event for whatever was in progress at the package level, and sets `Cancelled` on the in-progress component's `LastEvent`, rather than leaving either stuck at `InProgress`. See [Graceful Cancellation Handling](#graceful-cancellation-handling).
- `zarf package deploy` and `zarf package remove` gain an opt-in `--prune` flag, matching Helm / `kubectl` conventions, that removes any charts, components, or no-longer-referenced images orphaned by the new deployment.
- `DeployOptions`' config-affecting fields (`SetVariables`, `Values`, `NamespaceOverride`, `ValuesOverridesMap`) move into a new `DeployConfig` type, and `RemoveOptions`' equivalent fields (`Values`, `NamespaceOverride`) move into a new `RemoveConfig` type. Every deploy or remove `PackageEvent` records a `ConfigDigest` computed from the relevant one, so library users can tell whether a candidate deploy or remove's configuration differs from previous events.

### User Stories (Optional)

#### Story 1

**As** a package deployer, when I deploy a new version of a package that has removed a component or chart from a previous version, **I want** Zarf to detect that the component is no longer present and clean up the orphaned resources during `zarf package deploy --prune`, **so that** they are not left behind, untracked, in the cluster.

#### Story 2

**As** a package deployer, **I want** to know where a currently deployed package came from (e.g. the OCI reference or URL it was sourced from) **so that** I can redeploy it later, such as to recover from an incident, without having to keep a separate record of the source myself.

#### Story 3

**As** a package deployer or operator, **I want** a single, reliable way to check the overall status of a package - including while it is being removed or after a removal has failed - **so that** I don't have to infer status by piecing together logs or inspecting individual cluster resources.

#### Story 4

**As** a package deployer, before I redeploy or upgrade a package, **I want** to compare the currently deployed configuration with my new configuration, **so that** I can tell whether my new deployment would actually change anything.

### Risks and Mitigations

- **`--prune` incorrectly removes resources still in use.** If the orphan-detection logic incorrectly identifies a component as removed, or the cross-package image reference counting has a bug, `--prune` could delete a chart or image that's still needed - including one shared with an unrelated package. This is the only irreversible/destructive behavior this proposal adds.
  - *Mitigation:* `--prune` is opt-in and off by default on both `zarf package deploy` and `zarf package remove`, so Zarf's existing behavior is unchanged unless a user explicitly asks for it - anyone who hits a correctness issue can stop passing the flag.
- **The latest `PackageEvent`/`LastEvent` can be left at `InProgress`.** If a deploy or remove is interrupted, the most recent package `PackageEvent` or component `LastEvent` can be left indefinitely showing an operation that's no longer actually running.
  - *Mitigation:* see [Graceful Cancellation Handling](#graceful-cancellation-handling) - on a controlled stop (e.g. `SIGINT`/`SIGTERM`), Zarf appends a `Cancelled` package event and sets `Cancelled` `LastEvent`s for any components in progress before exiting. This does not cover an ungraceful termination (`SIGKILL`, node crash, power loss); no code runs in those cases, so the latest event can still be left at `InProgress`. That residual risk is accepted - Zarf is not becoming a continuously reconciling controller (see [Non-Goals](#non-goals)).
- **Unbounded `Events` growth.** `DeployedPackage` is stored in a Kubernetes Secret; an ever-growing `Events` list risks approaching the Secret size limit on a long-lived, frequently-redeployed package. `DeployedComponent.LastEvent` is a single value, not a list, so it doesn't carry this risk.
  - *Mitigation:* `Events` is capped at a fixed number of most-recent entries (see [`PackageEvent` retention](#packageevent-retention)), evicting the oldest entry once the cap is exceeded.
- **`DeployConfig.Digest()`/`RemoveConfig.Digest()` computation can fail on non-serializable config.** `Values` and `ValuesOverridesMap` are typed as `map[string]any`, so a library user could put something in them that `encoding/json` can't marshal (e.g. a function or channel value set programmatically rather than parsed from YAML).
  - *Mitigation:* document that `DeployConfig`/`RemoveConfig` fields must be JSON-marshalable, and have `packager.Deploy`/`packager.Remove` compute (and fail fast on) the digest at the start of the operation rather than only when a library user calls it later for comparison - so an unmarshalable config is caught immediately instead of surfacing as a confusing error somewhere else.
- **Config digest false negatives from numeric formatting.** `ConfigDigest` is computed by JSON-marshaling user-supplied values. Semantically-identical values that are formatted differently (e.g. `5` vs `5.0` in a values file) could produce different digests, causing Zarf to report a configuration change when there isn't a meaningful one.
  - *Mitigation:* document this as a known limitation for v1; if it proves disruptive in practice, add a canonicalization pass (e.g. normalize numeric types before hashing) in a follow-up.
- **A `Source` could persist embedded credentials.** A source string like `https://user:pass@host/...` recorded on a deploy `PackageEvent` would be stored as-is in the `DeployedPackage` secret and could resurface in `zarf package inspect` output or logs. In practice this should be rare - most sources are `oci://` or local tarballs, and Zarf already supports `.netrc` for credential storage, so embedding credentials in the source URI is uncommon.
  - *Mitigation:* document that credentials should be provided via `.netrc` rather than embedded in the source string; no code-level scrubbing planned given how rare this path is.

## Design Details

### `DeployedPackage` and `DeployedComponent` changes

#### Add `Events`

`DeployedPackage` gains an `Events` list and `DeployedComponent` gains a single `LastEvent`, replacing the existing `ComponentStatus` field:

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

// PackageEvent records a single deploy or remove attempt against a package.
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
	// ConfigDigest is a deterministic, self-describing digest of the DeployConfig (Deploy
	// events) or RemoveConfig (Remove events) used for this operation (e.g. "sha256:<hex>").
	ConfigDigest string `json:"configDigest,omitempty"`
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
```

`DeployedComponent.LastEvent` replaces `ComponentStatus` directly - `Type`+`Outcome` say whether that component is deploying, removing, or at a terminal state.

#### `PackageEvent` retention

Since `DeployedPackage` is stored inside a Kubernetes Secret, `Events` can't grow without bound. The list is capped at a fixed number of most-recent entries (e.g. `10`); once the cap is exceeded, the oldest entry is dropped when a new one is appended. This keeps enough history to answer "what was the last event, and the one before it" without risking the Secret's size limit on a package that's redeployed or removed and redeployed many times over its lifetime. `DeployedComponent.LastEvent` is a single value, not a list, so there's nothing to cap there - each deploy or remove attempt overwrites it.

#### `DeployConfig`, `RemoveConfig`, and `ConfigDigest`

Comparing a candidate deployment's configuration against what's already deployed - without redeploying - requires distinguishing fields that affect the resulting deployment from fields that only affect how the deploy operation runs. Today `packager.DeployOptions` mixes both. This proposal splits it as follows:

- **Config** (affects the resulting deployment, and should be part of the digest): `SetVariables`, `Values` (`value.Values`), `NamespaceOverride`, `ValuesOverridesMap`
- **Imperative** (affects only the deploy operation, and stays out of the digest): everything else, e.g. `Timeout`, `Retries`, `OCIConcurrency`, `IsInteractive`, `ForceConflicts`, `AdoptExistingResources`, `SkipVersionCheck`, `RemoteOptions`, and the Zarf init state used to configure a cluster (`GitServer`, `RegistryInfo`, `ArtifactServer`, `AgentTLS`, `AgentMutationPolicy`, `StorageClass`, `InjectorPort`)

The same split applies to `packager.RemoveOptions`, which today also mixes `Values` and `NamespaceOverride` in with imperative fields (`Cluster`, `Timeout`, `SkipVersionCheck`) - Zarf Values are usable in `onRemove` actions, so they affect what a remove actually does, not just how it runs. `RemoveOptions`' config subset is smaller than `DeployOptions`': it has no `SetVariables` or `ValuesOverridesMap` to begin with, since those only apply to Helm chart installs.

The config fields move into new `DeployConfig` and `RemoveConfig` types, embedded in `DeployOptions` and `RemoveOptions` respectively:

```go
// DeployConfig holds the subset of deploy options that affect the resulting deployment,
// as opposed to options that only affect how the deploy operation itself runs.
type DeployConfig struct {
	SetVariables       map[string]string
	Values             value.Values
	NamespaceOverride  string
	ValuesOverridesMap ValuesOverrides
}

// Digest returns a deterministic, self-describing digest of the config (e.g. "sha256:<hex>"),
// suitable for comparing against a previously deployed package's ConfigDigest. The algorithm
// prefix allows a future change to the hashing/canonicalization approach to be identified
// rather than silently comparing incompatible digests.
func (c DeployConfig) Digest() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// RemoveConfig holds the subset of remove options that affect the resulting removal (via
// onRemove actions), as opposed to options that only affect how the remove operation runs.
type RemoveConfig struct {
	Values            value.Values
	NamespaceOverride string
}

// Digest returns a deterministic, self-describing digest of the config, using the same
// approach as DeployConfig.Digest().
func (c RemoveConfig) Digest() (string, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
```

`encoding/json` sorts map keys alphabetically at every nesting level, so `json.Marshal` over either type is deterministic for the map-shaped fields they contain (`SetVariables`, `Values`, `ValuesOverridesMap`) without any extra canonicalization work.

`packager.Deploy` computes `DeployConfig.Digest()` at the start of every deployment and `packager.Remove` computes `RemoveConfig.Digest()` at the start of every removal; each sets it as `ConfigDigest` on the `PackageEvent` it appends - `Deploy` events also get `Source`, and both `Deploy` and `Remove` events get `Version` / `Flavor` / `Digest` - all at the same point that event's `Outcome` starts as `InProgress`. If the digest can't be computed (see [Risks and Mitigations](#risks-and-mitigations)), the operation fails fast before making any changes.

Library users get the `DeployConfig.Digest()`/`RemoveConfig.Digest()` helpers above so they can compute a digest for a candidate operation and compare it against a previous event without actually deploying or removing.  Packages deployed before `Events` existed will have no deploy events to find, so comparisons in that case should be treated as unknown/always-different, since Zarf has no record of what those packages were deployed with.

**Known limitation:** values passed through YAML/JSON can round-trip as `float64`, so numerically-equal-but-differently-formatted values (e.g. `5` vs `5.0`) could theoretically produce different digests even though they represent the same configuration. This proposal accepts that limitation for v1 - see [Risks and Mitigations](#risks-and-mitigations).

### Behavior Changes

Zarf also needs to reconcile differences between package upgrades to avoid orphaned charts or components, matching Helm / `kubectl` conventions.

- `zarf package deploy` gains a `--prune` flag, off by default. At the start of a deployment, Zarf compares the existing deployed state against the to-be-deployed state. Any charts or components that would be orphaned by the new package are removed, along with any images referenced only by those components. (Zarf itself does not perform registry garbage collection.) It's opt-in, so anyone who hits a correctness issue with it can omit the flag and get today's behavior.
- `zarf package remove` gains the same `--prune` flag, applying the equivalent image-removal behavior for components that are being uninstalled.

#### Chart, Component, and Image Reconciliation

`--prune` reconciles at three levels: charts within a component that's still deployed, components dropped entirely from the new package, and images no longer referenced by any deployed package. Each level compares the previously deployed package (fetched via `Cluster.GetDeployedPackage`) against the package about to be deployed (`pkgLayout.Pkg`, already filtered by OS).

##### Chart reconciliation

For a component present in both the old and new package, the deploy path already builds the full list of installed charts for that component into a single `[]state.InstalledChart` - both actual Helm charts (`installCharts`) and charts synthesized from `manifests` (`installManifests`), see `deploy.go:531-546`. Each entry is uniquely identified by `namespace/chartName`, the same key `state.MergeInstalledChartsForComponent` already uses to merge chart state across deployments. `--prune` diffs the component's previously recorded `InstalledCharts` against this new list by that key: any chart present in the old list but absent from the new one is uninstalled with the same `helm.RemoveChart(ctx, chart.Namespace, chart.ChartName, opts.Timeout)` call `zarf package remove` already uses, and dropped from the stored `InstalledCharts` for that component.

##### Component reconciliation

If an entire component is missing from the new package (its name isn't in `pkgLayout.Pkg.Components`), `--prune` removes it the same way `zarf package remove` would: running `Actions.OnRemove` (`Before`, then uninstalling every chart in `InstalledCharts` - Helm-defined and manifest-derived alike, since they already share one list - then `After`/`OnSuccess`/`OnFailure`), and deleting the component's `DeployedComponent` entry. `remove.go`'s per-component removal loop already implements exactly this behavior; this proposal extracts it into a shared helper so `zarf package deploy --prune` and `zarf package remove` stay behaviorally identical instead of reimplementing component teardown twice.

##### Image reconciliation

For each component that changed or was removed, `--prune` diffs the previously deployed component's `Images` against the new component's `Images` (both plain `[]string` image references on `v1alpha1.ZarfComponent`). An image present in the old set but not the new one is a pruning candidate only if no other deployed package still needs it: `--prune` calls `Cluster.GetDeployedZarfPackages` to list every `DeployedPackage` secret in the cluster and checks whether the candidate image appears in any other package's component images. If nothing else references it, both tags Zarf pushes for that image are removed from the internal registry - the plain tag and the CRC-32-suffixed tag the Zarf agent uses for transparent redirection (see `images/push.go`). If another package still references it, both tags are left alone.

#### Graceful Cancellation Handling

Today, if a `zarf package deploy` or `zarf package remove` is interrupted, the latest package `PackageEvent`/component `LastEvent` (once this proposal exists) can be left at `InProgress` indefinitely, with nothing to correct it afterward.

On a controlled stop - Zarf catching `SIGINT`/`SIGTERM` via a cancellable `context.Context` (e.g. `signal.NotifyContext`) - Zarf will attempt to append a `Cancelled` event to the package's `Events`, and set `LastEvent` to `Cancelled` on whatever component was in progress, before exiting; both use the `Type` of the operation that was interrupted. This lets library users tell which operation was in flight when Zarf stopped, without having to separately track what command was running. This is best-effort: an ungraceful termination (`SIGKILL`, node crash, power loss) gives Zarf no opportunity to run any code, so the latest event can still be left at `InProgress` in that case - see [Risks and Mitigations](#risks-and-mitigations).

### Test Plan

[x] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

The e2e suite will need to simulate a controlled stop (sending `SIGINT`/`SIGTERM` to a running `zarf package deploy`/`zarf package remove`) so [Graceful Cancellation Handling](#graceful-cancellation-handling) can be exercised end-to-end rather than only unit tested. Test fixtures will also be needed for `--prune`: package definitions that differ only by an added/removed component or chart, so orphan-detection can be tested against a real upgrade rather than a synthetic diff. Fixtures driving multiple sequential deploy/remove cycles against the same package will also be needed to exercise `Events` accumulation and retention eviction once the cap is exceeded.

##### Unit tests

- `DeployedPackage.Events` are appended (never overwritten) for every deploy/remove outcome, including `Cancelled`; `DeployedComponent.LastEvent` is overwritten on every deploy/remove attempt against that component.
- `DeployedPackage.Events` retention caps at the configured number of entries, evicting the oldest entry first.
- `DeployConfig.Digest()` is deterministic: repeated calls with the same config, and calls with map literals built in a different key order, all produce the same digest; the digest changes when `SetVariables`, `Values`, `NamespaceOverride`, or `ValuesOverridesMap` change, and does not change when only imperative `DeployOptions` fields change.
- `DeployConfig.Digest()` returns an error rather than panicking when given a value `encoding/json` can't marshal.
- `RemoveConfig.Digest()` has the same determinism and error-handling properties as `DeployConfig.Digest()`, scoped to its smaller `Values`/`NamespaceOverride` field set; it does not change when only imperative `RemoveOptions` fields (`Timeout`, `SkipVersionCheck`) change.
- The `--prune` orphan-detection logic correctly identifies charts/components that were removed between two `ZarfPackage` definitions (see [Chart, Component, and Image Reconciliation](#chart-component-and-image-reconciliation)), and correctly leaves an image alone if it's still referenced by any other deployed package, not just the one being pruned.

##### e2e tests

- `zarf package deploy --prune` against a package that removed a component/chart from the previously deployed version, confirming the orphaned chart, component, and any now-unreferenced images are removed.
- `zarf package remove --prune`, confirming the equivalent image cleanup for components being uninstalled.
- Two packages sharing an image, with one pruned/removed: confirming the shared image's tags survive while the other package still references it, and are removed once nothing references it anymore.
- Interrupting a `zarf package deploy`/`zarf package remove` with `SIGINT`, confirming a `Cancelled` event of the correct `Type` is appended to the package's `Events` and set on the in-progress component's `LastEvent`.
- Deploying a package twice with identical configuration produces the same `ConfigDigest` on both events; changing `SetVariables` or `Values` between deploys produces a different `ConfigDigest`.
- Removing a package twice (redeploying between removals) with identical `Values` produces the same `ConfigDigest` on both Remove events; changing `Values` between removals produces a different `ConfigDigest`.

### Graduation Criteria

`Events` on `DeployedPackage` is purely additive, and on `DeployedComponent`, `LastEvent` replaces the existing `Status` (`ComponentStatus`) field. A single Zarf version generally manages a given cluster, and integrations built around it are typically tailored to that version, so these schema changes can be absorbed as part of upgrading those integrations rather than needing a coordinated migration. Once Zarf deploys again in that environment, existing package status structs are upgraded (or downgraded) automatically as part of that deploy.

`--prune` is the part of this proposal that can destroy data (see [Risks and Mitigations](#risks-and-mitigations)), so it should stay opt-in to limit that risk and to match `kubectl` and `helm` conventions that already work this way.

### Upgrade / Downgrade Strategy

The change in Zarf behavior is optional and the state changes additive so upgrades should be able to happen automatically without breakages.  Downgrades in state would also be non-breaking since the new fields would just be stripped.  If users had opted to use the new `--prune` flag then they would need to manually migrate back.  This should be acceptable though given the adoption of `--prune` would be intentional in the first place.

This is a breaking change for SDK/library users: moving `SetVariables`, `Values`, `NamespaceOverride`, and `ValuesOverridesMap` out of `DeployOptions` and into the nested `DeployConfig`, and moving `Values`/`NamespaceOverride` out of `RemoveOptions` and into the nested `RemoveConfig` (see [`DeployConfig`, `RemoveConfig`, and `ConfigDigest`](#deployconfig-removeconfig-and-configdigest)), will not compile against existing caller code. The fix is mechanical, though - callers move those fields from a top-level `DeployOptions`/`RemoveOptions` struct literal into a `DeployConfig`/`RemoveConfig` struct literal, either inline or assigned to `DeployOptions.DeployConfig`/`RemoveOptions.RemoveConfig`. Downgrading is the same mapping in reverse. This kind of internal restructuring is similar to other breaking changes Zarf has shipped before, and should be documented in release notes the same way those were.

This also removes `DeployedComponent.Status` (`ComponentStatus`) in favor of `LastEvent`. That field already ships today, set during deploy (`ComponentStatusDeploying`/`Succeeded`/`Failed`), so this is a breaking change for any external SDK consumer reading it directly - even though nothing in the Zarf codebase itself reads it back beyond the code that sets it. Existing `DeployedPackage` secrets already in a cluster, deployed by a Zarf version that predates this proposal, will have their old `status` JSON keys ignored on unmarshal (Go's `encoding/json` drops unrecognized fields). The package starts with an empty `Events` list, and each component starts with a zero-value `LastEvent`. `LatestEvent` reports `Unknown` in that state until the package's next deploy or remove appends/sets an event.

### Version Skew Strategy

This proposal doesn't impact how Zarf's Agent and CLI interact, so no changes are needed there.

It does introduce skew between CLI versions reading the same `DeployedPackage`/`DeployedComponent` secret. A newer CLI reading a secret written by an older, pre-`Events` CLI sees an empty package `Events` list and zero-value component `LastEvent`s, and reports `Unknown` via `LatestEvent` until the package's next deploy or remove. An older CLI reading a secret written by a newer CLI doesn't see the `Events`/`LastEvent` fields at all (unrecognized JSON keys are ignored on unmarshal). No coordinated rollout is required.

## Implementation History

2026-07-01: Initial version of this document.
2026-07-01: Replaced the flat `PackageStatus`/`ComponentStatus` fields with an `Events` list; see [Alternatives](#alternatives).

## Drawbacks

This proposal introduces a breaking SDK change (splitting both `DeployOptions` into `DeployConfig` and `RemoveOptions` into `RemoveConfig`, each separated from their imperative fields) purely to enable configuration comparison. For a proposal primarily about status tracking, requiring every library consumer to update their integration code in two places introduces migration cost, even though the fix itself is mechanical (see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)).

`--prune` is also a larger maintenance commitment than the rest of this proposal. Correctly reconciling orphaned charts, components, and cross-package image references across arbitrary upgrade paths is nontrivial logic to get right and keep right, and a bug here can delete something a user still needed, which most Zarf features don't risk.

`ConfigDigest` approximates whether the configuration changed rather than guaranteeing it - the known numeric-formatting limitation (see [Risks and Mitigations](#risks-and-mitigations)) means it can report a change where there isn't a meaningful one. This should be documented clearly so users don't treat a digest mismatch as an authoritative diff.

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

This solves the enum-pairing problem the same way `Events` does - `Phase` and `Outcome` vary independently instead of multiplying - but it was rejected in favor of `Events` for two reasons. First, two independently-set fields can drift from each other the same way a single status field could go stale; `Events` is a single append-only list, so there's nothing for two separate writers to disagree about. Second, a pair of scalars still can't carry history or timestamps - `Events` answers "what was the last Deploy vs. the last Remove, and when" directly, and `Phase`/`Outcome` can't represent that without becoming a list of pairs.

`Phase`/`Outcome` would offer one capability `Events` doesn't: a persistent "health" value that changes independently of any operation in progress - useful for a system that reconciles continuously, which this proposal explicitly says Zarf isn't becoming (see [Non-Goals](#non-goals)). If that changes, this tradeoff should be revisited.

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
