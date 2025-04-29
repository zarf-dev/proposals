<!--
**Note:** When your ZEP is complete, all of these comment blocks should be removed.

To get started with this template:

- [ ] **Create an issue in zarf-dev/proposals.**
  When creating a proposal issue, complete all fields in that template. One of
  the fields asks for a link to the ZEP, which you can leave blank until the ZEP
  is filed. Then, go back and add the link.
- [ ] **Make a copy of this template directory.**
  Name it `NNNN-short-descriptive-title`, where `NNNN` is the issue number
  (with no leading zeroes).
- [ ] **Fill out as much of the zep.yaml file as you can.**
  At minimum, complete the "Title", "Authors", "Status", and date-related fields.
- [ ] **Fill out this file as best you can.**
  Focus on the "Summary" and "Motivation" sections first. If you've already discussed
  the idea with the Technical Steering Committee, this part should be easier.
- [ ] **Create a PR for this ZEP.**
  Assign it to members of the Technical Steering Committee who are sponsoring this process.
- [ ] **Merge early and iterate.**
  Don’t get bogged down in the details—focus on getting the goals clarified and the
  ZEP merged quickly. You can fill in the specifics incrementally in later PRs.

Just because a ZEP is merged doesn't mean it's complete or approved. Any ZEP marked
as `provisional` is a working document and subject to change. You can mark unresolved
sections like this:

```
<<[UNRESOLVED optional short context or usernames ]>>
Stuff that is being argued.
<<[/UNRESOLVED]>>
```

When editing ZEPs, aim for focused, single-topic PRs to keep discussions clear. If
you disagree with a section, open a new PR with suggested changes.

Each ZEP covers one "feature" or "enhancement" throughout its lifecycle. You don’t
need a new ZEP for moving from beta to GA. If new details emerge, edit the existing
ZEP. Once a feature is "implemented", major changes should go in new ZEPs.

The latest instructions for this template can be found in [this repo](/NNNN-zep-template/README.md).

**Note:** PRs to move a ZEP to `implementable`, or significant changes to an
`implementable` ZEP, must be approved by all ZEP approvers. If an approver is no
longer appropriate, updates to the list must be approved by the remaining approvers.
-->

# ZEP-0017: Chart Namespace Overrides

<!--
Keep the title short simple and descriptive. It should clearly convey what
the ZEP is going to cover.
-->

<!--
A table of contents helps reviewers quickly navigate the ZEP and highlights
any additional information provided beyond the standard ZEP template.
-->

