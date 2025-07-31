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

## Summary

This ZEP proposes Feature Flags for Zarf's CLI and SDK. Features are intended to provide a configuration model for users
to enable and disable logical features (e.g. a new API or behavior) throughout the stages of release and deprecation,
eg. `alpha`, `beta`, `ga`, `deprecated`. It introduces a new go package `feature` implementing storage and API for using
features. `feature` also contains a centralized location where default features are _declared_, ensuring users and
maintainers can easily find features and their associated documentation in code. Inspiration for the Feature model is
drawn from the Kubernetes project, with various changes made to scale back the implementation to Zarf's needs.

## Motivation

Feature flags have often been discussed in Zarf as a step on the path to v1.0.0. We believe they provide significant
enough benefits to offset the complexity they add. Maintainer flexibility is the foremost benefit, allowing for 
experimental features to be added and iterated on. Making opt-in alpha previews and beta testing also provides
additional ways for Zarf maintainers to get feedback and shape features into a complete state before they are made
fully available. We hope this increases trust in and awareness of new features and encourages further experimentation.

There's also documentation benefits. By associating features to their release version, users may
more easily track which minimum version is required to get the functionality they need. Currently, this information can
be dug up in GitHub releases and commits, but is not modeled explicitly in the codebase.

Lastly, Zarf's ZEP process is deeply tied to a staged (alpha, beta, GA) release flow, and feature flags offer an
opportunity to model and encourage that flow in the codebase itself.

### Goals

- Allow CLI and SDK users to opt in to new features with a comprehensive UX.
- Have a clear and documented reference for users on which versions of Zarf introduce or deprecate features.
- Encourage community feedback on features, helping them graduate out of beta to GA.
- Give users a clear signal when a deprecated feature will be removed soon and must be explicitly re-enabled.

### Non-Goals

- Elaborate configuration nightmares, where users don't know what is enabled or multiple flags are intertwined.
- _Build time_ feature flags. Go already provides build flags - features offers code-level runtime configuration.
- Mutable user-set flags. This is an explicit design decision to keep complexity low for maintainers and SDK users. If
the need for mutable flags arises in the future, it is possible to ease the restrictions in Set and add merge behavior.
(See "Value-based Merges of Feature Structs" under Alternatives)

## Proposal

This ZEP proposes we model and provide an API for features flags in Zarf. Within the model, features can be declared
along the same release path as in ZEPs (alpha, beta, GA, deprecated), as well as deprecation. We will centralize where
features are declared in the Zarf codebase, while also providing clear implementation guidance and examples for
maintainers to declare new features. CLI users will be offered multiple avenues of configuration: with CLI flags,
environment variables, and config files. Finally, the docs.zarf.dev page will document 1:1 each feature as they're
declared in code (as a stretch goal, this will be fully automated).

### User Stories (Optional)

- As a Zarf maintainer I want to add a new feature, get feedback, make revisions over time before fully releasing and 
maintaining it.
- As a Zarf maintainer I want to deprecate features with clarity on when they will be removed as well as trialing removal
by disabling the feature before it's fully removed.


- As a Zarf user I want to get involved in new experimental features, and opt-in to provide alpha feedback or beta testing.
- As a Zarf user I want to continue using Zarf without my critical workflows getting disrupted by experimental features.
- As a Zarf user I want deeper clarity if a feature that I use will be deprecated.

### Risks and Mitigations

#### Feature Sourcing
One critical pitfall to avoid with feature flags is creating _unexpected_ behavior, both for end users and maintainers.
Unclear state about which flags are enabled and not, and with and without defaults. State is managed plainly, a feature
can be set as enabled or disabled, and if no feature is set it is disabled by default. Clearly delineating Default vs. 
User-set (see "Global Feature Storage") makes the source of features clear.

#### UX and DX
User experience (UX) and developer experience (DX) and are both critical feature flag adoption. We've chosen to
implement features through multiple avenues of configuration to assist users in however they wish to run Zarf to aid in
UX. DX is made as expressive as possible through the API, with the ability to optimistically check for a flag with
`feature.Enabled(Name)` as well as query for the flags whether by source or all at once. User feedback remains key here.

