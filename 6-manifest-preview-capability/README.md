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

# ZEP-NNNN: Your short, descriptive title

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

Introduce a new command called `zarf package preview` which will have the ability to display a zarf package yaml after templated, imports, and flavors are applied. Alongside manifests or values files that will be deployed. Additionally, enhance the `zarf package inspect` command to give it the ability to display manifests or values files.

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

There is no easy way to see a zarf.yaml file that will be created after templating and importing are finished outside of `zarf package create`. Furthermore, there is not a convenient way to see the manifests and values files that will be deployed by a Zarf package. 

This feature would put is in parity with other tools that have similar pre-deploy viewing capabilities. The most obvious example is `helm template`, as Zarf would call the Helm library template functions within `zarf package preview`. Other tools have similar capabilities such as opentofu with `tf plan`.

This feature has been highly requested in recent months:
- request in Kubernetes slack - https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1730229638367829
- An issue has been created for this - https://github.com/zarf-dev/zarf/issues/2631
- Defense Unicorns, an organization that relies heavily on Zarf for their deployments, have received requests for this feature in a feedback session with their partners.

### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

The goal for creators is that they have all the information they need to build a working package as quickly as possible. 

The goal for deployers is that they feel confident that they know what they're deploying and that the Zarf variable templating will work properly.

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->

We do not want to show the manifests of a package that is already deployed. `zarf package inspect` works on packages in the cluster, if a user tries the `--show-manifests` flag or `--values` flag for a package in the cluster they will receive an error. There is already a [helm get manifest](https://helm.sh/docs/helm/helm_get_manifest/) command that the user may use when already deployed to a cluster. 

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

Introduce a new command called `zarf package preview` command. This command will print a zarf.yaml after templates, imports, and flavors are applied. The help text and flags of this command would look like below. Importantly, the `--show-manifests` and `--show-values-files` flags would print out what the manifests and values files would look like after templating has occurred. 

```
Usage:
  zarf package preview [ PACKAGE ] [flags]

Flags:
      --create-set stringToString   Specify package variables to set on the command line (KEY=value) (default [])
      --deploy-set stringToString   Specify deployment variables to set on the command line (KEY=value) (default [])
  -f, --flavor string               The flavor of components to include in the resulting package (i.e. have a matching or empty "only.flavor" key)
      --show-manifests              shows the manifests that would be deployed by this package
      --show-values-files           shows the values files that would be used by the charts in this package
```


### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1

As a creator of Zarf packages I want to see what my zarf.yaml will look like after templates, imports, and flavors are applied so I run `zarf package preview -f my-flavor --create-set=MY_TEMPLATE=my-val`

#### Story 2

As a creator of Zarf packages I want to make sure the variables in my package can get templated properly for the expected values of the deployers. I want to check this for both manifests and values files. `zarf package preview --deploy-set=MY_VAR=my-val --show-manifests` or `zarf package preview --deploy-set=MY_VAR=my-val --show-values-file`

#### Story 3

As a deployer of Zarf packages, I want to check that the variables I intend to deploy my package with are getting properly templated for both manifests and values files. 

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

This will not cause any security impacts. There is no sensitive data that this will print out that the user does not already have access to. This command will be able to print Kubernetes secrets to the console, but given this is expected given that it is printing manifests. 

This command could print Zarf variables with the `sensitive` value set to true. The value to these variables can be assigned with user input, config files, and the `default` key in the zarf.yaml. Given that these commands are expected to be run by a user actively managing a cluster or developing a package and not in an automated system we deem these risks acceptable.

**QUESTION** Am I making too broad an assumption that this risk is acceptable?  

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

[Internal variables](https://docs.zarf.dev/ref/values/#internal-values-zarf_) will be set using the default logic. This means the `ZARF_REGISTRY` variable will become `127.0.0.1:31999`, while secrets like the `GIT_AUTH_PUSH` variable will become a random string. This will follow the same logic as `zarf dev find-images`. 

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

##### Prerequisite testing updates

<!--
Based on reviewers feedback describe what additional tests need to be added prior
implementing this enhancement to ensure the enhancements have also solid foundations.
-->

N/A

##### Unit tests

<!--
In principle every added code should have complete unit test coverage, so providing
the exact set of tests will not bring additional value.
However, if complete unit test coverage is not possible, explain the reason of it
together with explanation why this is acceptable.
-->

<!--
Additionally, for Alpha try to enumerate the core package you will be touching
to implement this enhancement and provide the current unit coverage for those
in the form of:
- <package>: <date> - <current test coverage>
The data can be easily read from:
https://app.codecov.io/gh/zarf-dev/zarf


This can inform certain test coverage improvements that we want to do before
extending the production code to implement this enhancement.
-->

- `<package>`: `<date>` - `<test coverage>`

##### e2e tests

<!--
This question should be filled when targeting a release.
For Alpha, describe what tests will be added to ensure proper quality of the enhancement.

For Beta and GA, add links to the created E2E test(s) if applicable

We expect no non-infra related flakes in the last month as a GA graduation criteria.
-->

Given the simplicity of this feature, unit tests will be adequate. 

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

N/A

## Implementation History

<!--
Major milestones in the lifecycle of a ZEP should be tracked in this section.
Major milestones might include:
- the `Summary` and `Motivation` sections being merged, signaling acceptance of the ZEP
- the `Proposal` section being merged, signaling agreement on a proposed design
- the date implementation started
- the first Kubernetes release where an initial version of the ZEP was available
- the version of Kubernetes where the ZEP graduated to general availability
- when the ZEP was retired or superseded
-->

## Drawbacks

<!--
Why should this ZEP _not_ be implemented?
-->

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

One alternative was to use `zarf dev preview`, the thought being that this command will be run by creators while developing a package. However, given the similarities between this command and `zarf package inspect` we decided that it made for a more cohesive user experience to have both commands under the same parent.

I pondered if we should accept a component flag as we do on deploy during package inspect or package preview. This would allow people to get a more accurate view of the manifests that they intend to deploy if they are using optional components. However, this flag would only make sense alongside the  It does make the code more complicated, but it can give a better view of what will actually be deployed. Given the added complexity, we are not added this for now, but it could be an enhancement in the future. 

Another alternative would be to have completely separate commands to show the manifests. The `zarf package preview` command would exist solely to preview a package and the `zarf package inspect` command would be unchanged. Instead we'd introduce the commands `zarf dev show-manifests` and `zarf dev show-values-files`. These commands would have the `--deploy-set` and `--create-set` flags, while `zarf package preview` would only need `--set`. Additionally, this would make it easier to add a flag like `--components` which wouldn't make sense for the `zarf package preview` command as all components are included on create regardless of if they are used or not. The `show-manifests` and `show-values-files` commands would take either a zarf.yaml or a zarf package. This would also 
