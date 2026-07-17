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

All future package definition schemas will require a root level `apiVersion` field; the v1alpha1 schema already has an optional `apiVersion` field. If this field is not set, v1alpha1 will be assumed.

During Zarf's lifetime, it will introduce, deprecate, and remove support for ZarfPackageConfig API versions. When a version is initially deprecated, users will still be able to perform all package operations without a feature flag, but will receive warnings that they should upgrade. Removing an API version will require a new major version of Zarf and will take a phased approach. Before or during a new major version, package operations such as create, sign, inspect, publish, and deploy will be gated behind a feature flag for deprecated packages. The list command will show a warning if users have deprecated packages in their cluster, and encourage users to upgrade by deploying a package with a newer API version.

The zarf.yaml in a built package will include the package definition for every supported API version. When printing the package definition to the user, for example with the command `zarf package inspect definition`, the output will be the API version of the package by default. A new flag `--api-version` will be introduced to `zarf package inspect definition` and `zarf dev inspect definition` to allow configuring the output. 

A new command `zarf dev upgrade-schema` will be introduced to allow users to convert from one API version to another. The command will default to converting to the latest API version. It will output the converted package definition for the latest schema version. It will accept a path to a directory containing a zarf.yaml file and an optional flag, `--to`, to declare the API version. For instance, a user could run `zarf dev upgrade-schema . --to v1beta1` and they will receive the converted package definition to stdout. The command will not allow changing from a newer version to an older version, so running `zarf dev upgrade-schema . --to=v1alpha1` on a `v1beta1` schema will error. This command will only accept a local package definition, and will not accept created packages, published packages, or deployed packages. 

API versions of the package schema will not necessarily coincide with releases of the Zarf CLI. For instance, Zarf may release a 1.0 version, while the newest package definition API version is v1beta1.

Once an API version is released, fields will not be removed from it, and there will be no new required fields.

