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

# ZEP-0035: Feature Flags

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

This ZEP proposes Feature Flags for Zarf, intended to provide a configuration model for features throughout the stages
of release and deprecation, eg. "alpha", "beta", "GA", "deprecated". It introduces a ctx-based API, similar to previous
usages of feature flags in Zarf, a centralized location where all CLI feature flags are **declared** so that users may
easily find Flags and their documentation in code. Additionally, centralizing where Flags are declared allows us to
automate site documentation that corresponds one to one with the docs in code. Inspiration for the Flag fields is taken
from the Kubernetes project.

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

Feature flags have been discussed often in Zarf as a step on the path to v1.0.0 and we believe they provide significant
enough benefits to offset the complexity they add. Maintainer flexibility is the foremost benefit, allowing for 
experimental features to be added and iterated on. Allowing users to opt-in for alpha previews and beta testing also
provides additional ways for Zarf maintainers to get feedback and shape features into a complete state before they are
made fully available. There's also documentation benefits. By associating features to their release version, users may
more easily track which minimum version is required to get the functionality they need. Currently this information can
be dug up in GitHub releases and commits, but is not modeled explicitly in the codebase. 

### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

- Make new features opt-in.
- Have a clear and documented reference for users on which versions of Zarf introduce new features.
- Runtime-level feature configuration
- Stretch goal: multiple avenues of configuration (code, env vars, flags, config files)

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->
- Elaborate configuration nightmares, where users don't know what is enabled or multiple flags are intertwined.
- Forever features that are endless reworked and never make it to GA.
- _Build time_ feature flags. Go already provides these via build flags - features should be runtime configurable
whenever possible.

## TODO Proposal
<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

- Add feature flags API to Zarf. (See implementation below)
- Model release stages (alpha, beta, GA) and deprecation processes in the implementation. These flags can be queried,
and modified at runtime via a global or drawn from ctx for maximal flexibility. (TODO Storage and retrieval implications)
- Centralize implementation within the CLI to provide a clear implementation guidance on how to declare feature flags.
- TODO Discuss global API
- Generate 1:1 mapping to the website docs.

### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

- As a Zarf maintainer I want to add a new feature, get feedback, make revisions over time before fully releasing and 
maintaining it.
- As a Zarf maintainer I want to deprecate features with clarity on when they will be removed as well as trialing removal
by disabling the feature before it's fully removed.


- As a Zarf user I want to get involved in new experimental features, and opt-in to provide alpha feedback or beta testing.
- As a Zarf user I want to continue using Zarf without my critical workflows getting disrupted by experimental features.
- As a Zarf user I want deeper clarity if a feature that I use will be deprecated.

### FIXME Risks and Mitigations
<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

- One critical pitfall to avoid with feature flags is creating _unexpected_ behavior, both for end users and maintainers.
  Unclear state about which flags are enabled and not, and with and without defaults. We have some solutions below under
  risks and mitigations on how to account for this.
- Developer experience (DX) and user experience (UX) are both critical feature flag adoption.
- User feedback is key.
- Rollout considerations, like can we _backport_ flags to previous features so that they can be associated with verions
or disabled? 

## FIXME Design Details
<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

### Types
Below is the proposed data model for Feature Flags, including the `Flag` type and supporting fields in the `flag` pkg.
```
pkg flag 
...

type Name string
type Description string

# TODO Validate Since matches semver
type Since string

type Stage string
var (
  StageAlpha Stage = "alpha"
  StageBeta Stage = "beta"
  StageGA Stage = "ga"
  StageDeprecated Stage = "deprecated"
)

type Flag struct {
  # Name stores the name of the feature flag.
  Name
  # Description describes how the flag is used.
  Description
  # Enabled is set if a feature is set.
  Enabled bool
  # Default is set if a feature is enabled by default, without being set by users.
  Default bool
  # Since is the version a feature is first introduced in alpha stage.
  Since
  # Until is the version when a deprecated feature is fully removed. Historical versions included.
  Until
  # Stage describes what level of done-ness a feature is. TODO describe this better
  Stage
}
```

### API
Below are descriptions of each function in the flag API and their intended UX.

#### With()
```
pkg flag 
...

// With takes a context and a slice of one or many flags, setting the context on each. Duplicated flags will not error
// and are treated as idempotent. If flags are duplicated with different fields, the value from the "latest" flag at the
// tail slice will take precedence and merge over the prior field. The value of the field is not compared.
// TODO Example [{name: "foo", version:  }]
func With(ctx, []Flag) (context.Context, error) { ... }

// DISCUSS: Instead of merging two flags and taking the latest, should we instead treat flags as a unique set, and error
// if passed a duplicate flag?
```

#### IsEnabled()
```
pkg flag 
...

// IsEnabled allows users to optimistically check a flag from ctx without regard for errors. Useful for control flow.
func IsEnabled(ctx, flagName) bool { ... }
```

