<!--
**Note:** When your ZEP is complete, all of these comment blocks should be removed.

To get started with this template:

- [x] **Create an issue in zarf-dev/proposals.**
  When creating a proposal issue, complete all fields in that template. One of
  the fields asks for a link to the ZEP, which you can leave blank until the ZEP
  is filed. Then, go back and add the link.
- [x] **Make a copy of this template directory.**
  Name it `NNNN-short-descriptive-title`, where `NNNN` is the issue number
  (with no leading zeroes).
- [x] **Fill out as much of the zep.yaml file as you can.**
  At minimum, complete the "Title", "Authors", "Status", and date-related fields.
- [x] **Fill out this file as best you can.**
  Focus on the "Summary" and "Motivation" sections first. If you've already discussed
  the idea with the Technical Steering Committee, this part should be easier.
- [x] **Create a PR for this ZEP.**
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

## FIXME Summary

FIXME(mentions ctx api, rework)
This ZEP proposes Feature Flags for Zarf, intended to provide a configuration model for features throughout the stages
of release and deprecation, eg. "alpha", "beta", "GA", "deprecated". It introduces a ctx-based API, similar to previous
usages of feature flags in Zarf, a centralized location where all CLI feature flags are **declared** so that users may
easily find Flags and their documentation in code. Additionally, centralizing where Flags are declared allows us to
automate site documentation that corresponds one to one with the docs in code. Inspiration for the Feature fields is 
drawn from the Kubernetes project.

## Motivation

Feature flags have been discussed often in Zarf as a step on the path to v1.0.0 and we believe they provide significant
enough benefits to offset the complexity they add. Maintainer flexibility is the foremost benefit, allowing for 
experimental features to be added and iterated on. Allowing users to opt-in for alpha previews and beta testing also
provides additional ways for Zarf maintainers to get feedback and shape features into a complete state before they are
made fully available.

There's also documentation benefits. By associating features to their release version, users may
more easily track which minimum version is required to get the functionality they need. Currently, this information can
be dug up in GitHub releases and commits, but is not modeled explicitly in the codebase.

Lastly, Zarf's ZEP process is deeply tied to a staged (alpha, beta, GA) release flow, and feature flags offer an
opportunity to model and encourage that flow in the codebase itself.

### Goals

- Allow CLI and SDK users to opt in to new features with a comprehensive UX.
- Have a clear and documented reference for users on which versions of Zarf introduce or deprecate features.
- Discourage forever-features which never graduate to GA.
- Give users a clear signal when a deprecated feature will be removed soon and must be explicitly re-enabled.

### Non-Goals

- Elaborate configuration nightmares, where users don't know what is enabled or multiple flags are intertwined.
- _Build time_ feature flags. Go already provides these via build flags - features should be runtime configurable
whenever possible.

## Proposal

This ZEP proposes we model and provide an API for features flags in Zarf. Within the model, features can be declared
along the same release path as in ZEPs (alpha, beta, GA, deprecated), as well as deprecation. We will centralize where
features are declared in the Zarf codebase, while also providing clear implementation guidance and examples for
maintainers to declare new features. CLI users will be offered multiple avenues of configuration: with CLI flags,
environment variables, and config files. Finally, the docs.zarf.dev page will document 1:1 each feature as they're
declared in code (as a stretch goal, this will be fully automated).

- TODO/NOTE: Discuss global fallback to ctx in API: this section was removed while we evaluate global features, runtime
features with ctx, and potentially supporting both. Answering this question is critical to the SDK user experience.

### User Stories (Optional)

- As a Zarf maintainer I want to add a new feature, get feedback, make revisions over time before fully releasing and 
maintaining it.
- As a Zarf maintainer I want to deprecate features with clarity on when they will be removed as well as trialing removal
by disabling the feature before it's fully removed.


- As a Zarf user I want to get involved in new experimental features, and opt-in to provide alpha feedback or beta testing.
- As a Zarf user I want to continue using Zarf without my critical workflows getting disrupted by experimental features.
- As a Zarf user I want deeper clarity if a feature that I use will be deprecated.

