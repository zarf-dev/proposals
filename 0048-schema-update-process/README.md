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

New schema versions of the Zarf package config present the opportunity to improve the experience for package creators and provide a clear timeline for removing deprecated fields. However, handling multiple schema versions in Zarf presents unique challenges as packages can be created and deployed on different versions of Zarf.
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

ZEP [0051-v1beta1-schema](https://github.com/zarf-dev/proposals/pull/52) provides the specifics for what will change in the next schema version of Zarf.

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

During Zarf's lifetime, it will introduce, deprecate, and drop support for ZarfPackageConfig API versions. Once a version is deprecated, users will still be able to perform all package operations such as create, publish, and deploy, but will receive warnings that they should upgrade. Zarf will drop support for an API version one year after it is deprecated. Once an API version is no longer supported, Zarf will error if a user tries to perform any zarf package operations with that API version such as `zarf package create`, `zarf package publish`, or `zarf package deploy`. Even after an API version drops support there will be an exception for packages already deployed to the cluster. Currently, Zarf interacts with deployed packages through the commands `zarf package inspect definition` and `zarf package remove` and these commands will still work for an additional year after support is dropped for that API version.

The zarf.yaml in a built package will include the package definition for every supported API version. When printing the package definition to the user, for instance, with the command `zarf package inspect definition` the API version will be the version that the package was created with. A new field `.build.apiVersion` will be added to all schemas to track which API version was used at build time. 

A new command `zarf dev upgrade-schema` will be introduced to allow users to convert from one API version to another. The command will default to converting to the latest API version. It will create a new file `zarf-<apiversion>.yaml` with the converted package definition. It will accept a path to a directory containing a zarf.yaml file and an optional API version. For instance, a user could run `zarf dev upgrade-schema . v1beta1` and they will receive a file called `zarf-v1beta1.yaml`. Convert will not allow changing from a newer version to an older version, so running `zarf dev upgrade-schema . v1alpha1` on a `v1beta1` schema will error. This command will only accept with a local zarf.yaml file, and will not accept created packages, published packages, or deployed packages. 

API versions of the package schema will not necessarily coincide with releases of the Zarf CLI. One caveat is that Zarf will likely not release an official v1.0.0 version until there is a v1 version of the schema, however it could be the case that a v2 package schema is released while the CLI version is still v1.0.0 and vice versa. 

Once an API version is released, fields will not be removed from it, and there will be no new required fields.

Functions in Zarf will always accept the latest API version. This will result in several breaking changes in the SDK; about 30 public functions accept an object from the v1alpha1 package as of late 2025. Many SDK users should see only small changes to their workflows since common flows involve loading a package through functions such as `load.PackageDefinition()` or `packager.LoadPackage()` rather than defining specific API versions. See [SDK Breaking changes](#sdk-breaking-changes) for more details.
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

As a package creator, I want to create and publish packages using the newer API version; however, I still want my package to be deployable on older versions of Zarf that have not yet introduced this API version. I run `zarf package inspect definition <my-package>` and ensure that `.build.VersionRequirements.Version` is empty or less than my expected version.

#### Story 3

As a package creator, I want to update my package definition to the v1beta1 schema, so I run `zarf dev upgrade-schema` with a zarf.yaml in my current directory and it creates the converted package definition in a file called zarf-v1beta1.yaml.

#### Story 4

As a Zarf maintainer, I want to introduce a new API version so that I can deprecate fields, add new required fields, and rename fields in the current package schema. I want the process to do this to be straightforward. I want earlier versions of Zarf

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

### SDK breaking changes
There will be breaking changes to SDK functions every time a new API version is introduced. This could be frustrating for users who have complex integrations with the SDK. However, common user flows should generally be unchanged. For example, this flow will work regardless of the API version: 

```go
	pkgLayout, err := packager.LoadPackage(ctx, packageSource, loadOpt)
	if err != nil {
		return fmt.Errorf("unable to load package: %w", err)
	}
	_, err = packager.PublishPackage(ctx, pkgLayout, dstRef, packager.PublishPackageOptions{})
```

Additionally, there are several functions that accept a v1alpha1.ZarfPackage which are only applicable to built Zarf packages. These functions could instead accept a package layout limiting the amount of breaking changes in Zarf. Still, since most SDK users will call these functions with a package loaded from yaml, tar, or from the cluster rather than defining a ZarfPackage object, this shouldn't be a major issue for SDK users. 

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

### New Package Compatibility

#### Built packages
Once the latest schema is introduced, the built zarf.yaml file will contain the package definition for itself, as well as all older API versions that are still supported. For example, the built zarf.yaml in a v1beta1 package will include the v1beta1 package config and v1alpha1 package config. The built zarf.yaml for a v1alpha1 package will only include the v1alpha1 package. This is done because older API versions will always be able to convert to newer API versions without data loss, but newer API versions may include fields that are not represented in older API versions.

A new API version may coincide with packages being incompatible with earlier versions of Zarf, but the logic for determining compatibility will be decoupled from the API version. Zarf will introduce a new field `build.VersionRequirements` which will be automatically populated on create, and will error on deploy or remove if the user's version is older than the required version. See [#4256](https://github.com/zarf-dev/zarf/issues/4256)

Package definitions will be separated by the standard YAML `---`. Currently, Zarf only checks the first yaml object in the zarf.yaml file. To maintain backwards compatibility, API versions will always be placed in ascending order beginning with the v1alpha1 definition. Future versions of Zarf will check the API version of each package definition and select the latest version that it understands. This process will be implemented before any new API versions are released. If Zarf sees a version that it does not understand, Zarf will log to the user that there is a new API version available that the user should consider updating to. 

A new field on all future schemas called `.build.apiVersion` will be introduced to track which apiVersion was used at build time. This field will be used to determine which version of the package definition will be printed to the user during `zarf package inspect definition` and the interactive prompts of `zarf package deploy|remove`. 

#### Deployed packages

The current deployed package struct is seen below. The `DeployPackage` object is persisted to the cluster during `zarf package deploy` as a Kubernetes secret. The Data field is a json representation of the package.

```go
type DeployedPackage struct {
	Name               string               `json:"name"`
	Data               v1alpha1.ZarfPackage `json:"data"`
	CLIVersion         string               `json:"cliVersion"`
	Generation         int                  `json:"generation"`
	DeployedComponents []DeployedComponent  `json:"deployedComponents"`
	ConnectStrings     ConnectStrings       `json:"connectStrings,omitempty"`
	NamespaceOverride string                `json:"namespaceOverride,omitempty"`
}
```

In the future, when Zarf stores this secret, it will store the version the package was created with as well as all earlier API versions, mimicking the strategy used in built packages. This will enable older versions of Zarf that don't have the latest API version to be able to read newer packages. Additionally, when a user runs `zarf package inspect definition` on a cluster sourced deployed package, they will receive a printed yaml of the API version they built the package with. To track multiple versions, a new field named `PackageData` of type `map[string]json.RawMessage` will be introduced on the struct. The original Data object, will stay on the object for backwards compatibility, until the v1alpha1 package is no longer supported.

```go
type DeployedPackage struct {
	Name               string                     `json:"name"`
  // Data is kept for backwards compatibility, once support for reading v1alpha1 packages from the cluster is removed, this field will be deleted.
	Data               v1alpha1.ZarfPackage       `json:"data"`
  PackageData        map[string]json.RawMessage `json:"packageData"`
	CLIVersion         string                     `json:"cliVersion"`
	Generation         int                        `json:"generation"`
	DeployedComponents []DeployedComponent        `json:"deployedComponents"`
	ConnectStrings     ConnectStrings             `json:"connectStrings,omitempty"`
	NamespaceOverride string `json:"namespaceOverride,omitempty"`
}
```

### Conversions

Zarf will need to handle two use cases for conversions. The first is library convert functions. These functions will provide a path for existing packages to call packager functions after they change accept v1beta1 objects. The second is `zarf dev upgrade-schema` which will provide a simple way for users to convert their zarf.yaml files from one schema version to the next.

#### Type API changes

The [api](https://github.com/zarf-dev/zarf/tree/main/src/api) package will be structured as below:

```bash
├── internal
│   └── types
│     └── package.go
│   └── v1alpha1
│     └── convert.go
│     └── validate.go   
│   └── v1beta1
│     └── convert.go
│     └── validate.go   
├── v1alpha1
│   ├── convert.go
│   ├── package.go
│   ├── ...
├── v1beta1
│   ├── convert.go
│   ├── package.go
│   ├── ...
├── convert
│   ├── convert.go
```

The internal/types package will contain a superset of Zarf fields to enable conversions between API versions. Rather than having functions which convert v1alpha1 to v1beta1, functions will instead convert v1alpha1 to the generic Zarf package type then convert the generic Zarf package type to v1beta1. This means Zarf only needs N conversion functions (N API versions) rather than N² conversions between every pair of versions. 

The internal/types package will not be exposed by the SDK. Instead the convert package will expose functions such as `func V1Alpha1PkgToV1Beta1(in v1alpha1.ZarfPackage) v1beta1.ZarfPackage`. These functions will call on the internal API packages, for instance, `internalv1alpha1.ConvertToGeneric(in v1alpha.ZarfPackage) types.ZarfPackage` and `internalv1beta1.ConvertFromGeneric(in types.ZarfPackage) v1beta1.ZarfPackage`. This will give users a clean interface for SDK users while avoiding exposing the internal types. These conversion functions will be manually written as opposed to [automatically generating conversion functions](#automatically-generating-conversion-functions). 

The public API versioned packages will expose a method on the ZarfPackage object called `Validate()`. These methods will call the internal API versioned packages where the validation logic will live. The validation logic currently in src/pkg/lint/validate.go will be moved to internal/v1alpha1. This structure will be implemented before v1beta1 is released, and added to with each new API version. 

##### Converting 1:1 Replacements
If a field is renamed with a 1:1 replacement, then Zarf will automatically convert the field to its replacement. For example, if a field called `noWait` was changed to `wait` then the value of the field will flip during conversion

##### Converting Removed Fields

When Zarf internally converts an older schema version to a newer schema version (for example, while deploying a v1alpha1 package), it must always convert to the latest schema version without data loss. To achieve this, fields that were removed from earlier schema versions are preserved as private fields in newer objects. These private fields are kept out of the new schema. These fields will have private fields have getters and setters so that they can be set. After the API Version that these fields originated from is unsupported, these fields will be deleted. 

A concrete example of how this will be implemented is seen with `dataInjections` from v1alpha1 to v1beta1. Below is a code snippet for the v1beta1 schema object. `dataInjections` is set as a private field on the v1beta1 Zarf component so that it can be set during conversions between v1alpha1 and v1beta1. While it is an object on the struct, because it's a private field, `dataInjections` will not be included in the v1beta1 schema, and since Zarf validates against the schema on create, users will be unable to create v1beta1 packages with `dataInjections` set.

```go
type ZarfComponent struct {
	Name string `json:"name"`
  ...
	// data injections are kept as a backwards compatibility shim and should only be set when converting from v1alpha1
	dataInjections []v1alpha1.ZarfDataInjection
  ...
}
// DataInjections should only be set when converting from a v1alpha1 package. After v1alpha1 packages is not supported this will be removed. 
func (c ZarfComponent) SetDataInjections(di []v1alpha1.ZarfDataInjection)
func (c ZarfComponent) GetDataInjections() []v1alpha1.ZarfDataInjection
```

#### zarf dev upgrade-schema

`zarf dev upgrade-schema` will call the library conversion functions, however it will have additional checks. If a user's package contains a removed field that does not have a 1:1 replacement, then the command will error. The error message will recommend an alternative approach to replacing the field. 

The usage docs for `zarf dev upgrade-schema` will look like below:

```bash
upgrades the existing zarf package config to the given API version. Defaults to latest API version if not given. 

Usage:
  zarf dev upgrade-schema [ DIRECTORY ] [ API Version ] [flags]
```

### JSON Schema

Zarf publishes a JSON schema, see the [current version](https://raw.githubusercontent.com/zarf-dev/zarf/refs/heads/main/zarf.schema.json). Users often use editor integrations to have built-in schema validation for zarf.yaml files. This strategy is [referenced in the docs](https://docs.zarf.dev/ref/dev/#vscode). The Zarf schema is also included in the [schemastore](https://github.com/SchemaStore/schemastore/blob/ae724e07880d0b7f8458f17655003b3673d3b773/src/schemas/json/zarf.json) repository.

Zarf will use the if/then/else features of the json schema to conditionally apply a schema based on the `apiVersion`. If the `apiVersion` is `v1alpha1` then the schema will evaluate the zarf.yaml file according to the v1alpha1 schema. If the `apiVersion` is v1beta1 then the zarf.yaml will be evaluated according to the v1beta1 schema. It's useful to have a single schema file, so that users' text editors handle different API versions without file specific annotations. Zarf will still create and utilize individual version schemas. 


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

The new command `zarf dev upgrade-schema` should have unit tests in the src/cmd package.

Every child command under `zarf package -h` should have coverage with both the previous API version and the new API version. 

There should be tests to ensure fields removed from newer API versions still work when defined on older API versions. 

There should be tests for version skew. Newer API versions should be deployable on versions of Zarf before they are introduced. Likewise, there should be tests to ensure that cluster commands for these packages such as inspect and remove are successful. 

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

`zarf dev upgrade-schema` will be released alongside the v1beta1 schema. Given that this is a simple command with low amounts of risk, it will not go through a phased maturity process (i.e., alpha/beta/stable). 

When a new schema is introduced, creating a package using the newer version will be behind a feature flag. After the feature flag is enabled by default, there will be no more breaking changes to the schema. There will not a phased maturity process for the API version.  

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

This ZEP is an upgrade/downgrade strategy.

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

New packages will be compatible with older versions of Zarf. This is detailed in the [New Package Compatibility](#new-package-compatibility) section. 

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
- 2025-12-08: Updated proposal to focus more on the process Zarf maintainer should follow to ensure that new API versions can be introduced

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

### Public Facing Internal Type

Rather than updating functions to accept a newer version of the schema, Zarf could have a publicly facing internal type that has every field from every version and use that throughout the SDK. The upside of this approach is that we would avoid breaking changes throughout the lifetime of the SDK. The downside is that it would make it easy for anyone using the SDK to set deprecated fields. It would also make it confusing and unclear which fields attach to which versions. 

### Internal Type wrapped by Public Versioned Functions

Another way an internal type could be used would be to introduce public functions such as `packager.RemoveV1alpha1()` and `packager.RemoveV1beta1()`. These functions would then call a private `packager.remove()` function that accepts the internal type. This way SDK users don't have to deal with the internal type, and Zarf could avoid the strategy in [Removed Fields](#converting-removed-fields) where newer `ZarfPackage` structs track removed fields. This was rejected because while this strategy would work with some functions, many functions, especially in `packager`, accept a `packageLayout` object. Having multiple versions of these functions makes the SDK experience less user friendly since users would need extra calls between loading their packages and calling `packager` functions. Additionally, `packageLayout` has a public mutable field of type `v1alpha1.ZarfPackage`. Removing this field, limits the opportunity of SDK users to edit their packages before packager calls.

### Package Source Interface

Zarf could define an interface called `PackageSource`: 
```go
type PackageSource interface{
  // This function will be updated to the latest version whenever a new version is released
  GetPackageAtLatestAPIVersion() (v1beta1.ZarfPackage)
  GetV1Beta1Package() (v1beta1.ZarfPackage)
  GetV1Alpha1Package() (v1alpha1.ZarfPackage)
}
```
PackageLayout and DeployedPackage would both implement this interface. Functions such as `packager.Remove()` which accept either a built package or a cluster source would accept this interface. This would avoid specific package types in some function definitions. It could also allow for patterns like below where we could reach back to previous API versions to get to removed fields rather than storing [Removed Fields](#converting-removed-fields) on the objects. 

```go
if source.GetPackageAtLatestAPIVersion().Build.APIVersion == "v1alpha1" {
  dataInjections := source.GetV1alpha1Package().Components[x].DataInjections
  // ... run Data injection logic with this
}
```

This was rejected because packages in Zarf's lifecycle are mutable and not taken directly from the package YAML / Kubernetes secret. The filters package, for instance, frequently changes the package. Zarf wouldn't be able to filter a package, without needing to keep logic around to filter every API version, which would add a maintenance burden. SDK users wouldn't be able to edit their packages before running functions like `packager.Remove()` or `packager.Deploy()` as it would be difficult to propagate changes to every API version. 

### Interface Representation of Schema

Rather than updating functions to accept a newer version of the schema, Zarf could have an interface that all SDK functions accept. The interface would have getter functions for each item that is common between active schemas. 

The downside of this approach is that each API version has sub-structs for each item. For instance, each schema will it's own version of the [ZarfComponentActions](https://github.com/zarf-dev/zarf/blob/a26516131a5df8dd2ddc93ec1f2e59bd959c971d/src/api/v1alpha1/component.go#L246) and all of the sub-structs underneath this sub-struct. The interface would return an internal type that the concrete types would need to convert their data to. 

Another issue is that each function that accepts a sub-struct of the Zarf schema, would need to accept a the larger interface, even if it only a small part of the schema is required. Additionally, because there are items that are not common across schemas there would need to be type checks for certain schema versions. This would get more complex to maintain as more schemas version are added. 

### Automatically generating conversion functions

It would be possible to write automation to generate functions that will convert one the unchanged fields from one API version to another rather than a maintainer manually writing up these functions. Kubernetes takes this approach. 

Automation here is initially rejected because this is likely something that will only be done on a rare cadence, likely at least 6+ months between conversions. Additionally, the Zarf package schema is the only type to consider currently. For the foreseeable future, it will likely be simpler to generate manually. If we find that this takes up a significant amount of time, then this can be re-evaluated. 

### Map representation of Removed Fields

One option for storing removed fields on newer schemas is to use `.metadata.annotations` or a new field such as `Deprecated map[string]string`. Kubernetes takes the annotations approach. The downside of this approach is that annotations can easily get confusing and hard to read. When a list of objects such as `dataInjections` is removed, then Zarf needs to maintain a long string representation of YAML.
The reason Kubernetes takes this approach is because their data must make lossless round trips. Their objects might be written as v1beta1, stored as v1alpha1, then upgraded back to v1beta1, and they cannot lose any data. There is no place to store the information on their v1alpha1 object besides annotations. Zarf is going to write all active API versions to the zarf.yaml file so there is no chance of data loss. 

### Custom YAML marshalers for Removed Fields

Originally, [Removed Fields](#converting-removed-fields) proposed custom marshalers to track the private fields for backwards compatibility such as `dataInjections`. However, going through each case we see that custom YAML marshalers are not needed:
- A v1alpha1 package is created after function signatures are changed to accept v1beta1 objects.
  - This package is read as v1alpha1, then converted to v1beta1 for `packager.create()`, then written to the `zarf.yaml` in the package tar as v1alpha1. The package never needs to be written to disk as v1beta1.
- A v1alpha1 package is created before v1beta1 is introduced. The package is deployed after function signatures are changed to accept v1beta1 objects.
  - The package is read as v1alpha1, then converted to v1beta1 during deploy, then persisted to the cluster as v1alpha1. It is never represented on disk as v1beta1.
- A v1beta1 package is created.
  - The package is written as both v1alpha1 and v1beta1, but removed fields such as `dataInjections` cannot be set. 