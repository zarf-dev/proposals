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

# ZEP-0039: Enhance YOLO Experience

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
<!-- /toc -->

## Summary

This ZEP proposes a streamlined method of publishing Yolo and Airgap packages based on a single `Zarf.yaml` file. The intention is to provide a means of distributing a single configured package regardless of whether or not the destination environment is behind an airgap, similar to how packages for different architectures are distributed.

## Motivation

As a user of Zarf and a maintainer of a large package which may or may not be distributed to internet-connected or airgapped environments, I would like a way to distribute this large package without managing multiple `zarf.yaml` files. The zarf package I maintain includes a few Helm Charts which are deployed with Flux. Using Flux, users are able to add additional values.yaml files to these Helm charts at deploy time. The airgap package is 6GB in size, which is completely unecessary for my internet-connected users, who only require the Yolo package containing the Helm charts.

I would like use Zarf as a tool to create packages for both of these environments automatically, however the experience so far has been a lot of work and scripting to generate multiple `zarf.yaml` files and ensure they are correctly formatted and accurately implement the package I am trying to distribute. A process which in practice has not been foolproof.

Zarf should support Yolo packages natively, and in my opinion as a standard part of the packaging process and OCI image construction.

[kubernetes slack]: https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1754333786975669

### Goals

- Update the zarf package create process to build a Yolo package alongside the airgap package natively
- Allow users to specify Yolo-Only and Airgap-Only conditions/predicates for components in their `zarf.yaml`
- Remove `yolo: true/false` property from `zarf.yaml` schema (since this should be a toggle on package create/deploy)

### Non-Goals

- Integrating new Yolo features in the deploy process. In my opinion, this proposal should only apply to the creation of new packages as Zarf already has a `zarf dev deploy` command to emulate the deployment of local packages, and creating a Yolo package can be deployed already.
- Change the name `YOLO Mode` to something indicating production ready, internet-connected environments. That change should be a separate proposal and take feedback from usage of features in this proposal.

## Proposal

The most significant change to Zarf would be how it handles packaging at create, as once this optional feature is enabled, after creating a standard zarf package, it would remove all images from and create an additional Yolo package. Using optimized OCI layers, this Yolo package could be a subset of the airgap package layers, or the Airgap package could use the Yolo layers as a base. Additionally, users taking advantage of this feature should be able to specify whether or not an action, file, manifest, etc. should only be included in the Yolo package, the Airgap package, or both. A default naming convention should probably be used to standardize the name of the packages, similar to the package architecture.

### User Stories (Optional)

- As a package maintainer, I would like my release pipeline to generate zarf packages for both airgap and internet-connected automatically

- As a user, I would like more zarf packages to be available in Yolo-mode when I do not require airgap features

- As someone who thinks Zarf is pretty damn awesome, I would like to use it to manage my deployments to not just my airgap environments, but my cloud environments as well.

- As a user deploying into an internet-connected environment, I don't want to waste bandwidth downloading 6GB worth of images because the package maintainers didn't automatically build and publish a Yolo package.

### Risks and Mitigations

- Removing the `yolo` property would be a breaking change. An alternative to removing the property would be to disable this proposed feature if that flag exists.

- There may be some inherint differences in how packages are able to be deployed between airgap and connected environments, which may require constant feature updates to this feature to support.

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

### Test Plan

<!--
**Note:** *Not required until targeted at a release.*
The goal is to ensure that we don't accept proposals with inadequate testing.

All code is expected to have adequate tests (eventually with coverage
expectations). Please adhere to the [Zarf testing guidelines][testing-guidelines]
when drafting this test plan.

[testing-guidelines]: https://docs.zarf.dev/contribute/testing/
-->

[ ] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

### Graduation Criteria

<!--
**Note:** *Not required until you're targeting a release.*

Define what needs to happen for this feature to move from alpha to beta to GA
(General Availability). Focus on key signals or criteria that show the feature
is ready for each stage.

Consider the following stages when setting graduation criteria:
- Alpha: Feature is behind a feature flag, basic tests in place.
- Beta: Gather feedback from users, complete core features, add more tests.
- GA: Prove real-world usage, complete rigorous testing, gather feedback.

In general, features should wait at least two releases between Beta and GA to
allow time for feedback. For features moving to GA, include conformance tests
to ensure stability and compatibility.

#### Deprecation
If this feature will eventually be deprecated, plan for it:
- Announce deprecation and support policy.
- Wait at least two versions before fully removing it.
-->

### Upgrade / Downgrade Strategy

<!--
If applicable, how will the component be upgraded and downgraded? Make sure
this is in the test plan.

Consider the following in developing an upgrade/downgrade strategy for this
proposal:
- What changes (in invocations, configurations, API use, etc.) is an existing
  package definition or deployment required to make on upgrade, in order to
  maintain previous behavior?
- What changes (in invocations, configurations, API use, etc.) is an existing
  package definition or deployment required to make on upgrade, in order to
  make use of the proposal?
-->

### Version Skew Strategy

<!--
If applicable, how will the component handle version skew with other
components? What are the guarantees? Make sure this is in the test plan.

Consider the following in developing a version skew strategy for this
proposal:
- Does this proposal involve coordinating behavior between components?
  - (i.e. the Zarf Agent and CLI? The init package and the CLI?)
-->

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

<!--
Why should this ZEP _not_ be implemented?
-->

## Alternatives

- Manage Yolo and Airgap packages separately with two different `zarf.yaml` files. This is error prone creates issues with testing changes properly between Yolo and Airgap deployments. It opens the door to drift between versions. Using this approach either requires two declaritive `zarf.yaml` files with lots of duplication, or a declarative `zarf.yaml` that dynamically updates, and is therefore not declarative.
  [Slack suggestion](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1754485517547389?thread_ts=1754333786.975669&cid=C03B6BJAUJ3)

- Make `yolo` property of zarf schema accept a string instead of a boolean. This would allow a zarf template variable to dynamically change the yolo-mode at package create time. This is a decent enough workaround, but wouldn't enhance the capability of Zarf's Yolo-Mode experience. Additionally, it wouldn't provide opportunities to make yolo/airgap specific changes to components.