To keep the SDK stable as new API versions are introduced, `packager` will accept an implementer of the new [`PackageAccessor`](#the-packageaccessor-interface) interface that exposes each supported version. Most public functions will therefore stay unchanged across API versions.

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

As a package creator, I want to update my package definition to the v1beta1 schema, so I run `zarf dev upgrade-schema . --to v1beta1 > zarf.yaml` against the directory containing my zarf.yaml. The command writes the converted definition to stdout, replacing my existing zarf.yaml.

#### Story 4

As a Zarf maintainer, I want to introduce a new API version so that I can deprecate fields, add new required fields, and rename fields in the current package schema. I want the process to do this to be straightforward. I want earlier versions of Zarf to handle deploying packages from the new schema, even on versions not including the new schema. If the package is not deployable, then the user should see a clear version requirement. 

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

### Package Layout loses mutability
`layout.PackageLayout.Pkg` is currently a public, mutable field, and some SDK consumers edit it directly after loading a package — for example to rename it, rewrite annotations, or override the namespace. Replacing it with an opaque handle removes that general-purpose write access, which is a breaking change for those consumers.

This risk is tolerable as it makes sense to have safeguards on package mutations given that most arbitrary edits to a package would corrupt it. For instance, changing a chart name would cause a failure on deploy since the chart name is used to find the chart tarball within the package layout. The known post-load mutations are a small set and each may be exposed as a targeted setter (`SetName`, `SetAnnotations`, `OverrideNamespace`, `FilterComponents`); more can be added as consumer needs surface. See [Package Layout](#package-layout) for more detail.

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

A new API version may coincide with packages being incompatible with earlier versions of Zarf, but the logic for determining compatibility will be decoupled from the API version. Zarf will introduce a new field `build.VersionRequirements` which will be automatically populated on create, and will error on deploy or remove if the user's version is older than the required version. See [#4256](https://github.com/zarf-dev/zarf/issues/4256).

The zarf.yaml file within a built package will be separated by the standard YAML `---`. Currently, Zarf only checks the first YAML object in the zarf.yaml file. To maintain backwards compatibility, API versions will always be placed in ascending order beginning with the v1alpha1 definition. Future versions of Zarf will check the API version of each package definition and select the latest version that it understands. This process will be implemented before any new API versions are released. If Zarf sees a version that it does not understand, Zarf will log to the user that there is a new API version available that the user should consider updating to. 

### Conversions

Zarf will need to handle two use cases for conversions. The first is library convert functions. These functions will move a specific version to the internal, superset type. This is always lossless. The second is `zarf dev upgrade-schema`, which will provide a simple way for users to convert their zarf.yaml files from one schema version to the next.

#### Type API changes

The api packages will be structured as below:

```bash
# internal/api/types holds the superset working type; see below.
├── internal
│   └──api
│     └── types
│       └── package.go
│     └── v1alpha1
│       └── convert.go
│       └── validate.go   
│     └── v1beta1
│       └── convert.go
│       └── validate.go   
├── api
│   └──v1alpha1
│     ├── package.go
│     ├── ...
│   └──v1beta1
│     ├── package.go
│     ├── ...
│   └──convert
│     ├── convert.go
```

The `internal/api/types` package contains a superset of Zarf fields spanning all supported API versions. It plays two roles. First, it is the working representation that `PackageLayout` uses internally. Callers will obtain a versioned view through per-version read accessors, `AsV1alpha1()` and `AsV1beta1()`, which will translate the internal type to the specific version. Because the superset is never named in a public signature, introducing a new API version requires no function signature changes when `PackageAccessor` is accepted. 
Second, it is the pivot for conversions: rather than converting v1alpha1 directly to v1beta1, Zarf converts v1alpha1 to the superset then the superset to v1beta1, so Zarf needs only N conversion functions (one per API version) rather than N² conversions between every pair of versions.

The internal package will not be exposed by the SDK. Instead the convert package will expose functions such as `func V1Alpha1PkgToV1Beta1(in v1alpha1.ZarfPackage) v1beta1.Package`. These functions will call the internal API packages, `internalv1alpha1.ConvertToGeneric(in v1alpha1.ZarfPackage) types.Package` and `internalv1beta1.ConvertFromGeneric(in types.Package) v1beta1.Package`. This will provide a clean interface for SDK users while avoiding exposing the internal types. Zarf's own `src/packager` and `src/cmd` packages may import `internal/api/types` directly; the constraint is only that the superset never appears in a public SDK signature. This strategy will also keep the src/api/<version> packages focused solely on data rather than including validation or conversion logic. These conversion functions will be manually written as opposed to [automatically generating conversion functions](#automatically-generating-conversion-functions). 

Zarf will not expose a public method such as v1alpha1.Validate() as this is a subset of the package validation required, and contains only specific logic not covered by the schema. This validation logic, currently in src/pkg/lint/validate.go, will be moved to internal/api/v1alpha1. This structure will be implemented before v1beta1 is released, and added to with each new API version. Package validation will continue to occur in `load.PackageDefinition`, keeping the SDK flow the same. 

##### Converting 1:1 Replacements
If a field is renamed with a 1:1 replacement, then Zarf will automatically convert the field to its replacement. For example, if a field called `noWait` was changed to `wait` then the value of the field will flip during conversion.

##### Converting Removed Fields

When Zarf internally converts an older schema version to the internal superset type (for example, while deploying a v1alpha1 package), it must convert without data loss. Fields that are removed stay on the superset, but are absent from new API versions. A newer type such as `v1beta1` carries no backwards-compatibility fields. When an older package is loaded, its removed fields ride along on the superset for the lifetime of the in-memory package and are written back out whenever the package is rendered to that older version. Once the API version the fields originate from is no longer supported, that section of the superset is deleted.

#### zarf dev upgrade-schema

When running `zarf dev upgrade-schema`  if a user's package contains a removed field that does not have a 1:1 replacement, then the command will error. The error message will recommend an alternative approach to replacing the field. 

The usage docs for `zarf dev upgrade-schema` will look like below:

```bash
Converts and outputs the existing zarf package config to the given API version. Defaults to latest API version.

# Replace your zarf.yaml file with the latest API version
$ zarf dev upgrade-schema . > zarf.yaml

Usage:
  zarf dev upgrade-schema [ DIRECTORY ] [flags]
Flags:
  --to string      Specify the API version to upgrade the package definition to. Defaults to the newest schema version.
```

### The PackageAccessor interface

Zarf has three package sources: an on-disk built package (`PackageLayout`), a cluster-deployed package (`DeployedPackage`), and a loaded but not yet built package (`DefinedPackage`). `PackageLayout` and `DeployedPackage` implement a new interface that returns per-version package definitions.

```go
// PackageAccessor is the read contract implemented by built and cluster package sources.
type PackageAccessor interface {
	AsV1alpha1() (v1alpha1.ZarfPackage, error)
	AsV1beta1() (v1beta1.Package, error)
}
```

The accessors return an `error` because a cluster source (`DeployedPackage`) parses stored JSON that may be malformed or written at a version this Zarf does not understand.

Functions that operate on either a built package or a cluster source, such as `packager.Remove` and the `zarf package inspect` functions, accept a `PackageAccessor` rather than a concrete type. Functions specific to a single source still take that concrete type.

Once support is dropped for an API version, the interface will remove its associated reader. 

### Package Layout

`PackageLayout` is the handle SDK users receive from `LoadPackage`. It currently exposes a public, mutable `Pkg v1alpha1.ZarfPackage` field. This will be changed to an unexported field of the internal superset type. Reads will happen through the `PackageAccessor` accessors and mutations will go through methods instead of direct field access.

```go
type PackageLayout struct {
	dirPath string
	pkg     types.Package // was: Pkg v1alpha1.ZarfPackage
	// ...
}
```

Functions such as `packager.Deploy` will call the latest reader (for example `packageLayout.AsV1beta1`) for logic shared among all versions. When version-specific logic is reached, such as running data injections for a v1alpha1 package, then `packageLayout.AsV1alpha1()` will be called so this logic can run. 

#### Mutations

The accessors return copies, so persistent mutation goes through methods that edit the superset in place. These are a set of targeted setters. An initial list is below. There may be additional setters after evaluating SDK consumers.

```go
func (p *PackageLayout) SetName(name string)
func (p *PackageLayout) SetAnnotations(annotations map[string]string)
func (p *PackageLayout) OverrideNamespace(namespace string) error
func (p *PackageLayout) FilterComponents(filter filters.ComponentFilterStrategy) error
```

There is no generic `SetDefinition(v1beta1.Package)` function as replacing the package data with a versioned API package will be lossy if the package was initially created at another version. 

#### Deployed packages

The current deployed package struct is seen below. The `DeployedPackage` object is persisted to the cluster during `zarf package deploy` as a Kubernetes secret. The Data field is a JSON representation of the package.

```go
type DeployedPackage struct {
	Name               string               `json:"name"`
	Data               v1alpha1.ZarfPackage `json:"data"`
	CLIVersion         string               `json:"cliVersion"`
  ...
}
```

In the future, when Zarf stores this secret, it will store the version the package was created with as well as all earlier API versions, mimicking the strategy used in built packages. This will enable older versions of Zarf that don't have the latest API version to read newer packages. Additionally, when a user runs `zarf package inspect definition` on a cluster-sourced deployed package, they will receive a printed YAML of the API version they built the package with. To track multiple versions, a new field named `PackageData` of type `map[string]json.RawMessage` will be introduced on the struct. The original Data object will stay on the object for backwards compatibility, until the v1alpha1 package is no longer supported. DeployedPackage will implement the `PackageAccessor` interface. 

```go
type DeployedPackage struct {
	Name               string                     `json:"name"`
  // Data is kept for backwards compatibility, once support for reading v1alpha1 packages from the cluster is removed, this field will be deleted.
	Data               v1alpha1.ZarfPackage       `json:"data"`
  PackageData        map[string]json.RawMessage `json:"packageData"`
	...
}
```

### Filters

A filter is a component selection decision. It needs only a small projection of each component. The `filters` package owns that projection, so filters never see internal types and never break when a new schema ships:

```go
// ComponentView is the stable projection a filter sees
type ComponentView struct {
	Name        string
	Optional    bool
	Default     bool
	Group       string
	OnlyLocalOS string
}

type PackageView struct {
	Components []ComponentView
}

type ComponentFilterStrategy interface {
	// Apply returns the indices of the components to keep, in order.
	Apply(PackageView) ([]int, error)
}
```

`PackageLayout` will expose a function `FilterComponents(filter filters.ComponentFilterStrategy) error` to allow filtering on a package after it is loaded.

There will be other cases in the codebase where we decouple packages from an explicit API version, but they are omitted from this proposal for brevity. 

### JSON Schema

Zarf publishes a JSON schema, see the [current version](https://raw.githubusercontent.com/zarf-dev/zarf/refs/heads/main/zarf.schema.json). Users often use editor integrations to have built-in schema validation for zarf.yaml files. This strategy is [referenced in the docs](https://docs.zarf.dev/ref/dev/#vscode). The Zarf schema is also included in the [schemastore](https://github.com/SchemaStore/schemastore/blob/ae724e07880d0b7f8458f17655003b3673d3b773/src/schemas/json/zarf.json) repository.

Zarf will use the if/then/else features of the JSON schema to conditionally apply a schema based on the `apiVersion`. If the `apiVersion` is `v1alpha1` then the schema will evaluate the zarf.yaml file according to the v1alpha1 schema. If the `apiVersion` is v1beta1 then the zarf.yaml will be evaluated according to the v1beta1 schema. It's useful to have a single schema file, so that users' text editors handle different API versions without file-specific annotations. Zarf will still create and utilize individual version schemas. 


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

A new API version will not go through alpha/beta/GA, but instead will be GA when publicly released. Implementing a new schema version should follow this approach:
- Introduce the new Go type and schema.
- Implement the conversion and validation logic.
- Implement upgrading to the newer schema version with `zarf dev upgrade-schema`.
- Implement loading the new API version. This will enable commands such as `zarf dev inspect`, `zarf dev lint`, and `zarf dev find-images`.
- Implement creating and deploying packages with the new API version. Change public functions to use the new Go type. There should be extensive e2e tests for creating and deploying packages.

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

- 2025-10-18: Proposal submitted.
- 2025-12-08: Updated proposal to focus more on the process Zarf maintainers should follow to ensure that new API versions can be introduced.
- 2026-07-17: Design changed to have a PackageAccessor interface rather than using the latest version as the internal working type.

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

### Use only one API version

Instead of introducing a new schema, we could over time remove and introduce fields on the existing package schema. This was rejected as we believe a new schema will provide a more intuitive experience for developers and maintainers. For instance, attempting to implement the proposed Helm changes would result in a large, confusing schema. Likewise, it would be difficult to change defaults, such as moving from required to optional, since we wouldn't have a new API version to signal that there is a breaking change. Additionally, removing deprecated fields such as `dataInjections` would require a major version change. Keeping the API version and CLI version decoupled allows flexibility to make the changes we need.

### Latest version as the internal working type

An earlier version of this proposal made the latest API version (`v1beta1`) the internal working type that flows through `packager`, and changed every public function to accept the latest version. Deprecated fields removed in newer versions were carried on the newer type as private shim fields with getters and setters, rather than on the [internal type](#type-api-changes). This was rejected for two reasons:

- Conversion was lossless from an older version to the latest but not the reverse, since the latest version needed to carry every field to not break backwards compatibility when running packager functions. That asymmetry is confusing, and carrying every deprecated field on `v1beta1` is cumbersome to maintain.
- Changing every public function to the latest version is a costly breaking change for SDK consumers and hard to land piecemeal, since it amounts to a signature flip across the entire SDK.

### Public Facing Internal Type

Rather than updating functions to accept a newer version of the schema, Zarf could have a publicly facing internal type that has every field from every version and use that throughout the SDK. The upside of this approach is that we would avoid breaking changes throughout the lifetime of the SDK. The downside is that it would make it easy for anyone using the SDK to set deprecated fields. It would also make it confusing and unclear which fields attach to which versions. 

### Internal Type wrapped by Public Versioned Functions

Another way an internal type could be used would be to introduce public functions such as `packager.RemoveV1alpha1()` and `packager.RemoveV1beta1()`. These functions would then call a private `packager.remove()` function that accepts the internal type. This way SDK users don't have to deal with the internal type, and Zarf could avoid the strategy in [Removed Fields](#converting-removed-fields) where newer `ZarfPackage` structs track removed fields. This was rejected because while this strategy would work with some functions, many functions, especially in `packager`, accept a `packageLayout` object. Having multiple versions of these functions makes the SDK experience less user-friendly since users would need extra calls between loading their packages and calling `packager` functions. Additionally, `packageLayout` has a public mutable field of type `v1alpha1.ZarfPackage`. Removing this field limits the opportunity of SDK users to edit their packages before packager calls.

### Interface Representation of Schema

Rather than updating functions to accept a newer version of the schema, Zarf could have an interface that all SDK functions accept. The interface would have getter functions for each item that is common between active schemas. 

The downside of this approach is that each API version has sub-structs for each item. For instance, each schema will have its own version of the [ZarfComponentActions](https://github.com/zarf-dev/zarf/blob/a26516131a5df8dd2ddc93ec1f2e59bd959c971d/src/api/v1alpha1/component.go#L246) and all of the sub-structs underneath this sub-struct. The interface would return an internal type that the concrete types would need to convert their data to. 

Another issue is that each function that accepts a sub-struct of the Zarf schema would need to accept the larger interface, even if only a small part of the schema is required. Additionally, because there are items that are not common across schemas, there would need to be type checks for certain schema versions. This would get more complex to maintain as more schema versions are added. 

### Automatically generating conversion functions

It would be possible to write automation to generate functions that will convert the unchanged fields from one API version to another rather than a maintainer manually writing up these functions. Kubernetes takes this approach. 

Automation here is initially rejected because this is likely something that will only be done on a rare cadence, likely at least 6+ months between conversions. Additionally, the Zarf package schema is the only type to consider currently. For the foreseeable future, it will likely be simpler to generate manually. If we find that this takes up a significant amount of time, then this can be re-evaluated. 

### Map representation of Removed Fields

One option for storing removed fields on newer schemas is to use `.metadata.annotations` or a new field such as `Deprecated map[string]string`. Kubernetes takes the annotations approach. The downside of this approach is that annotations can easily get confusing and hard to read. When a list of objects such as `dataInjections` is removed, then Zarf needs to maintain a long string representation of YAML.

The reason Kubernetes takes this approach is because their data must make lossless round trips. Their objects might be written as v1beta1, stored as v1alpha1, then upgraded back to v1beta1, and they cannot lose any data. There is no place to store the information on their v1alpha1 object besides annotations. Zarf is going to write all active API versions to the zarf.yaml file so there is no chance of data loss. 
