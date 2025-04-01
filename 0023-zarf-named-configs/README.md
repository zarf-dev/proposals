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

# ZEP-0023: Zarf Named Configs

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

This ZEP proposes to enable a namespace override for charts similar to the namespace override [functionality available in UDS CLI](https://uds.defenseunicorns.com/reference/bundles/overrides/#namespace).  This would allow namespaces of charts within a Zarf package to be overridden so that multiples of the same Zarf package could be deployed to the same cluster under different namespaces (without needing to maintain variants of the same package).

## Motivation

Doing this allows more flexibility with certain Zarf packages where you may want to have multiples of them installed in the cluster with slightly different configurations (such as [GitLab Runners](https://github.com/defenseunicorns/uds-package-gitlab-runner)).  Right now the release namespace of any chart has to be hardcoded into the package and will be overwritten even if the chart allows namespace overrides for some manifests within the chart.  The current behavior is also different from what Helm does by default which may not be what users of Zarf expect (Helm allows the use of the `namespace` flag on install to set the Chart's namespace without it needing to be baked into the Chart).

### Goals

- Provide a way for an already created Zarf package containing Helm Charts to be easily installed more than once with different configurations

### Non-Goals

- Move away from the declarative nature of Zarf packages

## Proposal

The proposed solution is to introduce a new named override config to Zarf to allow for a managed way to provide overrides for namespaces and eventually different values.  This allows for potential future override expansion while also forcing the overrides to be named and versioned to a package rather than be as fluid as an existing zarf-config file helping reduce declarative loss. These overrides could also eventually be signed as an artifact if desired.

### User Stories (Optional)

#### Story 1

**As** Jacquline **I want** to be able to set namespace overrides **so that** I can install the same package with different configurations in different namespaces.

**Given** I have a Zarf Package with a chart named `my-chart` in a component named `my-component`
**And** I have a new ZarfOverrideConfig created from the following
```yaml
kind: ZarfOverrideConfig
metadata:
  name: test-override
  ref: oci://my-registry/test:0.1.0
  version: 0.1.0

overrides:
  my-component:
    my-chart:
      namespace: new-namespace
```
**When** I deploy that package with a `--override` like the below:
```yaml
zarf package deploy oci://my-registry/test:0.1.0 --override oci://my-registry/test-override:0.1.0
```
**Then** Zarf will change the chart's release namespace to `new-namespace`

### Risks and Mitigations

TODO - (@WSTARR)

## Design Details

TODO - (@WSTARR) - We need to discuss format if we co with named configs.  Items of discussion:

1. How will named override configs reference the Zarf package they are attached to.
2. Would named override configs have a local format (or be OCI only)
3. How would publishing / pulling overrides work in practice

This proposal will affect the release namespace of a chart (or manifest) so that the Helm release secrets and any templates that use the `.Release.Namespace` template would use the newly provided namespace.  This would ensure that charts wouldn't affect the history or objects of prior deployments and would be able to properly install alongside one another.  This would not affect namespaces that are defined under .Values as those would still be controlled by the package configuration and Zarf variables as they are today.

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

NA - This is a modification of existing behavior that should not require prerequisite testing updates.

##### Unit tests

TODO - (@WSTARR)

##### e2e tests

TODO - (@WSTARR)

### Graduation Criteria

TODO - (@WSTARR)

### Upgrade / Downgrade Strategy

NA - There would be no upgrade / downgrade of cluster installed components

### Version Skew Strategy

NA - This proposal doesn't impact how Zarf's components interact

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

TODO - (@WSTARR)

## Alternatives



## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