#### Backporting Considerations
Once Features are shipped, we have the option to backport flags for existing features. To avoid breaking changes, these
backported features should be enabled by default. There isn't a particular urgency to add flags for an existing stable
(but still in alpha or beta) feature, but it is an option available to maintainers once Features is released.

## Design Details

### Types
Below is the proposed data model for Features, including the `Feature` type and supporting fields in the `feature` pkg.
```
pkg feature 
...
// Mode describes the two different ways that Features can be set. These are used as keys for All()'s return map.
type Mode string
var Default Mode = "default"
var User Mode = "user"

type Name string
type Description string
type Enabled bool
// Validate Since matches semver
type Since string
// Validate Until matches semver
type Until string

type Stage string
var (
  Alpha Stage = "alpha"
  Beta Stage = "beta"
  GA Stage = "ga"
  Deprecated Stage = "deprecated"
)

type Feature struct {
  // Name stores the name of the feature flag.
  Name
  // Description describes how the flag is used.
  Description
  // Enabled describes whether a feature is explicitly enabled or disabled. A feature that does not exist in any set
  // is considered disabled.
  Enabled
  // Since is the version a feature is first introduced in alpha stage.
  Since
  // Until is the version when a deprecated feature is fully removed. Historical versions included.
  Until
  // Stage describes what level of done-ness a feature is. TODO describe this better
  Stage
}
```

### API
Below are descriptions of each function in the feature API and their intended UX.

#### IsEnabled()
```
// Enabled allows users to optimistically check for a feature. Useful for control flow. Any user-enabled or disabled
// features take precedence over the default setting.
func IsEnabled(name Name) bool {
  ...
}
```

#### Set() and SetDefault()
```
// Set takes a slice of one or many flags, inserting the features onto user-configured features. If a feature name is
// provided that is already a part of the set, then Set will return an error. 
// TODO: Should we allow users to call this multiple times even if we don't allow them to overwrite features?
func Set(features []Feature) error {
  ...
}

// SetDefault takes a slice of one or many flags, inserting the features onto the default feature set. If
// a feature name is provided that is already a part of the set, then SetDefault will return an error. This function
// can only be called once.
func SetDefault(features []Feature) error {
  ...
}
```

#### Get(), GetDefault(), GetUser()
```
// Get takes a flag Name and returns the Feature struct. If the doesn't exist then it will error. It will check both the
// default set and the user set, and if a flag exists in both it will return the user data for it.
func Get(name Name) (Feature, error) {
  ...
}

// GetDefault takes a flag Name and returns the Feature struct from the default set.
func GetDefault(name Name) (Feature, error) {
  ...
}

// GetUser takes a flag Name and returns the Feature struct from the user set.
func GetUser(name Name) (Feature, error) {
  ...
}
```

#### All(), AllDefault(), and AllUser()
```
// All returns all flags from both Default and User.
func All() map[Mode]map[Name]Feature {
  ...
}

// AllDefault returns all features with from the Default set for this version of Zarf. 
func AllDefault() map[Name]Feature {
   ...
}

// AllUser returns all features that have been enabled by users.
func AllUser() map[Name]Feature {
   ...
}


// EXAMPLE
// Getting back an empty collection from these functions is not considered an error. The intended way to check is:
m := All()
if len(m[User]) == 0 {
  ...
}
```

### Global Feature Storage 
While CLI users are unaffected, global feature state has advantages for developer ease of use in the SDK. They can get
sensible defaults for loading up Zarf as a library, without worrying about any ceremony before calling the SDK.
(See "Ctx-based Flags that can be Updated at Runtime" under Abandoned Ideas)

Very similar to the global default logger implementation, the intent is to atomically store and load the state in a
private package var and provide a managed thread-safe API.

```
// Atoms wrapping the default and user-set collections of Features. These do not require mutexes because they will each
// be modified very few times, with the public API for each ensuring there only an empty set can be written to.
// Alternatively these could be sync.Maps but they're not written to enough where it matters.
// e.g. These atoms are write once, ready many (WORM).
var default = atomic.Value // map[Name]Feature
var user    = atomic.Value // map[Name]Feature
```


