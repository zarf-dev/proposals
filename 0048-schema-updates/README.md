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

# ZEP-0048: Schema updates

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

New schema versions in Zarf present the opportunity to improve the experience for package creators and provide a clear timeline for removing deprecated fields. However, handling multiple schema versions in Zarf presents unique challenges as packages can be created and deployed on different versions of Zarf. Zarf should provide users with clear expectations for how long a schema can be used, and provide a simple path for users to upgrade their schema. Zarf maintainers should have a standardized approach to adopting a new schema in the codebase. 


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

There are several issues asking for enhancements to the schema, before Zarf introduces a new schema, we should have a plan for how to handle schema upgrades. The general theme of these changes is to make the ZarfPackageConfig schema more intuitive to use. 
- [Refactor charts definition in zarf.yaml #2245](https://github.com/zarf-dev/zarf/issues/2245)
- [Breaking Change: make components required by default #2059](https://github.com/zarf-dev/zarf/issues/2059)
- [Use kstatus as the engine behind zarf tools wait-for and .wait.cluster #4077](https://github.com/zarf-dev/zarf/issues/4077)

There exists ADR [0026-schema.md](https://github.com/zarf-dev/zarf/blob/main/adr/0026-schema.md) in the Zarf repo which discusses the changes to be made in a v1 schema. The schema was not yet implemented and likely the final version of the v1 schema will differ, however this goes document does correctly reflect many changes that are planned to improve upon the schema. There will instead be an additional ZEP to discuss changes to the schema. 
<!-- TODO add ZEP once available  -->

### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

- Provide clear guidelines for how Zarf package commands should behave when dealing with new schemas 
- Provide a maintainable way to keep the codebase updated when a new schema is introduced. 
- Introduce a command for users to upgrade their schema version.

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->

- Define the next API version of the ZarfPackageConfig

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

During Zarf's lifetime it will introduce, deprecate, and remove ZarfPackageConfig API versions. Once a version is deprecated users will still be able to preform all package operations such as create, publish, and deploy, but will receive warnings that they should upgrade. Once an API version is removed Zarf will error if a user tries to preform any package operations with that package. 

A new command `zarf dev convert` will be introduced to allow users to convert from one API version to another. By default the command will take the current version and migrate it to the latest schema. It will accept an optional API version, so a user could run `zarf dev convert v1beta1`. Convert will not allow changing from a newer version to an older version so running `zarf dev convert v1alpha1` on a `v1beta1` schema will error. 
<!-- FIXME, add cobra command docs -->

Deprecated fields on an API version, must not be removed until the next API version. 

Functions in Zarf will always accept the latest API version. This will result in several breaking changes in the SDK, about ~30 public functions accept an object from the v1alpha1 package as of late 2025. This breaking change should be acceptable since common flows usually involve loading a package through another functions such as `load.PackageDefinition` or `packager.LoadPackage()`. 
<!-- ^func (\([^)]+\) )?[A-Z][a-zA-Z0-9_]*\([^)]*\bv1alpha1\. with exclude **/internal/**  to figure out the amount of v1alpha1 uses in public functions -->

### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1

As a a package deployer, I want to be on the latest version, but I still want to deploy packages that were built using the v1alpha1 schema. I run `zarf package deploy <my-package.tar>` and it simply works. Fields such as data injections, which will be removed in newer versions of the schema will still work as intended when deployed with newer versions of Zarf. 

#### Story 2

As a package creator, I want to create packages using the newer API version, however I still want my package to be deployable on older versions of Zarf that have not yet introduced this API version. 

#### Story 3

As a package creator, I want to update my package definition to the latest schema so I run `zarf dev convert` with a zarf.yaml in my current directory and it automatically updates my package.  

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

### Converting between API versions

How Zarf will determine converting fields between schemas. 
- When two fields have the same name they will simply be copied over
- When a field is renamed with a direct replace, then Zarf will automatically convert the field to it's replacement doing any conversions necessary
  - For example, if a field called `noWait` was changed to `wait` then the value of the field will flip during conversion. 
- If a field is removed without a 1:1 replacement then convert behavior will differ depending on the use case
  - If a conversion is internal, (I.E. converting a v1alpha1 package to v1beta1 so that it can be used in packager functions) then fields without a 1:1 replacement must be added to the package as annotations. 
  <!-- FIXME is there a character limit for annotations? This could become problematic if a package is pushed to OCI with too large of an annotation. For example, a package with a lot of variables -->
  <!-- FIXME One option could be having private fields, however when those are used you lose the ability marshal / unmarshal to yaml. There might be a way around this with custom marshalers   -->  
  - If the conversion is user facing (I.E. `zarf dev convert`), then the conversion will fail, and the user will be asked to remove the field. The error message should suggest an alternative, and may link to documentation.
    - Alternatively, there may be situations where a conversion mostly works, but isn't 1:1. For example, the v1beta1 schema will add the field `.apiVersion` to `.cluster.wait`. The convert function would add a key for this field in the new zarf.yaml, but will leave it empty. Since this will be a required field in the new schema the package will fail on create if this field is not filled out. These decisions will be situational depending on the fields, the goal is to have the best user experience for `zarf dev convert`.  
  - If a field is introduced in the v1beta1 package and does not have a direct replacement, then the field will not exist in the v1alpha1 package. If this has potential to break a deployment, then the deploy should error.


When a field is renamed with a direct replacement then Zarf will automatically convert the field to it's replacement. 

When a field is removed without a replacement then during conversion Zarf will add this field in the annotations of the package. This way the package can still be deployed. 

### Updating packages

Once the latest schema is introduced packages will contain multiple files representing this package config. Regardless of the APIVersion of the ZarfPackageConfig created assuming the two existing schemas are `v1alpha1` (current default) and `v1beta1` then Zarf will create both a zarf.yaml file and zarf-v1beta1.yaml file within the package. This way older versions of Zarf can still deploy packages created with the v1beta1 schema. Newer versions of Zarf can use the zarf-v1beta1.yaml file when available and fallback to older versions.

### Minimum Zarf version requirements

Zarf should introduce a minimum version requirement for the package to be deployed. If there is a new field in the v1beta1 schema that changes how the deploy process is done, then the package should not be deployable on versions of Zarf without that feature. A new field `build.MinimumDeployVersion` will be introduced in version X. Once this field is introduced, Zarf will check against this field to ensure it can deploy packages. 


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

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Internal type throughout SDK

Rather than updating functions to accept a newer version of the schema, we could have a publicly facing internal type that has every field from every version and use that throughout the SDK. The upside of this approach is that we would avoid breaking changes throughout the lifetime of the SDK. The downside is that it would make it easy for anyone using the SDK to set deprecated fields.
<!-- FIXME I need to add more detail -->

### Storing removed fields on newer schemas

<!-- FIXME, we have to decide to either use annotations or private fields -->