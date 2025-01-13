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

# ZEP-8: show manifests and values files

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

Both creators and deployers need a way to view their manifests and values files after Zarf variables are applied.

This will be accomplished through new CLI commands `zarf package show manifests [PACKAGE | DIRECTORY]`, `zarf package show values-files [PACKAGE | DIRECTORY]`

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

Viewing manifests and values files after Zarf variable templating would be useful for both creators and deployers. Catching a mistake in templating early can reduce cycle time. A Helm template is almost instant, whereas create + deploy could take several minutes to hours.

A user can achieve a similar effect to `zarf package show manifests` by decompressing a package and running `helm template` on their chart. Not only is this a poor UX, but the `helm template` may fail depending on where Zarf variable templating is used within the chart.

This feature has been highly requested in recent months:
- Request in Kubernetes Slack - https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1730229638367829
- An issue has been created for this - https://github.com/zarf-dev/zarf/issues/2631
- Defense Unicorns, an organization that relies heavily on Zarf for their deployments, has received requests for this feature in a feedback session with their partners.

### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

- View manifests or values files after Zarf variable templating and Helm templating have been applied.
- Work with both package directories and already built packages.

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->

- Accept packages pulled from a live cluster as input.

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

Introduce two new commands. `zarf package show manifests [PACKAGE | DIRECTORY]`, `zarf package show values-files [PACKAGE | DIRECTORY]`. The commands will accept either an already built package, local or remote, or a directory containing a zarf.yaml file. `show manifests` will print both the manifests from a helm chart and the manifests in the `.components[x].manifests` key. Component actions will not run during any of these commands.

Before printing the manifest for each chart the name and version of the chart will be printed. Before printing Manifests from the `.components[x].manifests` key the name of the manifests block, `.components[x].manifests[x].name`, will be printed.

These commands will not prompt for optional components, package templates, or package variables. Users will be able to specify these values using flags.

Below is the intended help text for `zarf package show manifests`. `zarf package show values-files` will include the same flags.
```
Usage:
  zarf package show manifests [ PACKAGE | DIRECTORY ] [flags]

Flags:
      --create-set stringToString   Specify package variables to set on the command line. Only applicable for package directories (KEY=value) (default [])
      --deploy-set stringToString   Specify deployment variables to set on the command line (KEY=value) (default [])
  -f, --flavor string               The flavor of components to include in the resulting package (i.e. have a matching or empty "only.flavor" key). Only applicable for package directories
      --kube-version                Override the default helm template KubeVersion when performing a package chart template
      --components                  Comma-separated list of components whose manifests should be displayed.  Adding this flag will skip the prompts for selected components.  Globbing component names with '*' and deselecting 'default' components with a leading '-' are also supported. Only applicable for already built packages
```

### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1

As a creator of Zarf packages, I want to make sure that the variables in my package are properly rendered with the expected values. I want to check this for both manifests and values files so I run `zarf package show manifests path/to/package-dir --deploy-set=MY_VAR=my-val --flavor=my-flavor` and `zarf package show values-files path/to/package-dir --deploy-set=MY_VAR=my-val --flavor=my-flavor`

#### Story 2

As a deployer of Zarf packages, I want to make sure that the variables in my package are properly rendered for both manifests and values files before I deploy so I run `zarf package show manifests zarf-package-podinfo-amd64.tar.zst --set=MY_VAR=my-val --components=my-optional-component` and `zarf package show values-files zarf-package-podinfo-amd64.tar.zst --set=MY_VAR=my-val --components=my-optional-component`

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

This command could print Zarf variables with the `sensitive` key set to true. Zarf variables are set using values that a user already has access to: user input, configuration files, or their default value in the zarf.yaml file. Given that these commands are expected to be run by a user developing a package or preparing for a deployment and not in an automated system, we deem these risks acceptable.

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

[Internal variables](https://docs.zarf.dev/ref/values/#internal-values-zarf_) will be set using the default logic except for sensitive values which do not have defaults. Sensitive values will be set to "PLACEHOLDER" instead. For example, the `ZARF_REGISTRY` variable becomes `127.0.0.1:31999`, while `ZARF_GIT_AUTH_PUSH` will be set to "PLACEHOLDER". This is done to ensure these commands can run without needing a connection to a cluster with Zarf initialized.

Manifests and values files will be printed to standard out, while all other logs and output from this command will go to stderr.

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

Given the simplicity of this feature, unit tests will be adequate. 

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
- the first Zarf release where an initial version of the ZEP was available
- the version of Zarf where the ZEP graduated to general availability
- when the ZEP was retired or superseded
-->

## Drawbacks

<!--
Why should this ZEP _not_ be implemented?
-->

The `--create-set` and `--flavor` flags would not be applicable with an already built package. Additionally, the `--components` flag would only be applicable to already built packages, but could be made to work on package directories as a future enhancement. Clear help text for each flag and erroring out when an incorrect combination is used could mitigate user confusion here.

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Change the Command Structure

There are several different ways this command could be structured differently.

- `zarf dev show manifests [DIRECTORY]`, `zarf dev show values-files [DIRECTORY]`, `zarf package show manifests [PACKAGE]`, and `zarf package show values-files [PACKAGE]`. This would make the commands less overloaded as they wouldn't take either a package or a directory. It also would ensure every flag is relevant. For example, the `--create-set` and `--flavor` would only exist for the `dev` commands. However, This increases the surface area of the CLI with four new commands. Additionally, since `zarf dev show manifests` and `zarf package show manifests` have different parent commands they would be less discoverable than if under the same parent. It's easy to imagine a user being frustrated because they've found `package show manifests` and wished it worked on package directories, without realizing `dev show manifests` exists. 

- `zarf show manifests [PACKAGE|DIRECTORY]` and `zarf show values-files [PACKAGE|DIRECTORY]`. This is the most concise option and reads well. However, introducing the new root command `show` may limit discoverability. With no other commands under `show` users may not notice the new root word.

- `zarf package show manifests [PACKAGE]` and `zarf package show definition manifests [DIRECTORY]` This would have good discoverability, being under the `package` parent. However, `zarf package show definition manifests` is long at five words, and a word like `definition` may not be clearly articulate that the command is intended for package directories.