### Maintainers: Setting the Default Flags
```
package feature
...

func init() {
    features := [
        // Owner: @zarf-maintainers
        {
            Name: "foo",
            Description: "foo does the thing of course",
            Enabled: true,
            Since: "v0.60.0",
            Stage: GA
        },
        // Owner: @zarf-maintainers
        {
            Name: "bar",
            Description: "bar was honestly always a bit buggy, use baz instead" 
            Enabled: false,
            Since: "v0.52.0",
            Until: "v0.62.0",
            Stage: Deprecated
        },
    ]
    
    err := SetDefault(features)
    if err != nil {
        panic(err)
    }
}
```

### SDK: Setting User flags
```
feature.Set([
    {Name: "foo", Enabled: false},
    {Name: "bar", Enabled: true},
])

// My beautiful Zarf-enabled application
```

### CLI: Enabling and Disabling Flags
- Three approaches: `CLI Flags`, `Env Vars`, and `zarf-config.{yaml,toml}`, with precedence in that order.
- TODO CLI UX for Listing Flags: `zarf <COMMAND> -h` (K8s uses -h)
- CLI UX enable: `--feature-enable="Foo,Bar,Baz"`
- CLI UX disable: `--feature-disable="Fizz,Buzz,Qux"`
- Env UX enable: `ZARF_FEATURE_ENABLE="Foo,Bar,Baz" zarf p create ...`
- Env UX disable: `ZARF_FEATURE_DISABLE="Fizz,Buzz,Qux" zarf p create ...`
- Config UX enable: TODO, take a look at how these flow thru viper
- Config UX disable: TODO, take a look at how these flow thru viper

### Stretch Goal: Site Docs Automation
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

#### TODO Update the test plan after merge with more concrete details from the impl stage
- Create unit test file and tables for each feature flag function.
- Test errors in unit test cases as well
- TODO e2e testing. This may require a test flag for a mock feature. A/B test before and after flag is enabled? should
be pretty lightweight
- TODO Infra implications for more testing

### Graduation Criteria

#### Alpha
An alpha release of Features would contain the package in internal and would be mostly tested. Potentially, maintainers
could offer the configuration layer in a hidden and undocumented way. The main advantage to this is just getting to
exercise the implementation and its tests in a full release, rather than just existing on a branch. There is no real
utility for users to do so, though releasing it in an Alpha state, perhaps with something cute like `zarf say` printing
at the start of each run would give users the chance to trial the user experience and give us feedback.

#### Beta
A beta version of this feature would be feature-complete in a sense and fully tested. It should also be accompanied by
public config documentation on the CLI as well as a docs.zarf.dev article on all available features and the ways they
can be enabled or disabled. Trialing another alpha or beta feature with the feature API at this stage should be a
consideration, so we can get real user feedback at this point.

#### GA
Shipping to GA means proving production usage, offering as rigorous testing as possible, potentially with e2e tests,
as well as gathering extensive feedback from the community on its usage.

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
#### Create-time Feature State vs. Deploy-time Feature State
One significant open question which is what behavior to expect when creating a package with enabled or disabled
features, and whether those same feature should be set at deploy time. Much of this is the responsibility of maintainers
to manage within changes to create and deploy and the implementation of new features. However, _identifying_ at
deploy-time which features were enabled at create-time would be impossible without some addition to Zarf packages.

A potential solution is to serialize the state of all features (both enabled and disabled)

## Implementation History

Revision 1 of this doc is intended to include Summary, Motivation, Proposal and a first stab at an API implementation.
The reason why this is all done at once, is because we have prior art with feature flagging (TODO link to PR) and this
proposal is intended to generalize and provide long term support for this approach.

Revision 2 encapsulates feedback on the various design decision, and completes the API for the new impl. considerations.
Namely, storing multiple feature sets, offering API facilities to query for these separately, and doing using a
global API (not through ctx injection). Not solved in this revision is docs automation, which is considered a stretch
goal.

## Drawbacks

