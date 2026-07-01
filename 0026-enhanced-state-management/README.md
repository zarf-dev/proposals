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
  - [`DeployedPackage` changes](#deployedpackage-changes)
    - [Add overall `PackageStatus`](#add-overall-packagestatus)
    - [Add package `Source`](#add-package-source)
    - [Add package `ConfigDigest`](#add-package-configdigest)
    - [Add removal statuses to `ComponentStatus`](#add-removal-statuses-to-componentstatus)
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
  - [Split `PackagePhase` / `PackageOutcome`](#split-packagephase--packageoutcome)
- [Infrastructure Needed (Optional)](#infrastructure-needed-optional)
<!-- /toc -->

## Summary

This ZEP proposes to improve Zarf's tracking of package and component state - overall status, source, and configuration - so that Zarf and its library users can make smarter decisions about a deployed package: what changed, where it came from, and what to reconcile when redeploying or removing it.

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

To meet these goals, this proposal adds new state to Zarf's `DeployedPackage` secret, changes deploy/remove behavior so that state stays accurate over time, and reorganizes `packager.DeployOptions` so a subset of it can be hashed for comparison. See [Design Details](#design-details) for the specifics of each change.

- `DeployedPackage` gains a top level `PackageStatus` field describing the overall state of a package (e.g. `Deploying`, `Succeeded`, `Failed`, `Removing`, `RemoveFailed`, `Cancelled`, `RemoveCancelled`).
- `ComponentStatus` gains the same removal states, and its existing `Removing` state is wired into the `zarf package remove` flow.
- On a graceful stop, Zarf attempts to persist `Cancelled` or `RemoveCancelled` for whatever package/component was in progress, rather than leaving a stale `Deploying`/`Removing` behind. See [Graceful Cancellation Handling](#graceful-cancellation-handling).
- `DeployedPackage` gains a top level `Source` field recording where the package was deployed from (e.g. `oci://ghcr.io/packages/name:1.0.0`).
- `zarf package deploy` and `zarf package remove` gain an opt-in `--prune` flag, matching Helm / `kubectl` conventions, that removes any charts, components, or no-longer-referenced images orphaned by the new deployment.
- `DeployOptions`' config-affecting fields (`SetVariables`, `Values`, `NamespaceOverride`, `ValuesOverridesMap`) move into a new `DeployConfig` type, and `DeployedPackage` gains a `ConfigDigest` computed from it - so library users can tell whether a candidate deployment's configuration differs from what's currently deployed, without redeploying. See [Add package `ConfigDigest`](#add-package-configdigest).

### User Stories (Optional)

#### Story 1

**As** a package deployer, when I deploy a new version of a package that has removed a component or chart from a previous version, **I want** Zarf to detect that the component is no longer present and clean up the orphaned resources during `zarf package deploy --prune`, **so that** they are not left behind, untracked, in the cluster.

#### Story 2

**As** a package deployer, **I want** to know where a currently deployed package came from (e.g. the OCI reference or URL it was sourced from) **so that** I can redeploy it later, such as to recover from an incident, without having to keep a separate record of the source myself.

#### Story 3

**As** a package deployer or operator, **I want** a single, reliable field to check the overall status of a package - including while it is being removed or after a removal has failed - **so that** I don't have to infer status by piecing together logs or inspecting individual cluster resources.

#### Story 4

**As** a package deployer, before I redeploy or upgrade a package, **I want** to compare the currently deployed configuration with my new configuration, **so that** I can tell whether my new deployment would actually change anything.

### Risks and Mitigations

- **`--prune` incorrectly removes resources still in use.** If the orphan-detection logic incorrectly identifies a component as removed, or the cross-package image reference counting has a bug, `--prune` could delete a chart or image that's still needed - including one shared with an unrelated package. This is the only irreversible/destructive behavior this proposal adds.
  - *Mitigation:* `--prune` is opt-in and off by default on both `zarf package deploy` and `zarf package remove`, so Zarf's existing behavior is unchanged unless a user explicitly asks for it - anyone who hits a correctness issue can simply stop passing the flag.
- **`PackageStatus`/`ComponentStatus` can go stale or stuck.** If a deploy or remove is interrupted, a status like `Deploying` or `Removing` can be left behind indefinitely with nothing to reconcile it afterward.
  - *Mitigation:* see [Graceful Cancellation Handling](#graceful-cancellation-handling) - on a controlled stop (e.g. `SIGINT`/`SIGTERM`), Zarf will attempt to persist a new `Cancelled` (deploy) or `RemoveCancelled` (remove) status before exiting. This does not cover an ungraceful termination (`SIGKILL`, node crash, power loss); no code runs in those cases, so the status can still go stale. That residual risk is accepted - Zarf is not becoming a continuously reconciling controller (see [Non-Goals](#non-goals)).
- **`ConfigDigest` computation can fail on non-serializable config.** `Values` and `ValuesOverridesMap` are typed as `map[string]any`, so a library user could put something in them that `encoding/json` can't marshal (e.g. a function or channel value set programmatically rather than parsed from YAML).
  - *Mitigation:* document that `DeployConfig` fields must be JSON-marshalable, and have `packager.Deploy` compute (and fail fast on) the digest at the start of a deployment rather than only when a library user calls it later for comparison - so an unmarshalable config is caught immediately instead of surfacing as a confusing error somewhere else.
- **Config digest false negatives from numeric formatting.** `ConfigDigest` is computed by JSON-marshaling user-supplied values. Semantically-identical values that are formatted differently (e.g. `5` vs `5.0` in a values file) could produce different digests, causing Zarf to report a configuration change when there isn't a meaningful one.
  - *Mitigation:* document this as a known limitation for v1; if it proves disruptive in practice, add a canonicalization pass (e.g. normalize numeric types before hashing) in a follow-up.
- **`Source` could persist embedded credentials.** A source string like `https://user:pass@host/...` would be stored as-is in the `DeployedPackage` secret and could resurface in `zarf package inspect` output or logs. In practice this should be rare - most sources are `oci://` or local tarballs, and Zarf already supports `.netrc` for credential storage, so embedding credentials in the source URI is uncommon.
  - *Mitigation:* document that credentials should be provided via `.netrc` rather than embedded in the source string; no code-level scrubbing planned given how rare this path is.

## Design Details

### `DeployedPackage` changes

To better surface the deployment state of packages and their components, `DeployedPackage` gains the following changes:

#### Add overall `PackageStatus`

`PackageStatus` is a top level field on the `DeployedPackage` secret with the following states.

```go
// PackageStatus defines the overall deployment status of a Zarf package.
type PackageStatus string

// All the different status options for a Zarf Package
const (
	PackageStatusUnknown         PackageStatus = "Unknown"
	PackageStatusSucceeded       PackageStatus = "Succeeded"
	PackageStatusFailed          PackageStatus = "Failed"
	PackageStatusDeploying       PackageStatus = "Deploying"
	PackageStatusRemoving        PackageStatus = "Removing"
	PackageStatusRemoveFailed    PackageStatus = "RemoveFailed"
	PackageStatusCancelled       PackageStatus = "Cancelled"       // graceful stop during deploy, see Graceful Cancellation Handling
	PackageStatusRemoveCancelled PackageStatus = "RemoveCancelled" // graceful stop during remove, see Graceful Cancellation Handling
)
```

#### Add package `Source`

`DeployedPackage` gains a top-level `Source` field that records the source (e.g. oci:// URL, https:// URL, or tarball filepath) that this Zarf package was deployed from.

```go
// Source is the original source string used to deploy the package (e.g. oci:// URL, path to tarball).
Source string `json:"source,omitempty"`
```

#### Add package `ConfigDigest`

Comparing a candidate deployment's configuration against what's already deployed - without redeploying - requires distinguishing fields that affect the resulting deployment from fields that only affect how the deploy operation runs. Today `packager.DeployOptions` mixes both. This proposal splits it as follows:

- **Config** (affects the resulting deployment, and should be part of the digest): `SetVariables`, `Values` (`value.Values`), `NamespaceOverride`, `ValuesOverridesMap`
- **Imperative** (affects only the deploy operation, and stays out of the digest): everything else, e.g. `Timeout`, `Retries`, `OCIConcurrency`, `IsInteractive`, `ForceConflicts`, `AdoptExistingResources`, `SkipVersionCheck`, `RemoteOptions`, and the Zarf init state used to configure a cluster (`GitServer`, `RegistryInfo`, `ArtifactServer`, `AgentTLS`, `AgentMutationPolicy`, `StorageClass`, `InjectorPort`)

The config fields move into a new `DeployConfig` type embedded in `DeployOptions`:

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
```

`encoding/json` sorts map keys alphabetically at every nesting level, so `json.Marshal` over `DeployConfig` is deterministic for the map-shaped fields it contains (`SetVariables`, `Values`, `ValuesOverridesMap`) without any extra canonicalization work.

`DeployedPackage` gains a `ConfigDigest` field, recorded the same way `Source` and `Status` are today:

```go
// ConfigDigest is a deterministic, self-describing digest of the DeployConfig used for this
// deployment (e.g. "sha256:<hex>").
ConfigDigest string `json:"configDigest,omitempty"`
```

`packager.Deploy` computes `DeployConfig.Digest()` at the start of every deployment and records the result as `DeployedPackage.ConfigDigest` - the same point at which it sets `PackageStatus` to `Deploying` - so the stored digest always reflects the config that deployment actually used. If the digest can't be computed (see [Risks and Mitigations](#risks-and-mitigations)), the deployment fails fast before making any changes.

Library users get the `DeployConfig.Digest()` helper above so they can compute a digest for a candidate deployment and compare it against `DeployedPackage.ConfigDigest` without actually deploying. Packages deployed before this field existed will have an empty `ConfigDigest`, so a comparison against them always reports a mismatch, since Zarf has no record of what those packages were deployed with.

**Known limitation:** values passed through YAML/JSON can round-trip as `float64`, so numerically-equal-but-differently-formatted values (e.g. `5` vs `5.0`) could theoretically produce different digests even though they represent the same configuration. This proposal accepts that limitation for v1 - see [Risks and Mitigations](#risks-and-mitigations).

#### Add removal statuses to `ComponentStatus`

`ComponentStatus` is currently missing the ability to track remove failures and the current `Removing` state is not wired into the `zarf package remove` flow.

```go
// ComponentStatus defines the deployment status of a Zarf component within a package.
type ComponentStatus string

// All the different status options for a Zarf Component
const (
	ComponentStatusSucceeded       ComponentStatus = "Succeeded"
	ComponentStatusFailed          ComponentStatus = "Failed"
	ComponentStatusDeploying       ComponentStatus = "Deploying"
	ComponentStatusRemoving        ComponentStatus = "Removing"        // newly wired into remove flow
	ComponentStatusRemoveFailed    ComponentStatus = "RemoveFailed"    // newly added
	ComponentStatusCancelled       ComponentStatus = "Cancelled"       // graceful stop during deploy, see Graceful Cancellation Handling
	ComponentStatusRemoveCancelled ComponentStatus = "RemoveCancelled" // graceful stop during remove, see Graceful Cancellation Handling
)
```

### Behavior Changes

Zarf also needs to reconcile differences between package upgrades to avoid orphaned charts or components, matching Helm / `kubectl` conventions.

- `zarf package deploy` gains a `--prune` flag, off by default. At the start of a deployment, Zarf compares the existing deployed state against the to-be-deployed state. Any charts or components that would be orphaned by the new package are removed, along with any images referenced only by those components. (Zarf itself does not perform registry garbage collection.) Being opt-in, `--prune` can simply be omitted by anyone who hits a correctness issue with it, falling back to today's behavior.
- `zarf package remove` gains the same `--prune` flag, applying the equivalent image-removal behavior for components that are being uninstalled.

#### Chart, Component, and Image Reconciliation

`--prune` reconciles at three levels: charts within a component that's still deployed, components dropped entirely from the new package, and images no longer referenced by any deployed package. Each level compares the previously deployed package (fetched via `Cluster.GetDeployedPackage`) against the package about to be deployed (`pkgLayout.Pkg`, already filtered by OS).

##### Chart reconciliation

For a component present in both the old and new package, the deploy path already builds the full list of installed charts for that component into a single `[]state.InstalledChart` - both actual Helm charts (`installCharts`) and charts synthesized from `manifests` (`installManifests`), see `deploy.go:535-550`. Each entry is uniquely identified by `namespace/chartName`, the same key `state.MergeInstalledChartsForComponent` already uses to merge chart state across deployments. `--prune` diffs the component's previously recorded `InstalledCharts` against this new list by that key: any chart present in the old list but absent from the new one is uninstalled with the same `helm.RemoveChart(ctx, chart.Namespace, chart.ChartName, opts.Timeout)` call `zarf package remove` already uses, and dropped from the stored `InstalledCharts` for that component.

##### Component reconciliation

If an entire component is missing from the new package (its name isn't in `pkgLayout.Pkg.Components`), `--prune` removes it the same way `zarf package remove` would: running `Actions.OnRemove` (`Before`, then uninstalling every chart in `InstalledCharts` - Helm-defined and manifest-derived alike, since they already share one list - then `After`/`OnSuccess`/`OnFailure`), and deleting the component's `DeployedComponent` entry. `remove.go`'s per-component removal loop already implements exactly this behavior; this proposal extracts it into a shared helper so `zarf package deploy --prune` and `zarf package remove` stay behaviorally identical instead of reimplementing component teardown twice.

##### Image reconciliation

For each component that changed or was removed, `--prune` diffs the previously deployed component's `Images` against the new component's `Images` (both plain `[]string` image references on `v1alpha1.ZarfComponent`). An image present in the old set but not the new one is a pruning candidate only if no other deployed package still needs it: `--prune` calls `Cluster.GetDeployedZarfPackages` to list every `DeployedPackage` secret in the cluster and checks whether the candidate image appears in any other package's component images. If nothing else references it, both tags Zarf pushes for that image are removed from the internal registry - the plain tag and the CRC-32-suffixed tag the Zarf agent uses for transparent redirection (see `images/push.go`). If another package still references it, both tags are left alone.

#### Graceful Cancellation Handling

Today, if a `zarf package deploy` or `zarf package remove` is interrupted, the `DeployedPackage` state can be left showing `Deploying`/`Removing` (or a component in the same state) indefinitely, with nothing to correct it afterward.

On a controlled stop - Zarf catching `SIGINT`/`SIGTERM` via a cancellable `context.Context` (e.g. `signal.NotifyContext`) - Zarf will attempt to persist a `Cancelled` or `RemoveCancelled` `PackageStatus`/`ComponentStatus` for whatever was in progress before exiting, mirroring the existing `Failed`/`RemoveFailed` split: `Cancelled` means the interrupted operation was a deploy, `RemoveCancelled` means it was a remove. This lets library users tell which operation was in flight when Zarf stopped, without having to separately track what command was running. This is best-effort: an ungraceful termination (`SIGKILL`, node crash, power loss) gives Zarf no opportunity to run any code, so the status can still be left stale in that case - see [Risks and Mitigations](#risks-and-mitigations).

### Test Plan

[x] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

The e2e suite will need infrastructure for simulating a controlled stop (sending `SIGINT`/`SIGTERM` to a running `zarf package deploy`/`zarf package remove`) so [Graceful Cancellation Handling](#graceful-cancellation-handling) can be exercised end-to-end rather than only unit tested. Test fixtures will also be needed for `--prune`: package definitions that differ only by an added/removed component or chart, so orphan-detection can be tested against a real upgrade rather than a synthetic diff.

##### Unit tests

- `PackageStatus`/`ComponentStatus` are set correctly for every deploy/remove outcome, including the new `Cancelled`/`RemoveCancelled` states.
- `DeployConfig.Digest()` is deterministic: repeated calls with the same config, and calls with map literals built in a different key order, all produce the same digest; the digest changes when `SetVariables`, `Values`, `NamespaceOverride`, or `ValuesOverridesMap` change, and does not change when only imperative `DeployOptions` fields change.
- `DeployConfig.Digest()` returns an error rather than panicking when given a value `encoding/json` can't marshal.
- The `--prune` orphan-detection logic correctly identifies charts/components that were removed between two `ZarfPackage` definitions (see [Chart, Component, and Image Reconciliation](#chart-component-and-image-reconciliation)), and correctly leaves an image alone if it's still referenced by any other deployed package, not just the one being pruned.
- `Source` round-trips correctly through the `DeployedPackage` secret.

##### e2e tests

- `zarf package deploy --prune` against a package that removed a component/chart from the previously deployed version, confirming the orphaned chart, component, and any now-unreferenced images are removed.
- `zarf package remove --prune`, confirming the equivalent image cleanup for components being uninstalled.
- Two packages sharing an image, with one pruned/removed: confirming the shared image's tags survive while the other package still references it, and are removed once nothing references it anymore.
- Interrupting a `zarf package deploy`/`zarf package remove` with `SIGINT`, confirming `Cancelled`/`RemoveCancelled` is persisted to the `DeployedPackage` secret.
- Deploying a package twice with identical configuration produces the same `ConfigDigest` both times; changing `SetVariables` or `Values` between deploys produces a different `ConfigDigest`.

### Graduation Criteria

`PackageStatus`, `Source`, and `ConfigDigest` are purely additive fields on `DeployedPackage` with no opt-in required, so they can graduate to GA as soon as the test plan above is complete - there's no user-facing risk in shipping them broadly. `--prune`, on the other hand, is the one part of this proposal with real destructive potential (see [Risks and Mitigations](#risks-and-mitigations)), so it should stay opt-in to mitigate this risk and to match to `kubectl` and `helm` conventions that already work this way.

### Upgrade / Downgrade Strategy

The change in Zarf behavior is optional and the state changes additive so upgrades should be able to happen automatically without breakages.  Downgrades in state would also be non-breaking since the new fields would just be stripped.  If users had opted to use the new `--prune` flag then they would need to manually migrate back.  This should be acceptable though given the adoption of `--prune` would be intentional in the first place.

This is a breaking change for SDK/library users: moving `SetVariables`, `Values`, `NamespaceOverride`, and `ValuesOverridesMap` out of `DeployOptions` and into the nested `DeployConfig` (see [Add package `ConfigDigest`](#add-package-configdigest)) will not compile against existing caller code. The fix is mechanical, though - callers move those four fields from a top-level `DeployOptions` struct literal into a `DeployConfig` struct literal, either inline or assigned to `DeployOptions.DeployConfig`. Downgrading is the same mapping in reverse. This kind of internal restructuring is similar to other breaking changes Zarf has shipped before, and should be documented in release notes the same way those were.

### Version Skew Strategy

This proposal doesn't impact how Zarf's Agent and CLI interact, so no changes are needed there.

It does introduce skew between CLI versions reading the same `DeployedPackage` secret: `PackageStatus`/`ComponentStatus` are plain strings, so an older CLI encountering a value it predates (e.g. `RemoveCancelled`) will read it as an opaque, unrecognized string rather than failing to parse it - it just won't have specific handling for it. `Source` and `ConfigDigest` are `omitempty` and additive, so an older CLI simply won't see them. No coordinated rollout is required.

## Implementation History

2026-06-01: Initial version of this document.

## Drawbacks

This proposal introduces a breaking SDK change (splitting `DeployOptions` into `DeployConfig` and imperative fields) purely to enable configuration comparison. For a proposal primarily about status tracking, requiring every library consumer to update their integration code introduces migration cost, even though the fix itself is mechanical (see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)).

`--prune` is also a larger maintenance commitment than the rest of this proposal. Correctly reconciling orphaned charts, components, and cross-package image references across arbitrary upgrade paths is nontrivial logic to get right and keep right, and a bug here can delete something a user still needed, which most Zarf features don't risk.

Finally, `ConfigDigest` approximates whether the configuration changed rather than guaranteeing it - the known numeric-formatting limitation (see [Risks and Mitigations](#risks-and-mitigations)) means it can report a change where there isn't a meaningful one. This should be documented clearly so users don't treat a digest mismatch as an authoritative diff.

## Alternatives

### Split `PackagePhase` / `PackageOutcome`

Instead of a single `PackageStatus`/`ComponentStatus` enum, phase (what Zarf is currently doing) and outcome (the result of the last completed operation) could be tracked as two separate fields, closer to Kubernetes' phase/conditions pattern:

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

This was rejected for two reasons. First, to be fully worth doing it would need to apply to both `PackageStatus` and `ComponentStatus` to keep the two normalized - doubling the breaking change this proposal already introduces (see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)). Library users would now read two fields instead of one, and Zarf would need backward-compatibility logic to translate previously-deployed packages' single-field status into the new phase/outcome pair. Second, the main benefit of separating phase from outcome is representing a phase and a health that change independently - e.g. a stable resource whose health degrades with no operation in progress. That doesn't apply to Zarf: operations only run during an active CLI or library call, and Zarf is explicitly not becoming a continuously reconciling controller (see [Non-Goals](#non-goals)), so status only ever changes during an active deploy/remove. `Idle` + `Healthy`/`Degraded` would just recombine into the same terminal values (`Succeeded`/`Failed`) a single field already provides, without representing a state the flat enum can't.

The flat enum's cost is that every new phase needs its own paired terminal states, the pattern already visible in `Failed`/`RemoveFailed` and `Cancelled`/`RemoveCancelled`. If Zarf later adds another lifecycle operation (e.g. a verify or rollback step), the enum keeps growing instead of composing; this is also why Kubernetes' own docs discourage relying solely on `.status.phase`. If Zarf ever does become a continuously reconciling controller, this tradeoff should be revisited.

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
