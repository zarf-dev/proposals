<!--
**Note:** When your ZEP is complete, all of these comment blocks should be removed.

To get started with this template:

- [X] **Create an issue in zarf-dev/proposals.**
  When creating a proposal issue, complete all fields in that template. One of
  the fields asks for a link to the ZEP, which you can leave blank until the ZEP
  is filed. Then, go back and add the link.
- [X] **Make a copy of this template directory.**
  Name it `NNNN-short-descriptive-title`, where `NNNN` is the issue number
  (with no leading zeroes).
- [X] **Fill out as much of the zep.yaml file as you can.**
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

# ZEP-0048: Schema update process

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

New schema versions of the Zarf package config presents the opportunity to improve the experience for package creators and provide a clear timeline for removing deprecated fields. However, handling multiple schema versions in Zarf presents unique challenges as packages can be created and deployed on different versions of Zarf.
Zarf should provide users with clear expectations around a schema's lifetime, and provide a simple path for users to upgrade their package definitions. Zarf maintainers should have a standardized approach to adopting a new schema in the codebase. 


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

There are several open issues requesting enhancements to the schema, but before Zarf introduces a new schema, there must be a plan to handle schema upgrades. The general theme of these changes is to make the ZarfPackageConfig schema more intuitive to use. 
- [Refactor charts definition in zarf.yaml #2245](https://github.com/zarf-dev/zarf/issues/2245)
- [Breaking Change: make components required by default #2059](https://github.com/zarf-dev/zarf/issues/2059)
- [Use kstatus as the engine behind zarf tools wait-for and .wait.cluster #4077](https://github.com/zarf-dev/zarf/issues/4077)

ZEP [0051-v1beta1-schema](https://github.com/zarf-dev/proposals/pull/52) provides the specifics for what will change in the next schema version of Zarf

### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

- Provide clear guidelines for how Zarf package commands should behave when handling new or old schema versions.
- Design a strategy for updating the codebase when a new schema is introduced. 
- Introduce a command for users to upgrade the schema version of their Zarf package config.

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

During Zarf's lifetime it will introduce, deprecate, and drop support for ZarfPackageConfig API versions. Once a version is deprecated users will still be able to perform all package operations such as create, publish, and deploy, but will receive warnings that they should upgrade. Zarf will drop support for an API version one year after it is deprecated. Once an API version is no longer supported, Zarf will error if a user tries to perform any zarf package operations with that API version. 

The zarf.yaml in a built package will include the package definition for every supported API version. When printing the package definition to the user, for instance, with the command `zarf package inspect definition` the API version will be the version that the package was created with. A new field `.build.apiVersion` will be added to all schemas to track which API version was used at build time. 

A new command `zarf dev convert` will be introduced to allow users to convert from one API version to another. The command will default to converting to the latest API version. It will create a new file zarf-*apiversion*.yaml with the converted package definition. It will accept an optional API version, so a user could run `zarf dev convert v1beta1` and they will receive a file called zarf-v1beta1.yaml. Convert will not allow changing from a newer version to an older version so running `zarf dev convert v1alpha1` on a `v1beta1` schema will error. 

Deprecated fields will not be removed until a future API version. Newer API versions will track fields removed from one API version for lossless conversions, but will not allow creation with removed fields. For instance, Data Injections will be removed in v1beta1. Users will still be able to deploy existing v1alpha1 packages until v1alpha1 support is removed, but they will not be able to create a new v1beta1 package with Data Injections. 

Functions in Zarf will always accept the latest API version. This will result in several breaking changes in the SDK, about 30 public functions accept an object from the v1alpha1 package as of late 2025. This breaking change should be acceptable since common flows involve loading a package through public functions such as `load.PackageDefinition()` or `packager.LoadPackage()`. 
<!-- ^func (\([^)]+\) )?[A-Z][a-zA-Z0-9_]*\([^)]*\bv1alpha1\. with exclude **/internal/**  to figure out the amount of v1alpha1 uses in public functions -->

### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1

As a package deployer, I want to use the latest version of Zarf, but I still want to pull and deploy packages that were built using the v1alpha1 schema. I run `zarf package deploy oci://<package>` and it simply works.

#### Story 2

As a package creator, I want to create packages using the newer API version, however I still want my package to be deployable on older versions of Zarf that have not yet introduced this API version. I run `zarf package inspect definition <my-package>` and ensure that `.build.deployRequirements.version` is empty or less than my expected version.
<!-- TODO, the field name will likely change -->

#### Story 3

As a package creator, I want to update my package definition to the latest schema so I run `zarf dev convert` with a zarf.yaml in my current directory and it automatically updates my package to the newest version of the Zarf package config.  

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

### Conversions

Zarf will need to handle two use cases for conversions. The first is `zarf dev convert`, providing a simple way for users to convert their zarf.yaml files from one schema version to the next. The second is internal convert functions which will allow for backwards compatible lossless conversions. This will provide a path for existing packages to call packager functions when they change to use the v1beta1 objects.

There will be an internal `ZarfPackage` object used solely for conversions. Rather than having functions which convert v1alpha1 to v1beta1, functions will instead convert v1alpha1 to the internal Zarf package type then convert the internal Zarf package type to v1beta1. This means Zarf only needs N conversion functions (N API versions) rather than N² conversions between every pair of versions. 

#### Converting 1:1 Replacements
If a field is renamed with a 1:1 replacement, then Zarf will automatically convert the field to its replacement. For example, if a field called `noWait` was changed to `wait` then the value of the field will flip during conversion. 

#### Converting Removed Fields

The logic for converting removed fields will differ depending on the use case. When `zarf dev convert` is used a field may still be converted if it has a near 1:1 replacement. For example, in the v1beta1 schema a new required field `.apiVersion` will be added to `.cluster.wait`. The convert command would add a key for `apiVersion` in the new zarf package config, but it will be left empty. Since this will be a required field in the v1beta1 schema the package will fail on create if the user doesn't set `apiVersion`. The conversion gets users most of the way there, and the fix should be simple.

If there is not a replacement for a field then  `zarf dev convert` will error and recommend an alternative approach to replacing the field. 

During an internal conversion, for example, converting an an already built v1alpha1 package during `zarf package deploy` Zarf must always convert the package to the latest schema version without any data loss. To achieve this, removed fields from an earlier schema version will be added to newer objects as private fields. Keeping the fields private, stops users from setting them outside of conversions and keeps them out of the new schema.

In the below example, `dataInjections` is set as a private field on Zarf components so that it can be set during conversions between v1alpha1 and v1beta1. It will not be included in the v1beta1 schema, and since Zarf validates against the schema on create users will be unable to create v1beta1 packages with `dataInjections`. Likewise, since `dataInjections` is a private field, SDK users will not be able to set it directly. 

```go
type ZarfComponent struct {
	Name string `json:"name"`
  ...
	// data injections are kept as a backwards compatibility shim and can only be set when converting from v1alpha1
	dataInjections []v1alpha1.ZarfDataInjection
  ...
}
```

### JSON Schema

Zarf currently publishes a JSON schema, see the [current version](https://raw.githubusercontent.com/zarf-dev/zarf/refs/heads/main/zarf.schema.json). Users often use editor integrations to have built-in schema validation for zarf.yaml files. This strategy is [referenced in the docs](https://docs.zarf.dev/ref/dev/#vscode). The Zarf schema is also included in the [schemastore](https://github.com/SchemaStore/schemastore/blob/ae724e07880d0b7f8458f17655003b3673d3b773/src/schemas/json/zarf.json) repository.

Zarf will use the if/then/else features of the json schema to conditionally apply a schema based on the `apiVersion`. If the `apiVersion` is `v1alpha1` then the schema will evaluate the zarf.yaml file according to the v1alpha1 schema. If the `apiVersion` is v1beta1 then the zarf.yaml will be evaluated according to the v1beta1 schema. 

### Updating Packages

Once the latest schema is introduced, the built zarf.yaml file will contain the package definition for itself, as well as all older apiVersion that is still supported. For example, the built zarf.yaml in a v1beta1 package will include the v1beta1 package config and v1alpha1 package config. The built zarf.yaml for a v1alpha1 package will only include the v1alpha1 package. This is done because earlier API versions will always be able to convert to newer schema versions without data loss, but newer API versions may include fields that are not represented in older API versions. In these cases, Zarf will still attempt to keep the package backwards compatible, but will make set a minimum required version according to [minimum-version-requirements](#minimum-version-requirements). 

Package definitions will be separated by the standard YAML `---`. Currently, Zarf only checks the first yaml object in the zarf.yaml file. To maintain backwards compatibility newer packages must place the v1alpha1 definition at the beginning of the zarf.yaml file. Future versions of Zarf will check the API version of each package definition and select the latest version that it understands.
<!-- TODO I have to check if it's true that Zarf only checks the first definition in the file, or it dynamically checks whichever one it can read -->

A new field on all future schemas called `.build.apiVersion` will be introduced to track which apiVersion was used at build time. This field will be used to determine which version of the package definition will be printed to the user during `zarf package inspect definition` and the interactive prompts of `zarf package deploy|remove`. 

### Minimum Version Requirements

Zarf will introduce a new field `build.deployRequirements` which will be automatically populated on create. If there is a new field in any schema that changes the deploy process, then the package should not be deployable on versions of Zarf without that feature. This field will be checked on deploy to prevent users from deploying packages that may break. This will not work on versions of Zarf where this field is not yet implemented. The field will look like below:
```go
type DeployRequirements struct {
	// the minimum version of the Zarf CLI that can deploy the package
	Version string 
	// Reasons for why the package can't be deployed
	// EX: "values was not introduced until v0.64.0, package structure changed in v0.65.0"
	Reasons []string
}
```
<!-- FIXME, the actual design will look different -->

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

- The new command `zarf dev convert` will have e2e tests.
- There will be e2e tests that will build a v1beta1 package and verify that a version of Zarf prior to v1beta1 being introduced can still deploy that package.

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

`zarf dev convert` will be released alongside the v1beta1 schema. Given that this is a simple command with low amounts of risk, it will be released as GA. 

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

This ZEP is an upgrade / downgrade strategy

### Version Skew Strategy

<!--
If applicable, how will the component handle version skew with other
components? What are the guarantees? Make sure this is in the test plan.

Consider the following in developing a version skew strategy for this
proposal:
- Does this proposal involve coordinating behavior between components?
  - (i.e. the Zarf Agent and CLI? The init package and the CLI?)
-->

The Zarf agent will not be impacted as it does not interact with the package config.

Zarf package definition that is persisted to cluster will change depending on `.build.apiVersion`. The rest of the data that is persisted to the cluster will remain the same. 

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

- 2025-10-18: Proposal submitted

## Drawbacks

<!--
Why should this ZEP _not_ be implemented?
-->

### SDK breaking changes
There will be breaking changes to SDK functions every time a new API version is introduced. This could be frustrating for users which have complex integrations with the SDK, However, common user flows should generally be unchanged. For example this flow will work regardless of the API version: 

```go
	pkgLayout, err := packager.LoadPackage(ctx, packageSource, loadOpt)
	if err != nil {
		return fmt.Errorf("unable to load package: %w", err)
	}
	_, err = packager.PublishPackage(ctx, pkgLayout, dstRef, packager.PublishPackageOptions{})
```

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Public Facing Internal Type

Rather than updating functions to accept a newer version of the schema, Zarf could have a publicly facing internal type that has every field from every version and use that throughout the SDK. The upside of this approach is that we would avoid breaking changes throughout the lifetime of the SDK. The downside is that it would make it easy for anyone using the SDK to set deprecated fields. It would also make it confusing and unclear which fields attach to which versions. 

### Map representation of Removed Fields

One option for storing removed fields on newer schemas is to use annotations. Kubernetes takes this approach. This would avoid the need of a custom YAML marshaler. The downside is that annotations could easily get confusing and hard to read. Assuming a list of objects such as `variables` is deprecated, then we need to maintain an long string representation of YAML.
The reason Kubernetes takes this approach is because their data must make lossless round trips. Their objects might be written as v1beta1, stored as v1alpha1, then upgraded back to v1beta1, and they cannot lose any data. There is no place to store the information on the v1alpha1 object besides annotations. Zarf is going to write all active API versions to the zarf.yaml file so there is no chance of data loss. 

Another option is introducing a new field `Deprecated map[string]string` to store removed fields from previous schema versions, however this still leaves the complexity of unfurling complex objects from a string. 

### Custom YAML marshalers for Removed Fields

Originally, [Removed Fields](#converting-removed-fields) proposed custom marshalers to track the private fields for backwards compatibility such as `dataInjections`. However, going through each case we see that custom YAML marshalers are not needed:
- a v1alpha1 package is created after function signatures are changed to accept v1beta1 objects. 
  - This package is read as v1alpha1, then converted to v1beta1 for `packager.create`, then written to the `zarf.yaml` in the package tar as v1alpha1. The package never needs to be written as v1beta1.
- a v1alpha1 package is created before v1beta1 is introduced. The package is deployed after function signatures are changed to accept v1beta1 objects
  - The package is read as v1alpha1, then converted to v1beta1 during deploy, then persisted to the cluster as v1beta1, it is never written to YAML as v1beta1.
- A v1beta1 package is created
  - The package is written as both v1alpha1 and v1beta1, but removed fields such as `dataInjections` cannot be set. 