- Yet Another Process
- Increased complexity and config surface area for end-users. Effective docs makes this better but doesn't solve it.
- Features add some contributor friction and increased lead time on new features. However, this can also be framed as a
positive, because it encourages the design work necessary to see a feature through alpha, beta, and ga - exactly what
the ZEP process is intended to encourage.
- Increased contributor complexity when having to support backwards compatibility during new feature rollout and
deprecation strategy for the backwards compatible feature. Similar to contributor friction, this design work could be
seen as a _positive_ forcing function, not just a drawback.

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Value-based Merges of Feature Structs
During revision 1 of the proposal, we discussed whether the API should compare for highest version or latest feature
(e.g. beta takes precedence over alpha) for each field every time a flag is declared multiple times. Merge functions are
useful in many cases, but overkill in this situation. Default features are declared centrally and should be set at most
once. Version control can be used to update these declaratively, rather than relying on multiple declarations merging
during runtime. Any attempt to declare a flag more than once will error.

Users are afforded a similar UX, where duplicated flags return an error. However, there is one key and difference: if a
user sets a different state on a _default_ flag then the user-set state will take precedence. Feature flags wouldn't be
very useful for users if they couldn't enable or disabled. While this is more of an implementation detail, it is
important for UX in cases where features are _disabled_ by default and a user wants
to reenable it. Take for example a deprecated feature. The opposite is also true, where a feature may have graduated to
beta and is enabled by default -- users will be able to explicitly _disable_ the flag via any of the available config
paths (CLI flags, environment variables, or `zarf-config.{yaml,toml}`)

### Ctx-based Flags that can be Updated at Runtime
The original version of this proposal, and the original implementation it was based off of, used an implementation and API
for feature flags based on the ctx object. This has certain advantages like forgoing global feature state in favor of
dependency injection. This was relatively lightweight and worked well for flagging out the logging overhaul. What we
found however, is that for all of its advantages in avoiding global state for SDK users, it was neither intuitive or
discoverable UX for those same SDK users. In order to get defaults, SDK users calling into a random part of the API
would have to know that the feature library exists and set the defaults themselves. Errors could be provided downstream
if an expected flag did not exist, but adding this ceremony just to avoid managed global state is a poor tradeoff in API
design and developer experience.

### Using an Off-the-Shelf Library or Pre-existing Solution

It's fair to say that Feature Flags are a known space, and even a solved problem in many ways. Many libraries exist,
whether fully-fledged multi-language specs and APIs like OpenFeature or mature project-specific implementations like
K8s' Feature Gates. This raises a clear question of: why roll your own implementation then?

In both of the mentioned cases, these implementations bring on complexity outside the scope of what the Zarf project
needs. OpenFeature for example provides abstractions for different feature providers in a common API and spec, and this
is particularly useful for server-based web applications and services in a connected environment. K8s' feature gates
provide runtime-dynamic configuration and multiple feature flags under specific gates.

In an attempt to get exactly what we need and nothing more, we've taken significant inspiration from K8s in how to
_model_ the fields on our Feature type. The release stages very closely mirror the ZEP and KEP processes, and the version
fields `Since` and `Until` directly mirror K8s features. The main way we diverge is simplifying both API and the storage
for a smaller project like Zarf.

Should our needs change in the future, then we have clear boundaries in the config layer for CLI users, and a
straightforward store/load API with clear guarantees for users. Any API migration could maintain backwards compatibility
with the existing API relatively easily while migrating over to a new implementation.

## Future Work

### Docs Automation
In the current iteration, documentation for new features will be created and maintained by hand. This is currently how
the Kubernetes does so as well, though automation would be ideal. The flip side is that once a feature is declared and
documented, the overall change area should be relatively low throughout its lifecycle. As a note for future
implementation, centralizing Features declarations and using a consistent doccomment format should make site automation
relatively simple, keeping the website updated with `make docs-and-schema` should help keep the site one to one with
what is available in code.

## (TODO) Infrastructure Needed (Optional)

<!--
Use this section if you need things from the project. Examples include a new repo,
cloud infrastructure for testing or GitHub details. Listing these here
allows the process to get these resources to be started right away.
-->

TODO Implementation considerations for Test plan. Esp. additional e2e tests.