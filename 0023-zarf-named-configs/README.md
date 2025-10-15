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

This ZEP proposes to allow configuration specific to a Zarf package deployment to be named, versioned and published to a registry so that it can simplify the deployment experience for end users.

## Motivation

This proposal comes from a desire to even further lower the barrier to entry for the deploy persona by pre-baking some deployment configuration for a Zarf package into named configurations that can be selected from.  In some environments a user deploying a Zarf package may not have system administrator experience and an SRE may want to pre-configure the package for them to make the package even more declarative and easier to manage.  Additionally many Zarf packages cross security domains and might not be able to contain their related configuration inside the package at create time.  Having a way to marry the package with the configuration within the deployment environment would help with this as well.  Some packages also may have a set of configurations that could be selected from (i.e. `device1`, `device2`, `device3`) that would be useful to manage sets of similar but different deployments.

### Goals

- Provide a way for Zarf packages to reference pre-baked configurations during deployments
- Enhance the declarative design of Zarf

### Non-Goals

- Provide configuration outside of the deployment of the package (i.e. geared towards the Ashton persona)
- Include security relevant deployment configuration in the registry (i.e. package signing keys)

## Proposal

The proposed solution introduces a new named configuration type to Zarf to allow for a managed way to provide deployment configuration for a package.  This would include most options that are available in a `zarf-config` file under `package.deploy` including the new options mentioned in [ZEP-0021](../0021-zarf-values/README.md) and [ZEP-0017](../0017-chart-namespace-overrides/README.md).  This file would be tied to a specific Zarf package name and version in addition to its own name and version.  This configuration could then be published in a registry and/or be referenced on `zarf package deploy` or `zarf dev deploy`.

### User Stories (Optional)

#### Story 1

**As** Jacquline **I want** to be able to pre-bake package configuration **so that** I can provide a more declarative package to Ashton.

**Given** I have a Zarf Package created from the below:
```yaml
metadata:
  kind: ZarfPackageConfig
  name: example
  version: 0.1.0
  namespace: example

variables:
  - name: EXAMPLE

values:
  - values-default.yaml

components:
  - name: first
    ...
  - name: second
    ...
```
**And** I have a new configuration published from the following
```yaml
# option 1
kind: ZarfNamedConfig
# option 2
kind: ZarfDeployConfig
# option 3
kind: ZarfConfig
metadata:
  name: example-config
  version: 0.1.0

package:
  name: example
  version: 0.1.0

components: [ first ]

namespace: new-namespace

set:
  EXAMPLE: example

values:
  - values-override.yaml

adopt-existing-resources: true
```
**When** I deploy that package with a `--config` like the below:
```bash
zarf package deploy oci://my-registry/example:0.1.0 --config oci://my-registry/example-config:0.1.0
```
**Then** Zarf will set the deploy options in accordance with the referenced config

### Risks and Mitigations

This would make it easy to potentially accidentally store secrets in the registry which is not desireable.  We should add documentation about this and potentially prevent the storage of variables that are marked `sensitive` in named configs.

## Design Details

<!-- This is negotiable and would like to hear others' thoughts on a local format for named configs vs having them be OCI-only -->
Named configs would be created and published through a set of new CLI commands (`zarf config create` and `zarf config publish`).  This would pull together any referenced files or necessary artifacts and either create a local `tar.zst` or publish an OCI reference similar to a package.  This config would then be referenced and applied during a package deployment.

Because package configurations refer to packages a `zarf config list oci://<registry>/<package>` command would also be added that, for an OCI reference, would list the available named configs for that package.

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

NA - This is a modification of existing behavior that should not require prerequisite testing updates.

##### Unit tests

Unit tests would need to be added to ensure that the config was passed into the package deploy library interface correctly.

##### e2e tests

Additional end to end tests would need to be added to ensure that the CLI flags and Viper config properly configured Zarf to use the published config.  Config publishing from the CLI would also need to be tested.

### Graduation Criteria

Pending review / community input these changes would be moved from alpha status and be marked as stable within Zarf's Package definition.  This would be based on user adoption of the feature and confidence in its continued stability.

### Upgrade / Downgrade Strategy

NA - There would be no upgrade / downgrade of cluster installed components

### Version Skew Strategy

NA - This proposal is an entirely new feature and does not impact existing behavior

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

This is limited to only the package deployment options and does not allow for the full range of options that one may want to change in a Zarf package.  This means that while an SRE can set configuration for a package they cannot override the package contents in any way and would need to contact the original package creator for assistance.

## Alternatives

Similar to [ZEP-0017](../0017-chart-namespace-overrides/README.md) we could allow for Zarf package remixing:

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

To allow multiple configured packages to be created from one base package. This has the potential though to introduce Zarf package sprawl and could clog a registry with references to Zarf packages that are mostly the same.

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