#### From()
```
pkg flag 
...

// From takes a ctx and the name of a flag, and returns it from the ctx object. If the doesn't exist, then it will
// return an error.
func From(ctx, flagName) (Flag, error) { ... }

// alt, ok + empty flag on false? I feel like "ok" has gone out of style. dogsledding the flag itself isn't the worst UX
// though. Kinda just expect Flag, error if we're returning a struct type.
func From(ctx, flagName) (bool, Flag) { ... }

// e.g.
// if ok, _ := foo(); ok {
//   ...		
// }
```

#### All()
```
pkg flag 
...

// From takes a ctx and returns all flags from the ctx object.
func All(ctx) ([]Flag)

TODO alt
func All(ctx) (map[flagName string]Flag)

// NOTE: Getting back an empty collection here is not considered an error. The intended way to check for no flags is
// if len(flags) == 0 {...}
```

### TODO DISCUSS: Global API
- Enable globals? Potentially an abandoned idea.
- Proposal:
- Allow for global flags and have the API fallback to globals if no ctx is provided or ctx is empty?
- Globals have advantages for developer ease of use in the SDK. e.g. get sensible defaults for loading up Zarf as a library, don't worry about enabling anything yourself.
- Also if maintainers rely on enabling feature flags at specific versions then this keeps SDK to CLI parity.
- Without global defaults we have to flip the boolean from enabled to disabled when providing a beta feature by default.
- A major drawback with ctx is that it's difficult to inspect for users. "where did this flag come from?"
- Maybe this is a good nudge to get features into GA sooner. More to discuss here.
- Implementation:
- Not unlike logger. Atomically store a reference to a set of flags. Making the API transparent is harder. Should we
have a flag store/collection instance on ctx which is what we query? Currently we just assume it's going to be a
bare collection type, e.g. slice or map, but supporting atomic updates would be better with an abstraction.


### TODO DISCUSS Other details:
- Having a central location for flags is ideal, both to generate documentation from and to give users a centralized place
  to reference flag defaults in code.
- maybe map[Name("myFeature")]Flag
- Each entry is fully documented, has an owner, has a ZEP if it needs it.

### TODO Site Docs Automation
- Generate docs.zarf.dev page with `make docs-and-schema` 
- Parse flags in defaults table, generate markdown table from the list.
- K8s website has a good example here (link)

### TODO Test Plan

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

- Create unit test file and tables for each feature flag function.
- Test errors in unit test cases as well
- TODO e2e testing. This may require a test flag for a mock feature. A/B test before and after flag is enabled? should
be pretty lightweight
- TODO Infra implications for more testing

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

Feature flags are implemented on an opt-in basis. Upgrading to versions with feature flagging supported, and with
toggle-able feature flags ought to be simple and transparent. There are no backwards compatibility concerns for previous
versions of Zarf as feature flags deal with the code itself.

### Version Skew Strategy

<!--
If applicable, how will the component handle version skew with other
components? What are the guarantees? Make sure this is in the test plan.

Consider the following in developing a version skew strategy for this
proposal:
- Does this proposal involve coordinating behavior between components?
  - (i.e. the Zarf Agent and CLI? The init package and the CLI?)
-->
TODO Should we consider build flags as well with feature flags? Optimistically we'd be able to handle all resources in the
Go codepaths, but that assumption may not always hold. What if a new feature adds a full component or requires signific
resources like configuration?

BK: Backporting features to flags?

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
Revision 1 of this doc is intended to include Summary, Motivation, Proposal and a first stab at an API implementation.
The reason why this is all done at once, is because we have prior art with feature flagging (TODO link to PR) and this
proposal is intended to generalize and provide long term support for this approach.

## TODO Drawbacks
<!--
Why should this ZEP _not_ be implemented?
-->

- Contributor friction and increased lead time on new features (what if this is a positive because of the design work)
- Yet Another Process (yap yap yap)
- Increased end user complexity and config surface area (effective docs makes this better but doesn't solve it.)
- Increased contributor complexity when having to support backwards compatibility during new feature rollout and
deprecation strategy for the backwards compatible feature. (Again, this is kinda looking like a GOOD constraint)

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Value-based flag merges on flags.With()
This is overkill, we don't need to compare for highest version or latest feature (e.g. beta takes precedence over alpha)
for each field every time a flag is declared multiple times. A flag should be declared once, and if it is declared
multiple times, take the latest.

## (TODO) Infrastructure Needed (Optional)

<!--
Use this section if you need things from the project. Examples include a new repo,
cloud infrastructure for testing or GitHub details. Listing these here
allows the process to get these resources to be started right away.
-->

TODO Implementation considerations for Test plan. Esp. additional e2e tests.