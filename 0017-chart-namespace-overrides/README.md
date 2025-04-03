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

This ZEP proposes to enable namespace overrides for charts by leveraging the Go templating that is being designed as part of [ZEP-0021](./0021-zarf-values/README.md) in addition to making namespaces in Zarf more Helm-like.

## Motivation

Doing this allows more flexibility with certain Zarf packages where you may want to have multiples of them installed in the cluster with slightly different configurations (such as [GitLab Runners](https://github.com/defenseunicorns/uds-package-gitlab-runner)).  Right now the release namespace of any chart has to be hardcoded into the package and will be overwritten even if the chart allows namespace overrides for some manifests within the chart.  The current behavior is also different from what Helm does by default which may not be what users of Zarf expect (Helm allows the use of the `namespace` flag on install to set the Chart's namespace without it needing to be baked into the Chart).  This is made slightly more complex in Zarf because a package often contains multiple namespaces that need to be correlated across multiple Zarf primitives (such as a Helm chart and a wait action).

### Goals

- Provide a way for a Zarf package containing Helm Charts, Actions and other namespace-aware primitives to be easily installed more than once with different configurations
- Design a paradigm that feels familiar to existing Helm users

### Non-Goals

- Move away from the declarative nature of Zarf packages

## Proposal

The proposed solution is to add a new `namespace` field under the Zarf Package configuration `metadata` and to allow this to be exposed in Go templating under `charts`, `manifests`, and `actions`.  Charts and manifests would allow templating of their `namespace` and `releaseName` fields and actions would be templateable as designed in [ZEP-0021](./0021-zarf-values/README.md).

### User Stories (Optional)

#### Story 1

**As** Jacquline **I want** to be able to set namespace values **so that** I can install the same package with different configurations in different namespaces.

**Given** I have a Zarf Package created from the following
```yaml
kind: ZarfPackageConfig
metadata:
  name: example
  ref: oci://my-registry/test:0.1.0
  version: 0.1.0
  namespace: example-namespace

components:
  - name: example-component
    charts:
      - name: example-chart
        # note: we may want to do this closerr to Helm with a .Deploy or .Release prefix instead since this does not refer to what was originally in the Zarf package and may be confusing
        namespace: "{{ .Package.metadata.namespace }}"
        url: https://example.com/helm-chart
    actions:
      onDeploy:
        after:
          - wait:
              cluster:
                kind: Deployment
                name: example
                namespace: "{{ .Package.metadata.namespace }}"
                condition: "Available"
```
**When** I deploy that package with a `--namespace` flag like the below:
```bash
zarf package deploy oci://my-registry/test:0.1.0 --namespace new-namespace
```
**Then** Zarf will change the chart's release namespace to `new-namespace`
**And** Zarf will change the action's cluster namespace to `new-namespace`

### Risks and Mitigations

We would need to be careful to document and outline which fields are templateable and which are not - fields that control the contents of the Zarf package (i.e. `url`, `valuesFiles`, `name`, etc.) should not be templated since if changed they would refer to things that cannot be deployed.  Only fields that control the deployment of the chart or manifest would be templateable and this would need to be clearly communicated to reduce confusion.  It may also be useful to outline other areas that Go templating should apply within a Zarf package definition to further reduce confusion.  We also should fail fast during package create if a template is found somewhere it is not allowed - this will allow a package creator to realize that they need to make a change before they attempt to deploy the package.

## Design Details

In addition to the new package `metadata.namespace` field, the Go templates would also allow the use of Zarf Values as well for Zarf packages that needed to deploy or control different namespaces.  The main requirement driving the addition of this new field is that the Zarf package secret that is deployed to the cluster needs to be namespaced so that Zarf can continue to keep track of all of the deployments of a given package.  Without this field, package names would overlap and Zarf would "forget" which version of the package was deployed.

This proposal would retain the current mapping of a `chart` or `manifest`'s `namespace` field being tied to its release namespace. This would ensure that Helm release secrets and any templates that use the `.Release.Namespace` template would use the newly provided namespace, and that carts wouldn't affect the history or objects of prior deployments under different namespaces.  This implementation would not affect namespaces that are defined under Helm .Values as those would still be controlled by the package configuration and Zarf Variables (or Zarf Values) as they are today.

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

NA - This is a modification of existing behavior that should not require prerequisite testing updates.

##### Unit tests

Unit tests would need to be added to ensure that the go templating worked as expected.

##### e2e tests

Additional end to end tests would need to be added to ensure that the go templating worked as expected.

### Graduation Criteria

Pending review / community input these changes would be moved from alpha status and be marked as stable within Zarf's Package definition.  This would be based on user adoption of the feature and confidence in its continued stability.

### Upgrade / Downgrade Strategy

NA - There would be no upgrade / downgrade of cluster installed components

### Version Skew Strategy

NA - This proposal doesn't impact how Zarf's components interact and is only adding new features - existing behavior will continue to work

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

## Drawbacks

This furthers the use of Go templating in Zarf which has been avoided up to this point due to the potential to conflict with Helm templates.  This is discussed more in [ZEP-0021](./0021-zarf-values/README.md), though we should be careful to ensure that it is clear where this templating is allowed and whereit is not.

This requires a change to a package from the Zarf package creator to be able to deploy the package to multiple namespaces and does not allow adhoc namespace overrides like UDS CLI.  This puts more of a burden on package creators to be responsive, but also allows a package to expose a much simpler interface to deployers and allows for some issues with the original UDS design to be mitigated (i.e. availability of the namespaces in `actions`).

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

While this would allow a package deployer to override any namespaces they wanted on any package its configuration is relatively complex and that complexity needs to be specified on the host computer and cannot easily be transmitted to it.  It also does not allow namespaces to be easily present elsewhere such as in `actions`.

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

This would allow for some customization but may not provide enough flexibility in some cases.  Some clusters for security reasons will only authorize the deploy user to access specific namespaces which may not line up with what was originally in the package.  It also does not allow namespaces to be easily present elsewhere such as in `actions`.

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

While this is the most "Helm-way" option, many packages have multiple namespaces they need to deploy to (even simpler ones like `uds-package-postgres-operator`).  This would then at least require some templating to make work, but again that could have the same drawbacks as prefixes / suffixes.

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

This is similar to the proposed option but allows many Zarf packages to be stuck together.  While managing the configuration of multiple packages together in this way can be nicer for simple deployments it can get difficult to manage many deployments together where say a database may be included and need to be deployed many times because there is no true DAG behind the dependency tree.  It also does not allow namespaces to be easily present elsewhere such as in `actions`.

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
