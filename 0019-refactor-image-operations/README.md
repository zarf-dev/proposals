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

# ZEP-0019: Refactor Image operations

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

<!--
This section is key for creating high-quality, user-focused documentation
like release notes or a roadmap. You should gather this info before
implementation starts to keep the focus on development, not writing. ZEP
editors should ensure the `Summary` is clear and useful for a broad audience.

A good summary should be at least a paragraph long.

Follow the [documentation style guide] for this section and the rest of the ZEP.
Keep line lengths reasonable to make it easier for reviewers to provide
feedback and reduce unnecessary changes.

[documentation style guide]: https://docs.zarf.dev/contribute/style-guide/
-->

One of the core operations of Zarf is to push and pull container images across environments. Zarf uses Crane for this functionality, but has faced several bugs and issues in its Crane use. By switching the [images package](https://github.com/zarf-dev/zarf/tree/main/src/internal/packager/images) to utilize [oras-go](https://github.com/oras-project/oras-go) Zarf will resolve many of the issues it faces with Crane. Additionally, Zarf will reap benefits from using the same library for container image operations and Zarf OCI package operations.

## Motivation

<!--
This section is for explicitly listing the motivation, goals, and non-goals of
this ZEP.  Describe why the change is important and the benefits to users. You
can also optionally include links to [experience reports], [community slacks],
or other references to show the community's interest in the ZEP.

[experience reports]: https://go.dev/wiki/ExperienceReports
[openssf slack]: https://openssf.slack.com/archives/C07AKUMBDMJ
[kubernetes slack]: https://kubernetes.slack.com/archives/C03B6BJAUJ3
-->

Crane does not have optimal behavior and flexibility around concurrency and the Crane cache often errors when pulling non container OCI images such as Helm charts or cosign signatures. Attempts to bring up these issues have gone unanswered in the Crane repository. The OCI specification is continuously updating, and Zarf needs to ensure it can handle the changes. 

oras-go is a CNCF project with active development and use from many other projects in the community such as Helm and Flux. It solves many, if not all, of the issues we face with Crane. It is committed to stay up to date with the ocispec. Zarf already uses oras-go to publish Zarf packages so using oras-go adds no new dependencies.  

### Zarf issues related to Crane

There are several issues in the Zarf repository involving Crane: 
- Unable to use OCI artifacts that are not all image layers [#3113](https://github.com/zarf-dev/zarf/issues/3113)
- flake: failing during image pull when building podinfo-flux package in test-external [#3194](https://github.com/zarf-dev/zarf/issues/3194)
- Intermittent Hangs at crane.Push() on Registry Push [#2104](https://github.com/zarf-dev/zarf/issues/2104)

The issue in [#3113](https://github.com/zarf-dev/zarf/issues/3113) seems to stem from a similar issue as [#1955](https://github.com/google/go-containerregistry/issues/1955) where Crane does not properly handle non container OCI images in its cache. [#3194](https://github.com/zarf-dev/zarf/issues/3194) is likely caused by issues with concurrent pulls as seen in [#1941](https://github.com/google/go-containerregistry/issues/1941), this will be solved in the migration since oras-go is go routine safe. It is not easy to verify that [#2104](https://github.com/zarf-dev/zarf/issues/2104), is 100% caused by Crane rather than connection issues to the registry. Still, the oras-go migration will give users flexibility to choose how many layers they push concurrently, which has potential to improve reliability.

### Non responses to Crane issues

The team has made three bug reports in the Crane repository with reproducible steps:
- crane: incorrectly uses compressed layer of a cosign .sig file to write OCI image from cache google/go-containerregistry [#1955](https://github.com/google/go-containerregistry/issues/1955)
- ggcr: Image write concurrency errors google/go-containerregistry [#1941](https://github.com/google/go-containerregistry/issues/1941)
- ggcr: Docker with Containerd snapshotter gives wrong config name [#1954](https://github.com/google/go-containerregistry/issues/1954)

Each of these bug reports have had no responses and were automatically closed as not planned after being marked as stale. oras-go image operations are goroutine safe which solves [#1941](https://github.com/google/go-containerregistry/issues/1941). [#1955](https://github.com/google/go-containerregistry/issues/1955) stems from Crane not properly handling non-container OCI images in it's cache. The cache solution in this proposal handles non container OCI images. oras-go does not provide a native way to import Docker images. However, we will be able to avoid [#1954](https://github.com/google/go-containerregistry/issues/1954) in our custom implementation. 


### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

- Improve reliability of container image pushing and pulling.
- Introduce shared cache between Zarf package OCI operations and Zarf image OCI operations.

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->

- Remove the `zarf tools crane` command which provides users with a CLI to interact with container registries. Crane will still be a dependency of Zarf because of this. Additionally, the Syft requires Crane objects to generate SBOMs. 

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

Use the oras-go library to replace Crane for image pull and image push operations. 

Image pull and push operations will respect the global `--oci-concurrency` flag to determine how many layers to push or pull concurrently. This flag is currently used for Zarf package OCI operations, but not respected during image operations with Crane. The default `--oci-concurrency` flag value will increase to six. The default in oras-go and Zarf is currently three, but this number is conservative. [Skopeo](https://github.com/containers/skopeo), for instance, has a default of six. Crane pushes and pulls all layers concurrently always.

The oras-go implementation will only pull one image at a time. The current implementation pulls up to ten images concurrently, while this may improve speed in certain cases, in many others, Zarf is over saturating the network and worsening reliability. If in the future Zarf would like to go back to concurrently pulling images, oras-go will handle this without issue. This can be seen by the [DoOrasPullConcurrent](https://github.com/zarf-dev/image-pull-experiments/blob/main/oras/main.go#L55) function created to test this behavior.

A shared cache will be used for Zarf package OCI operations and Zarf image OCI operations, see issue [2033](https://github.com/zarf-dev/zarf/issues/2033) which requests this feature. 

### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

The general flow of users should not change, if image pulls were already working without issues for users the change should be largely not noticeable. For reliability we're hoping the extra options enable users when they face.

#### Story 1

As a user deploying packages in an environment with a poor network connection to the Zarf registry, I run `zarf package deploy --oci-concurrency=1` so that it is less likely that my image push will flake. 

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

oras-go does not provide a blob storage natively, however the ORAS CLI does. While it is marked as internal, it is simple to vendor into the Zarf project. Additionally, issue [#881](https://github.com/oras-project/oras-go/issues/881) in oras-go requests caching as part of the library. The maintainers have noted that it seems like a valuable feature to add.

oras-go does not natively support pulling images from the Docker daemon. Zarf will instead pull from the Docker daemon directly, which results in an OCI formatted tar file. Once extracted into a directory it can be treated as a normal oci-layout for use with the oras-go library. 

### Test Plan

<!--
**Note:** *Not required until targeted at a release.*
The goal is to ensure that we don't accept proposals with inadequate testing.

All code is expected to have adequate tests (eventually with coverage
expectations). Please adhere to the [Zarf testing guidelines][testing-guidelines]
when drafting this test plan.

[testing-guidelines]: https://docs.zarf.dev/contribute/testing/
-->

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

Since image pulling and pushing is a core functionality of Zarf most of the end to end tests will, by nature, test image pulling and pushing. 

Packages created with an older version of Zarf must deploy fine using a newer version of Zarf. See [Version Skew Strategy](#version-skew-strategy)

Additionally, the image push should have unit tests, and the image pull unit tests should be improved. 

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

As this is not a feature, but a refactor of an implementation this change will be introduced as stable immediately. 

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

N/A

### Version Skew Strategy

<!--
If applicable, how will the component handle version skew with other
components? What are the guarantees? Make sure this is in the test plan.

Consider the following in developing a version skew strategy for this
proposal:
- Does this proposal involve coordinating behavior between components?
  - (i.e. the Zarf Agent and CLI? The init package and the CLI?)
-->

There are differences between the annotations on the index.json for images pulled by Crane vs ORAS. Zarf must continue to work in the following situations:

- A package created with older version of Zarf, using Crane, must deploy with newer version of Zarf, using ORAS.
- A package created with a newer version of Zarf, using ORAS, must deploy with an older version of Zarf, using Crane.

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

- 2025-03-05: Document created

## Drawbacks

<!--
Why should this ZEP _not_ be implemented?
-->

There is potential for image operations to be slower for some use cases since we are defaulting to less concurrency. This is a lower risk than the current reliability issues that users face which can completely break create + deploy operations. 

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Alternative libraries to ORAS

[containers/image](https://github.com/containers/image) was looked at as an alternative replacement for image operations. [Skopeo](https://github.com/containers/skopeo/) uses this tool for most of it's image operations and has built a large user base. containers/image has a builtin way to extract images from the docker daemon. However, containers/image does not have a blob cache which is a feature many users rely on to quickly iterate on packages. 

[Regclient](https://github.com/regclient/regclient) was evaluated, but since it also lacked a blob cache it was ruled out. 

### Alternative concurrency model

As discussed in the proposal this implementation will only pull one image at a time so that the network is less likely to be over saturated while pulling images. One potential idea would be to dynamically set the concurrency depending on the size of images being pulled. For example, Zarf could pull as many layers concurrently as possible until the total size of concurrent layers is greater than 1 Gigabyte. For now this is not done to avoid extra complexity. 