### FIXME Risks and Mitigations

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
Below is the proposed data model for Features, including the `Feature` type and supporting fields in the `feature` pkg.
```
pkg feature 
...

type Name string
type Description string

type Enabled bool
type Default bool

# TODO Validate Since matches semver
type Since string
# TODO Validate Until matches semver
type Until string

type Stage string
var (
  Alpha Stage = "alpha"
  Beta Stage = "beta"
  GA Stage = "ga"
  Deprecated Stage = "deprecated"
)

type Feature struct {
  # Name stores the name of the feature flag.
  Name
  # Description describes how the flag is used.
  Description
  # Enabled is set if a feature is set.
  Enabled
  # Default is set if a feature is enabled by default, without being set by users.
  Default
  # Since is the version a feature is first introduced in alpha stage.
  Since
  # Until is the version when a deprecated feature is fully removed. Historical versions included.
  Until
  # Stage describes what level of done-ness a feature is. TODO describe this better
  Stage
}
```

### API
Below are descriptions of each function in the feature API and their intended UX.

#### WithDefault()
```
// WithDefault takes a context and a slice of one or many flags, inserting the features onto the default feature set. If
// a feature name is provided that is already a part of the set, then WithDefault will return an error. 
// TODO Example [{Name: "foo", Enabled true, Default: "v0.63.0", Since: "v0.60.0", Stage: GA}]
func WithDefault(ctx context.Context, features []Feature) (context.Context, error) {
  ...
}
```

#### With()
```
// With takes a context and a slice of one or many flags, inserting the features onto the feature set. If a feature name
// is provided that is already a part of the set, then With will return an error. 
// TODO Example [{Name: "foo", Enabled true, Default: "v0.63.0", Since: "v0.60.0", Stage: GA}]
func With(ctx context.Context, features []Feature) (context.Context, error) {
  ...
}
```

#### IsEnabled()
```
// IsEnabled allows users to optimistically check a feature from ctx without erroring. Useful for control flow.
func IsEnabled(ctx context.Context, name Name) bool {
  ...
}
```

#### From()
```
// From takes a ctx and the Name of a flag, and returns a full Feature type from the ctx object. If the doesn't exist,
// then it will error.
func From(ctx context.Context, name Name) (Feature, error) {
  ...
}
```

#### All(), AllDefault(), and AllEnabled()
```
// All takes a ctx and returns all flags from the ctx object.
func All(ctx context.Context) map[Name]Feature {
  ...
}

// NOTE: Getting back an empty collection from these functions is not considered an error. The intended way to check
for no flags is
// if len(All(ctx)) == 0 {
  ...
}

// AllDefault takes a ctx and returns all features with Default set, e.g. features that have not been modified by users.
func AllDefault(ctx context.Context) map[Name]Feature {
   ...
}

// AllEnabled takes a ctx and returns all features that have been enabled by users.
func AllEnabled(ctx context.Context) map[Name]Feature {
   ...
}

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

### CLI: Enabling and Disabling Flags
- Three approaches: `CLI Flags`, `Env Vars`, and `zarf-config.yaml`, with precedence in that order.
- TODO CLI UX for Listing Flags: `zarf <COMMAND> -h` (K8s uses -h)
- CLI UX enable: `--feature-enable="Foo,Bar,Baz"`
- CLI UX disable: `--feature-disable="Fizz,Buzz,Qux"`
- Env UX enable: `ZARF_FEATURE_ENABLE="Foo,Bar,Baz" zarf p create ...`
- Env UX disable: `ZARF_FEATURE_DISABLE="Fizz,Buzz,Qux" zarf p create ...`
- Config UX enable: TODO, take a look at how these flow thru viper
- Config UX disable: TODO, take a look at how these flow thru viper

### SDK: Enabling and Disabling flags
- TODO, SDK examples pending API rework

### TODO Site Docs Automation
- Generate docs.zarf.dev page with `make docs-and-schema` 
- Each entry should be fully documented with an owner and optionally an attached ZEP. We can enforce this with automation
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
TODO

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

Revision 2 encapsulates feedback on the various design decision, and completes the API for the new impl. considerations.
Namely, storing multiple feature sets, offering API facilities to query for these separately, and doing using a
global API (not through ctx injection). Not solved in this revision is docs automation, as this may become a stretch goal.

## Drawbacks

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


### TODO Using an Off-the-Shelf Library or Pre-exisitng Solution
TODO

## (TODO) Future Work

### TODO

## (TODO) Infrastructure Needed (Optional)

<!--
Use this section if you need things from the project. Examples include a new repo,
cloud infrastructure for testing or GitHub details. Listing these here
allows the process to get these resources to be started right away.
-->

TODO Implementation considerations for Test plan. Esp. additional e2e tests.