<!-- toc -->
- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
  - [Non-Goals](#non-goals)
- [Proposal](#proposal)
  - [User Stories (Optional)](#user-stories-optional)
    - [Story 1](#story-1)
    - [Story 2](#story-2)
  - [Risks and Mitigations](#risks-and-mitigations)
- [Design Details](#design-details)
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
- [Infrastructure Needed (Optional)](#infrastructure-needed-optional)
<!-- /toc -->

## Summary

This ZEP proposes to enable namespace overrides for Zarf packages by adding a `--namespace` flag that will support overriding single-namespace packages.

## Motivation

Doing this allows more flexibility with certain Zarf packages where you may want to have multiples of them installed in the cluster with slightly different configurations (such as [GitLab Runners](https://github.com/defenseunicorns/uds-package-gitlab-runner)).  Currently, the release namespace of any chart has to be hardcoded into the package and will be used even if the chart allows namespace overrides via values for some manifests within the chart.  The current behavior is also different from what Helm does by default which may not be what users of Zarf expect (Helm allows the use of the `namespace` flag on install to set the Chart's release namespace without it needing to be baked into the Chart).  This is made slightly more complex in Zarf because a package could contain multiple namespaces that need to be correlated across multiple Zarf primitives (such as a Helm chart and a wait action) and these namespaces may differ across components.

### Goals

- Provide a way for a Zarf package containing Helm Charts, Actions and other namespace-aware primitives to be easily installed more than once with different configurations
- Design a paradigm that feels familiar to existing Helm users

### Non-Goals

- Move away from the declarative nature of Zarf packages
- Create a solution that will work for multi-namespace packages (90% of packages in the wild are single namespace across their chart and manifest resources so the initial implementation doesn't need to cover this)

## Proposal

The proposed solution is to add a new `namespace` flag that would allow a user to set a namespace for a Zarf package and override the release namespace of all charts and manifests in that package globally.  This value would also be available through the templates in [ZEP-0021](./0021-zarf-values/README.md) though no additional templating would be added elsewhere.

### User Stories (Optional)

#### Story 1

**As** a platform engineer **I want** to be able to set namespace values **so that** I can install the same package with different configurations in different namespaces.

**Given** I have a Zarf Package created from the following
```yaml
kind: ZarfPackageConfig
metadata:
  name: example
  version: 0.1.0

components:
  - name: example-component
    charts:
      - name: example-chart
        namespace: "my-namespace"
        url: https://example.com/helm-chart
    manifests:
      - name: example-manifest
        namespace: "my-namespace"
        files:
          - "example.yaml"
    actions:
      onDeploy:
        after:
          - wait:
              cluster:
                kind: Deployment
                name: example
                namespace: "{{ .Deploy.Namespace }}"
                condition: "Available"
          - cmd: "echo {{ .Deploy.Namespace }}"
```
**When** I deploy that package with a `--namespace` flag like the below:
```bash
zarf package deploy oci://my-registry/test:0.1.0 --namespace new-namespace
```
**Or When** I deploy that package with a `zarf-config.yaml` like the below:
```yaml
package:
  deploy:
    namespace: new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`
**And** Zarf will change the manifest's release namespace to `new-namespace`
**And** Zarf will change the action's cluster namespace to `new-namespace`
**And** Zarf will echo `new-namespace`

### Risks and Mitigations

We would need to be careful to document and clearly outline which fields are overriden in these cases since `charts` and `manifests` would be forced to the new value but actions would need to use templating to access the new value.

Additionally we would not want this to work for a Zarf package that declared multiple namespaces since that may have unintended consequences and squish deployments on top of each other.  To mitigate this we should have a pre-deployment validation that will fail if the `--namespace` flag is given to a package with multiple namespaces defined.

## Design Details

When the new `--namespace` flag is given to a Zarf package, Zarf will override the release namespace of all charts and manifests in that package with the new namespace value.  This value will also be available through the templates in [ZEP-0021](./0021-zarf-values/README.md). This would be global and occur across all components defined within a package.

Zarf will also use the namespace value to store a separate state Secret for the package.  This would allow the package to be deployed multiple times with different namespaces and the state Secret would be unique to each deployment allowing for proper inspection and removal of the packages.

This proposal would also retain the current mapping of a `chart` or `manifest`'s `namespace` field being tied to the chart's release namespace. This would ensure that Helm release secrets and any templates that use the `.Release.Namespace` template would use the newly provided namespace, and that updates wouldn't affect the history or objects of prior deployments under different namespaces.  This implementation would not affect namespaces that are defined under Helm `.Values` as those would still be controlled by the package configuration and Zarf Variables (or Zarf Values) as they are today.

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

NA - This is a modification of existing behavior that should not require prerequisite testing updates.

##### Unit tests

Unit tests would need to be added to ensure that the namespaces are updated/tracked correctly in library code.

##### e2e tests

Additional end to end tests would need to be added to ensure that the namespace override takes effect correctly and maintains the ability to inspect and remove multiple packages.

### Graduation Criteria

Pending review / community input these changes would be moved from alpha status and be marked as stable within Zarf's Package definition.  This would be based on user adoption of the feature and confidence in its continued stability.

Additionally we will need to evaluate how we might implement multi-namespace support in the future as we look to bring this feature out of alpha.  This status will need to be clearly communicated as this change may also have breaking changes to this proposed feature when implemented.  When this redesign happens we would want to strongly consider the implications on state and ideally move to a better model such as the one proposed in [ZEP-0026](./0026-enhanced-state-management/README.md).

### Upgrade / Downgrade Strategy

NA - There would be no upgrade / downgrade of cluster installed components

### Version Skew Strategy

Because this will impact the storage of state in the cluster, we will need to be careful to ensure that the new behavior to add namespaces to the package secret does not break older versions of the CLI.  Packages deployed with namespace overrides only need to be inspectable or removeable with the current CLI but we should test to ensure that these operations are not broken when interacting with the cluster from an older CLI.

## Implementation History

<!--
Major milestones in the lifecycle of a ZEP should be tracked in this section.
Major milestones might include:
- the `Summary` and `Motivation` sections being merged, signaling acceptance of the ZEP
- the `Proposal` section being merged, signaling agreement on a proposed design
- the date implementation started
- the first Zarf release where an initial version of the ZEP was available
- the version of Zarf where the ZEP graduated to general availability
- when the ZEP was retired or superseded
-->

- 2025-02-03: Initial version of this document

## Drawbacks

This proposal only supports single-namespace packages (which from research seems to be most packages) but may be limiting short term.  This also means that this feature will need to be redesigned once more feedback is gathered which may result in breaking changes at that time.

Overriding packages globally may also be slightly confusing to users used to Helm and it's templates since generally manifests are deployed to the namespace that is templated, not to a force-overridden namespace.  This can be mitigated with documentation but should be taken into account when designing multi-namespace support.

## Alternatives

Below are some alternatives that were also discussed but were ultimately not chosen.

### Option 1 (Flatter Namespaces)

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**When** I deploy that package with a `zarf-config.yaml` like the below*:
```yaml
package:
  deploy:
    namespaces:
      my-component:
        my-chart: new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**When** I deploy that package with a `--namespace` like the below:
```bash
zarf package deploy zarf-package-test.tar.zst --namespace my_component.my_chart=new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

#### Reason for Rejection

While this would allow a package deployer to override any namespaces they wanted on any package its configuration is relatively complex and that complexity needs to be specified on the host computer and cannot easily be transmitted to it.  It also would complicate the templating of the namespace elsewhere such as in `actions`.

### Option 2 (UDS CLI Config Style)

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**When** I deploy that package with a `zarf-config.yaml` like the below*:
```yaml
package:
  deploy:
    overrides:
      my-component:
        my-chart:
          namespace: new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**When** I deploy that package with a `--overrides` like the below:
```bash
zarf package deploy zarf-package-test.tar.zst --overrides my-component.my-chart.namespace=new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

#### Reason for Rejection

This is similar to option #1 and has the same flexibility but also all of the same drawbacks.  Additionally it opens the door for more overrides in Zarf configs which because they are simply files on the system are not directly tracked/managed and could break the declarative nature of a package (i.e. if UDS CLIs value overrides were directly implemented here too).

### Option 3 (Zarf Variable Style)

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**And** the chart's namespace contains the variable `MY_CHART_NAMESPACE`
```yaml
variables:
  - name: MY_CHART_NAMESPACE
    default: the-namepace
...
 charts:
  - name: my-chart
    namespace: "###ZARF_VAR_MY_CHART_NAMESPACE###"
```
**When** I deploy that package with a `zarf-config.yaml` like the below*:
```yaml
package:
  deploy:
    set:
      MY_CHART_NAMESPACE: new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**And** the chart's namespace contains the variable `MY_CHART_NAMESPACE`
```yaml
variables:
  - name: MY_CHART_NAMESPACE
    default: the-namepace
...
 charts:
  - name: my-chart
    namespace: "###ZARF_VAR_MY_CHART_NAMESPACE###"
```
**When** I deploy that package with a `--set` like the below:
```bash
zarf package deploy zarf-package-test.tar.zst --set MY_CHART_NAMESPACE=new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

#### Reason for Rejection

This is one of the more "Zarf-way" options however, this would be the only place in the Zarf package definition that we allowed the `###ZARF_VAR` syntax to exist and the ### syntax is not very conducive to being maintained in a Zarf package (i.e. it is  treated as a comment when unquoted adding complexity for users and/or when trying to parse Zarf packages since this cannot be pre-templated like other `###ZARF_PKG_TMPL`s).

### Option 4 (Prefix/Suffix Style)

**Given** I have a Zarf Package with a chart deploying to `my-namespace`
**When** I deploy that package with a `zarf-config.yaml` like the below*:
```yaml
package:
  deploy:
    namespace-prefix: new-
    namespace-suffix: -kitteh
```
**Then** Zarf will change the chart's release namespace to `new-my-namespace-kitteh`

**Given** I have a Zarf Package with a chart deploying to `my-namespace`
**When** I deploy that package with a `--set` like the below:
```bash
zarf package deploy zarf-package-test.tar.zst --namespace-prefix new- --namespace-suffix -kitteh
```
**Then** Zarf will change the chart's release namespace to `new-my-namespace-kitteh`

#### Reason for Rejection

This would allow for some customization but may not provide enough flexibility in some cases.  Some clusters for security reasons will only authorize the deploy user to access specific namespaces which may not line up with what was originally in the package.  It also does not allow namespaces to be easily templated elsewhere such as in `actions`.

### Option 5 (Package Namespace Style)

**Given** I have a Zarf Package that implements a new package `namespace` field
```
kind: ZarfPackageConfig
metadata:
  name: test
  namespace: my-namespace
```
**And** That package has `chart` and `manifest` resources that omit `namespace`
```
 charts:
  - name: my-chart
    url: oci://...
```
**When** I deploy that package with a `zarf-config.yaml` like the below*:
```yaml
package:
  deploy:
    namespace: new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

**Given** I have a Zarf Package that implements a new package `namespace` field
```
kind: ZarfPackageConfig
metadata:
  name: my-package
  namespace: my-namespace
```
**And** That package has `chart` and `manifest` resources that omit `namespace`
```
 charts:
  - name: my-chart
    url: oci://...
```
**When** I deploy that package with `--namespace` like the below:
```bash
zarf package deploy zarf-package-test.tar.zst --namespace new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

#### Reason for Rejection

While this is the most "Helm-way" option, this pushes further into packages having a single namespace which may not be desireable - since the namespace is implicit and not a template it also may be confusing to some users.  Also if the namespace was templated this would need to be clearly communicated since many parts of a Zarf package should not be templated at deploy time (i.e. `images`).

### Option 6 (UDS Bundle Style)

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**And** I have a new ZarfBundleConfig created from the following
```yaml
kind: ZarfBundleConfig
metadata:
  name: test-override
  version: 0.1.0
    
packages:
  - ref: oci://my-registry/test:0.1.0
    overrides:
      my-component:
        my-chart:
          namespace: new-namespace
```
**When** I deploy that bundle with like the below:
```bash
zarf bundle deploy zarf-bundle-test.tar.zst
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

#### Reason for Rejection

While managing the configuration of multiple packages together in this way can be nicer for simple deployments it can get difficult to manage many deployments together where say a database may be included and need to be deployed many times because there is no true DAG behind the dependency tree.  It also does not allow namespaces to be easily templated elsewhere such as in `actions`.

### Option 7 (Package Remix Style)

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**And** I have a new ZarfRemixConfig created from the following
```yaml
kind: ZarfRemixConfig
metadata:
  name: test-override
  version: 0.1.0
  ref: oci://my-registry/test:0.1.0
    
remix:
  my-component:
    my-chart:
      namespace: new-namespace
```
**When** I create a new package from that with:
```bash
zarf package create zarf-remix.yaml
```
**Then** Zarf will change the chart's release namespace to `new-namespace` in the new package
**And When** I deploy that package
**Then** the chart will be in the `new-namespace` namespace.

#### Reason for Rejection

This option is most Kustomize-like and could allow packages to be rebuilt with even more options than just namespaces and values but could also lead to a lot of package sprawl in more complex environments since there is not another artifact really besides the Zarf package (since the remix config just creates another Zarf package).

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
