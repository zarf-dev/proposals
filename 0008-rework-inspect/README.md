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

# ZEP-0008: Add zarf dev inspect and change zarf package inspect 

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
    - [Story 3](#story-3)
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

Introduce a command to enable end-users to inspect package contents. This will be implemented as the `zarf package inspect` command, with five subcommands: `definition`, `sbom`, `images`, `manifests`, and `values-files`. All the commands will work with an existing package, whether local or remote. 

Additionally, introduce a command for developers to preview a package during implementation. This will be achieved through the `zarf dev inspect` command, with three subcommands: `definition`, `sbom` and `values-files`. These commands will work a directory containing a `zarf.yaml` file. 

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

Users need an easier way to view their package definition after it's rendered by Zarf, but before `zarf package create`. A rendered package definition has had templating, imports, and flavors applied. The only path to view the rendered package definition is running `zarf package create` and viewing the printed yaml before the (y/n) prompt. Having a separate command, `zarf dev inspect definition`, improves the UX by providing users with an easier way to view the rendered package definition. It also opens the possibility of allowing `zarf package create` to proceed without requiring user confirmation.

Viewing manifests and values files after Zarf variable templating would be useful for both creators and deployers. Catching a mistake in templating early can save time. A Helm template is almost instant, whereas create + deploy could take several minutes to hours.

A user can achieve a similar effect to `zarf package inspect manifests` by decompressing their package and running `helm template` on their chart. Not only is this a poor UX, but the `helm template` may fail depending on where Zarf variable templating is used within the chart.

Features relating to these problems have been highly requested in recent months:
- Request in Kubernetes Slack - https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1730229638367829
- An issue has been created for this - https://github.com/zarf-dev/zarf/issues/2631
- Defense Unicorns, an organization that relies heavily on Zarf for their deployments, has received requests to view manifests in a feedback session with their partners.

### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

- View manifests or values files after Zarf variable templating and Helm templating have been applied.
- View rendered package definition before creation 

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->

- Accept a cluster source of the package for new functionality. Cluster source will continue to work for `zarf package inspect definition` and `zarf package inspect images`, but not for `zarf package inspect sbom`, `zarf package inspect manifests`, or `zarf package inspect values-files`
- Print out multiple types of inspects in the same command. A user will not be able to print out the package definition and manifests in the same command. 
- Support directly opening the SBOM in browser

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

The newly added commands will not have a confirmation flag and will not prompt for optional components, package templates, or package variables. Users will be able to specify these values using flags, when applicable. None of these commands will run any Zarf actions.

### zarf package inspect
`zarf package inspect` will be deprecated and replaced by the five commands below.
#### zarf package inspect definition
Displays the `zarf.yaml` definition of the given package. Will accept a package in the cluster in addition to local or OCI packages. This will mirror behavior of the current `zarf package inspect <package>` command.

#### zarf package inspect images
Lists the images of the specified package. Will accept a package in the cluster in addition to local or OCI packages. This will mirror behavior of the current `zarf package inspect <package> --list-images` command

#### zarf package inspect sbom
Extracts the package SBOM into the specified directory. If no directory is specified it will default to the current directory. This will mirror behavior of the current `zarf package inspect <package> --sbom-out` command. Note that the `zarf package inspect <package> --sbom` flag which opens the sbom in browser will be removed without replacement. 

#### zarf package inspect manifests
Templates Helm charts and displays all Kubernetes manifests. Accepts package variables and components as flags.

#### zarf package inspect values-files
Prints the values files of Helm charts. Accepts package variables and components as flags.

### zarf dev inspect
A new command `zarf dev inspect` will be introduced with the three subcommands specified below.
#### zarf dev inspect definition
Displays the 'zarf.yaml' definition after flavors, templating, and component imports are applied.

#### zarf dev inspect manifests
Templates Helm charts and displays all Kubernetes manifests. Accepts package templates, package variables, flavors, and components as flags. 

#### zarf dev inspect values-files
Prints the values files of Helm charts. Accepts package templates, package variables, flavors, and components as flags.

### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1

As a creator of Zarf packages, I want to make sure that the variables in my package are properly rendered with the expected values. I want to check this for both manifests and values files. 

#### Story 2

As a deployer of Zarf packages, I want to make sure that the variables in my package are properly rendered for both manifests and values files before I deploy. 

#### Story 3

As a creator of Zarf packages I want to see what my package definition will look like after templates, imports, and flavors are applied.

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

The `inspect manifest` and `inspect values-files` commands could print Zarf variables with the `sensitive` key set to true. Zarf variables are set using values that a user already has access to: user input or configuration files. Given that these commands are expected to be run by a user developing a package or preparing for a deployment and not in an automated system, we deem these risks acceptable.

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

For the `inspect manifests` commands, before printing the manifests of each chart, the name and version of the chart will be printed. Before printing Manifests from the `.components[x].manifests` key the name of the manifests block, `.components[x].manifests[x].name`, will be printed. 

All of these commands, besides `zarf package inspect sbom`, will provide a user with text output. The output will go to stdout, while all other logs will go to stderr. 

For commands printing deployment variables [Internal variables](https://docs.zarf.dev/ref/values/#internal-values-zarf_) will be set using the default logic except for sensitive values which do not have defaults. Sensitive values will be set to "PLACEHOLDER" instead. For example, the `ZARF_REGISTRY` variable becomes `127.0.0.1:31999`, while `ZARF_GIT_AUTH_PUSH` will be set to "PLACEHOLDER". This is done to ensure that these commands can run without requiring a connection to a cluster.

The following commands will accept a `--components` flag: `zarf package inspect manifests`, `zarf package inspect values-files`, `zarf dev inspect manifests`, and `zarf dev inspect values-files`. If `--components` is not used these commands will print out the requested resource from all components. 

Any commands running a Helm template will accept a `--kube-version` flag. This is to avoid situations where the chart [KubeVersion](https://helm.sh/docs/topics/charts/#the-kubeversion-field) field doesn't match the kube version field in Zarf which prevents templating.

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

These new commands will be implemented directly in a stable state. The commands are simple to test, do not interact with a Kubernetes cluster, and do not have state to track between runs. 

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

The current `zarf package inspect` command will become deprecated, after a year it will be removed. In cobra, when a command is added as a child command it will take priority over the parent command. For example, if a user calls `zarf package inspect images <my-package.tar.zst>` it will call the child `images` command, however if a user calls `zarf package inspect <my-package.tar.zst>` Cobra will call the parent command. Zarf will use the behavior to introduce the new commands, and leave a deprecation note on the `zarf package inspect` command.

The only functionality of `zarf package inspect` that will be removed instead of deprecated is the `--sbom` flag, which opens an HTML viewer of the SBOM in the browser. Users can instead run `zarf package inspect sbom` and then point their browser to the HTML file in their filesystem. Removing this option keeps the `zarf package inspect sbom` command simple.

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

- 2024-12-04: Document created
- 2025-01-15: Rework proposal, squashing with ZEP-0006 introduce zarf package preview  

## Drawbacks

<!--
Why should this ZEP _not_ be implemented?
-->

Since `zarf package inspect` and `zarf dev inspect` have different roots they may not be immediately discoverable. It's easy to imagine a user who doesn't know about `zarf dev inspect` so they build their package each time before running `zarf package inspect`. Still, `zarf dev find-images` has had significant usage from community members so it's reasonable to expect users to find commands under the `dev` root. In the future, we could provide tutorials going over different situations where these commands may be useful.

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Remove, instead of deprecate zarf package inspect
`zarf package inspect <package.tar.zst>` inspect will remain usable, but deprecated for a year until it's removed, see [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy). Given that the `zarf package inspect` command is not a part of the core flow, and likely not often used in automated systems, we considered removing it right away. However, given the low anticipated maintenance cost of keeping the command around, we decided to institute a deprecation process

### Change the Command Structure

There are several different ways this command could be structured differently.

- `zarf package inspect` with flags for the different commands such as `zarf package inspect --manifests` or `zarf package inspect --images`. The issue here is that many flags would apply to some commands and not others. For example, `--kube-version` and `--set` are relevant to `zarf package inspect manifests` and values-files but irrelevant to `zarf package inspect images`. By separating the commands it becomes clear which flags apply to which resources. 
- `zarf package show manifests [PACKAGE|DIRECTORY]` and `zarf package show values-files [PACKAGE|DIRECTORY]`. This limits the amount of new commands and reads well. However, we wanted to avoid commands accepting either a package, directory, or OCI source. We felt the multiple possible sources would not be intuitive. Additionally, the resulting commands would have flags that would apply to packages, but not apply to directories and vice versa.
- `zarf dev show manifests [DIRECTORY]`, `zarf dev show values-files [DIRECTORY]`, `zarf package show manifests [PACKAGE]`, and `zarf package show values-files [PACKAGE]`. This would make the commands less overloaded as they wouldn't take either a package or a directory. It also would ensure every flag is relevant. For example, the `--create-set` and `--flavor` would only exist for the `dev` commands. However, This increases the surface area of the CLI with four new commands. Additionally, since `zarf dev show manifests` and `zarf package show manifests` have different parent commands they would be less discoverable than if under the same parent.