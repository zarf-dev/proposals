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

# ZEP-0051: v1beta1 schema

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

Several fields in the v1alpha1 ZarfPackageConfig should be restructured to provide a more intuitive experience. Other fields that have a poor user experience and add unnecessary overhead to Zarf should be removed. Introducing a new schema version, v1beta1, provides the opportunity to make these changes.

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

There are several open issues requesting enhancements to the schema. The general theme of these changes is to make it easier to create Zarf packages.
- [Refactor charts definition in zarf.yaml #2245](https://github.com/zarf-dev/zarf/issues/2245)
- [Breaking Change: make components required by default #2059](https://github.com/zarf-dev/zarf/issues/2059)
- [Use kstatus as the engine behind zarf tools wait-for and .wait.cluster #4077](https://github.com/zarf-dev/zarf/issues/4077)

Additionally, users often struggle to use data injections. Usually, they would be better served by using a Kubernetes native solution [#3926](https://github.com/zarf-dev/zarf/issues/3926).

### Goals

<!--
List the specific goals of the ZEP. What is it trying to achieve? How will we
know that this has succeeded?
-->

- Detail the schema changes in the ZarfPackageConfig from v1alpha1 to v1beta1.

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->

- Discuss how the Zarf codebase will shift to handle multiple API versions. This is detailed in [0048-schema-update-process](../0048-schema-update-process/README.md).

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

Zarf will determine the schema of the package definition using the existing optional field `apiVersion`. Until v1alpha is removed, if `apiVersion` is not set, then Zarf will assume it is a v1alpha1 package. `apiVersion` will be a required field in v1beta1 and all future schemas. 

Type names will remove the prefix Zarf where applicable. For example, the type `ZarfMetadata` will become `Metadata`. This has no impact on the package yaml.

Users will be able to upgrade their package definitions using `zarf dev upgrade-schema`, which writes the converted definition to stdout. 

The v1beta1 schema will remove, replace, and rename several fields. View this [zarf.yaml](zarf.yaml) to see a package definition with reasonable values for each key.

### Schema changes

#### Removed Fields

If a package has these fields defined then `zarf dev upgrade-schema` will error and print a recommendation for an alternative.

- `.components.[x].group` will be removed. A similar functionality was introduced with the field `components[x].target.flavor`. This shifts component selection to the create side, and is the recommended replacement. 
- `.components.[x].default` will be removed. It was used to determine the default `.components[x].group`. It also gave the default to the (Y/N) prompt during interactive deploys, this use was secondary and not important enough to keep the field around. 
- `.components.[x].dataInjections` will be removed. https://docs.zarf.dev/best-practices/data-injections-migration/ details migrating off of this field.
- `.components.[x].charts.[x].variables` will be removed. Users are encouraged to use [Zarf values](../0021-zarf-values/) instead.
- `.variables` and `.constants` will be removed. Users are encouraged to use [Zarf values](../0021-zarf-values/) instead. While values will always be mutable on deploy, creators will be able to choose which values in their chart are mutable using `sourcePath/targetPath`. Similarly, creators decide which fields in manifests are mutable through values. 
- `.components.[x].actions.[onAny].setVariable` and `.components.[x].actions.[onAny].setVariables` will be removed. The existing `.components.[x].actions.[onAny].setValues` field is the replacement.
- `.metadata.yolo` will be removed. Its successor is connected deployments [#4580](https://github.com/zarf-dev/zarf/issues/4580).
- `.components.[x].only.cluster.distro` will be removed. This field was never used for anything and there are no plans to use it currently.

#### Replaced / Restructured Fields

`zarf dev upgrade-schema` will automatically migrate these fields.

- `.components.[x].actions.[onAny].after` will be combined with and replaced by `actions.[onAny].onSuccess`. Any `after` actions will be prepended to the `actions.[onAny].onSuccess` list.
- `.components.[x].scripts` will be removed. This field is already deprecated and will be migrated to the existing field `.components.[x].actions`.
- `.components.[x].only.cluster.architecture` will be inlined to `.components.[x].target.architecture`. This is more accurate as the field checks the `.metadata.architecture` on create, rather than the cluster during deploy. Note that `.only` was renamed to `.target`. Since `.cluster.distro` will be removed, the `.cluster` parent field will be deleted. 
- `.metadata` fields `image`, `source`, `documentation`, `url`, `authors`, and `vendor` will be removed. `zarf dev upgrade-schema` will move these fields under `.metadata.annotations`, which is a generic map of strings.
- `.components.[x].healthChecks` will be removed and appended to `.components.[x].actions.onDeploy.after.wait.cluster`. This will be accompanied by a behavior change in `zarf tools wait-for` to perform kstatus-style readiness checks when `.wait.cluster.condition` is empty. See [wait changes](#wait-changes).
- `.components.[x].charts` will be restructured to move fields into different sub-objects depending on the method of consuming the chart. See [Helm Chart Changes](#zarf-helm-chart-changes).
- `.components.[x].images` will move from a list of strings to a list of objects. The `Image` object will have a required field, `name`, and an optional enum, `source`. Allowed values for `source` will be `daemon` and `registry`. Zarf will no longer fall back to pulling images from the Docker Daemon. During component imports, the merge strategy will change from a simple append to a merge based on `name`. `source` and any future fields will favor the base component value if set, and otherwise use the imported component value. 
- `.components.[x].import.name` will be removed given that components will only be importable from component config files so there is not a name to select. See [ZarfComponentConfig](#zarfcomponentconfig).
- `.components.[x].import.path` and `.components.[x].import.url` will be changed into `.components.[x].import.local.[x].path` and `.components.[x].import.remote.[x].url`. All entries from both are combined when applying component compatibility rules. See [ZarfComponentConfig](#zarfcomponentconfig). These fields are a list of objects instead of a list of strings to enable future sibling fields. For instance, we may introduce a field `.components.[x].import.remote.[x].verify` to enable verifying the signature of signed remote components.
- `.components.[x].manifests.[x].kustomizations`, `.components.[x].manifests.[x].kustomizeAllowAnyDirectory`, and `.components.[x].manifests.[x].enableKustomizePlugins` will be moved into a `.components.[x].manifests.[x].kustomize` sub-object. The fields become `kustomize.files`, `kustomize.allowAnyDirectory`, and `kustomize.enablePlugins` respectively.

#### Renamed Fields

`zarf dev upgrade-schema` will automatically migrate these fields.

- `.metadata.aggregateChecksum` will move to `.build.aggregateChecksum`.
- `.build.terminal` will be renamed to `.build.hostname`.
- `.components.[x].manifests.[x].noWait` and `.components.[x].charts.[x].noWait` will be renamed to `skipWait`.
- `.components[x].required` will be renamed to `.components[x].optional`. `optional` will default to false. Since `required` currently defaults to false, components will now default to being required.
- `.components.[x].actions.[default/onAny].maxRetries` will be renamed to `.components.[x].actions.[default/onAny].retries`.
- `.components.[x].actions.[default/onAny].mute` will be renamed to `.components.[x].actions.[default/onAny].silent`.
- `.components.[x].manifests.[x].template`, `.components.[x].files.[x].template`, and `.components.[x].actions.[onAny].template` will be renamed to `enableValues`.
- `.components.[x].only` will be renamed to `.components.[x].target`.
- `.components.[x].only.localOS` will be renamed to `.components.[x].target.os`.
- `.components.[x].repos` will be renamed to `.components.[x].repositories`.
- `.components.[x].files.[x].target` will be renamed to `.components.[x].files.[x].destination`.
- `.components.[x].files.[x].shasum` will be renamed to `.components.[x].files.[x].checksum`. The field accepts the format `<algorithm>:<checksum>` (e.g. `sha256:abc123`); if no algorithm prefix is provided, sha256 is assumed.

### New Fields

- `.components[x].service` will be introduced to avoid magic names in Init package components. See [Zarf Services](#zarf-services) for more details.

### Behavior Changes

#### Wait Changes

There will be a behavior change in `.components[x].actions.[onAny].wait.cluster`. In the v1alpha1 ZarfPackageConfig, when `.cluster.condition` is empty, Zarf waits until the resource exists. In the v1beta1 schema, when `.cluster.condition` is empty, Zarf will wait for the resource to be ready using kstatus readiness checks.

#### Zarf Services

In the v1alpha1 schema, Zarf looks at init component names to determine when to run certain logic. For instance, the injector is always run when an init component has the name "zarf-seed-registry". These magical names have caused confusion for custom init package creators, [#4528](https://github.com/zarf-dev/zarf/issues/4528), and leave little room for configurability.

A new `service` key under components will make the inherent coupling between the init package and the Zarf CLI more transparent. The field is an enum with the allowed values `registry`, `seed-registry`, `injector`, `agent`, and `git-server`.

View the full schema in [package.go](package.go#L200).

```yaml
- name: zarf-registry
  service: registry
- name: zarf-agent
  service: agent
  ...
```

### ZarfInitConfig will be Removed

The `Kind` "ZarfInitConfig" will be removed. Every package will be of kind "ZarfPackageConfig". `zarf init` will default to deploying a package called `zarf-package-init-<arch>-<cli-version>.tar.zst`. A template will be created that exposes the CLI version, so a `zarf.tpl.yaml` file could set the `.metadata.version` field to `[[ .cli.version ]]`. If a package called `zarf-package-init-<arch>-<cli-version>.tar.zst` is not found in the cache or current directory, Zarf will prompt the user to pull the default zarf-dev init package. `zarf init` will continue to accept custom packages, for example, `zarf init <zarf-package-my-custom-init>`. If no component in the package declares a `.service`, Zarf will error and ask the user to run `zarf package deploy` instead. 

### ZarfComponentConfig

The v1beta1 APIVersion will introduce a new `Kind` alongside ZarfPackageConfig called ZarfComponentConfig. ZarfComponentConfig files will allow declaring a component to be imported from other packages. It will have its own schema, and this schema will be verified on create and publish. ZarfComponentConfigs will be importable only by v1beta1 packages. Components from other ZarfPackageConfigs will not be importable in v1beta1 packages.

Each ZarfComponentConfig declares exactly one component under the `component` field. If a user wants multiple variations of a component differentiated by flavor, OS, or architecture, they create one ZarfComponentConfig file per variation and set the `.component.target` field on each. View the ZarfComponentConfig schema in [design details](#zarf-component-config-schema).

The component in a ZarfComponentConfig will be able to import another ZarfComponentConfig. Cyclical imports will error. ZarfComponentConfig files will not have a default filename such as zarf.yaml. This will encourage users to give their files descriptive names and promote a flatter directory structure as users will not default to having a new folder for each component. ZarfComponentConfigs will be able to define their own values and valuesSchema.

`.import.local` is a list of local file path references to ZarfComponentConfig files; directories are not accepted. `.import.remote` is a list of `oci://` URL references to remote component configs pulled at create time. All entries from both fields are combined when applying compatibility rules: when more than one entry is given, every referenced component must share the same name, and at most one of them must be compatible with the active package target (flavor, OS, architecture) at create time otherwise Zarf will error.

The `zarf dev` commands that accept a directory containing a `zarf.yaml` (lint, inspect, and find-images) will accept component config files. For example, `zarf dev inspect definition my-component-config.yaml`.

#### Remote Components

Skeleton packages will be replaced by remote components. Instead of publishing an entire package, users will be able to publish a ZarfComponentConfig. This component will behave similarly to Skeleton packages in that local resources will be published alongside it, while remote resources will be pulled at create time.

Remote components will be published using the new command `zarf component publish <component-file> <oci-repo>`. This command will have the flag `--flavor` to publish a component whose `.component.target.flavor` matches the supplied value.

Unlike Skeleton packages, which are published with unresolved templates, remote components must be fully templated before publishing. By templating before publish, we avoid issues with validating a non-templated package ([#4491](https://github.com/zarf-dev/zarf/issues/4491)) and stay aligned with the overall [Package Templates](#package-templates) strategy.

### Package Templates

The Zarf v1alpha1 schema allows for package templates during create using the ###ZARF_PKG_TMPL_*### format. This format will be replaced in the v1beta1 schema with Go templating. Additionally, instead of templating on create, a new command `zarf dev template` will be introduced. This command will take in a zarf.tpl.yaml file, and will output a zarf.gen.yaml file based on the Go templating result. The command will accept a flag `--set` to set templates and a flag `--set-file` which will accept a values style file to define templates.

The `.gen` extension will be used to easily discern between generated and included packages. It will also make it simple to ignore these files within Git repositories. When `zarf package create`, or any other relevant command, is run on a directory, it will first look for a `zarf.yaml`, then fall back to a `zarf.gen.yaml`.

`zarf dev template` will have logic to follow local component imports. For any entry in `.import.local` whose `path` points to a file called `<base>.tpl.yaml`, Zarf will template the `<base>.tpl.yaml` file and rewrite the entry to `<base>.gen.yaml`. Users that prefer to template in separate steps may set their import path entries to `<base>.gen.yaml` directly. Zarf will template imports after the current file is finished templating, so a user will be able to template a value into an entry of `.import.local` and Zarf will template the resulting file.

Package templates will be required to have a value; otherwise the command will fail.

The delimiter for Go templates during `zarf dev template` will be `[[ ]]`. This will separate package templates from the standard Go template delimiter `{{ }}`, which is used during on-deploy actions.

### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1

As a user of Helm charts in my package, I have the following existing `zarf.yaml`:

```yaml
kind: ZarfPackageConfig
metadata:
  name: helm-charts
components:
  - name: demo-helm-charts
    required: true
    charts:
      - name: podinfo-local
        version: 6.4.0
        namespace: podinfo-from-local-chart
        localPath: chart
        valuesFiles:
          - values.yaml

      - name: podinfo-oci
        version: 6.4.0
        namespace: podinfo-from-oci
        url: oci://ghcr.io/stefanprodan/charts/podinfo
        valuesFiles:
          - values.yaml

      - name: podinfo-git
        version: 6.4.0
        namespace: podinfo-from-git
        url: https://github.com/stefanprodan/podinfo.git
        gitPath: charts/podinfo
        valuesFiles:
          - values.yaml

      - name: podinfo-repo
        version: 6.4.0
        namespace: podinfo-from-repo
        url: https://stefanprodan.github.io/podinfo
        name: podinfo
        releaseName: cool-release-name
        valuesFiles:
          - values.yaml
```

I want to upgrade to the v1beta1 schema, so I run `zarf dev upgrade-schema . > zarf.yaml`, which produces:

```yaml
apiVersion: zarf.dev/v1beta1
kind: ZarfPackageConfig
metadata:
  name: helm-charts
components:
  - name: demo-helm-charts
    optional: false  # Changed from `required: true`
    charts:
      - name: podinfo-local
        namespace: podinfo-from-local-chart
        local:
          path: chart  # Changed from `localPath`
        # version field removed - uses version from local chart.yaml
        valuesFiles:
          - values.yaml
      - name: podinfo-oci
        namespace: podinfo-from-oci
        oci:
          url: oci://ghcr.io/stefanprodan/charts/podinfo
          version: 6.4.0
        valuesFiles:
          - values.yaml

      - name: podinfo-git
        namespace: podinfo-from-git
        git:
          url: https://github.com/stefanprodan/podinfo.git@6.4.0
          path: charts/podinfo  # Changed from `gitPath`
        # version field removed - uses version from chart.yaml at git tag
        valuesFiles:
          - values.yaml

      - name: podinfo-repo
        namespace: podinfo-from-repo
        helmRepository:
          url: https://stefanprodan.github.io/podinfo
          name: podinfo  # Changed from `repoName`
          version: 6.4.0
        releaseName: cool-release-name
        valuesFiles:
          - values.yaml
```

#### Story 2

As a package creator, I have a logging component that I maintain locally and want to use a monitoring component published by my team to an OCI registry. I combine both into a single v1beta1 package.

First, I define my local logging component in `logging.yaml`:

```yaml
apiVersion: zarf.dev/v1beta1
kind: ZarfComponentConfig
metadata:
  name: logging
component:
  charts:
    - name: loki
      namespace: logging
      local:
        path: loki-chart
      valuesFiles:
        - loki-values.yaml
```

My teammate has published a monitoring component to our registry. Its source file, `monitoring.yaml`, looked like this before publishing:

```yaml
apiVersion: zarf.dev/v1beta1
kind: ZarfComponentConfig
metadata:
  name: monitoring
  version: 1.0.0
component:
  charts:
    - name: kube-prometheus-stack
      namespace: monitoring
      helmRepository:
        url: https://prometheus-community.github.io/helm-charts
        name: kube-prometheus-stack
        version: 60.0.0
      valuesFiles:
        - prometheus-values.yaml
```

They published it with:

```bash
zarf component publish monitoring.yaml oci://ghcr.io/my-org/components
```

Now I create a v1beta1 package that imports both components -- the local one by file path and the remote one by URL:

```yaml
apiVersion: zarf.dev/v1beta1
kind: ZarfPackageConfig
metadata:
  name: observability
  description: Combines logging and monitoring into a single package

components:
  - name: logging
    import:
      local:
        - path: logging.yaml
  - name: monitoring
    import:
      remote:
        - url: oci://ghcr.io/my-org/components/monitoring:1.0.0
```

I can then create my package as usual:

```bash
zarf package create
```

#### Story 3

As a package creator, I want to template image references and metadata into my package at build time. I write a `zarf.tpl.yaml` that uses Go templates with the `[[ ]]` delimiter:

```yaml
apiVersion: zarf.dev/v1beta1
kind: ZarfPackageConfig
metadata:
  name: app
  description: "app [[ .ENVIRONMENT ]]"

components:
  - name: app
    charts:
      - name: app
        namespace: app
        oci:
          url: oci://ghcr.io/my-org/charts/my-app
          version: 1.0.0
    images:
      - name: [[ .MY_IMAGE ]]
```

I generate a `zarf.gen.yaml` for a specific release:

```bash
zarf dev template --set ENVIRONMENT=personal --set MY_IMAGE=ghcr.io/my-org/my-image:0.0.1
```

This produces `zarf.gen.yaml`:

```yaml
apiVersion: zarf.dev/v1beta1
kind: ZarfPackageConfig
metadata:
  name: app
  description: "app personal"

components:
  - name: app
    charts:
      - name: app
        namespace: app
        oci:
          url: oci://ghcr.io/my-org/charts/my-app
          version: 1.0.0
    images:
      - name: ghcr.io/my-org/my-image:0.0.1
```

I can then create my package from the generated file:

```bash
zarf package create
```

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

The field `.components.[x].dataInjections` will be removed without a direct replacement in the schema. The docs website added a [migration page](https://docs.zarf.dev/best-practices/data-injections-migration/) to inform users how to switch.

The alpha field `.components.[x].charts.[x].variables` has seen significant adoption and there will be no automatic conversion to its replacement Zarf values. There must be documentation on how users can utilize Zarf values as an alternative to chart variables.

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

### Zarf Helm Chart Changes

The `Chart` object will be restructured as seen in [package.go](package.go#L242-L308). Exactly one of sub-objects `helmRepository`, `git`, `oci`, or `local` is required for each entry in `components.[x].charts`. The fields `localPath`, `gitPath`, `URL`, and `repoName` will be removed from the top level of `components.[x].charts`. See [#2245](https://github.com/zarf-dev/zarf/issues/2245).

During conversion, Zarf will detect the method of consuming the chart and create the proper sub-objects. If a git repo is used, then `@` + the `.version` value will be appended to `.git.URL`. This is consistent with the current Zarf behavior.

Zarf uses the top-level `version` field to determine where in the package layout file structure it will place charts. This makes the field necessary for deploy, and therefore it must be carried over using the strategy defined in the removed fields section of [0048-schema-update-process](../0048-schema-update-process/README.md#converting-removed-fields). Newer versions of Zarf will ensure that Zarf works whether or not `version` is set. Packages created with the v1beta1 schema will leave `version` empty, and therefore will not work with earlier versions of Zarf. When support is dropped for v1alpha1 packages, the `version` field will be dropped entirely. Note that this process is applied to internal conversion so that there is no change in behavior when v1alpha1 packages use function signatures that contain v1beta1 objects. `zarf dev upgrade-schema` will simply move the top-level `version` field to the right sub-object, or drop it when not applicable.

### Zarf Component Config Schema

View the schema for the Zarf component config in [componentConfig.go](componentConfig.go).

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

There will be e2e tests for creating, deploying, and publishing a v1beta1 package. Existing tests will shift to use the v1beta1 schema over time.

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

The v1beta1 schema will not have an alpha/beta/GA phase. It will follow the graduation criteria laid out in [0048-schema-update-process](../0048-schema-update-process/README.md#graduation-criteria).

Deprecation:
- This schema will likely be deprecated one day in favor of a v1 schema. It will not be deprecated until after the next schema version is generally available. Once deprecated, Zarf will still support the v1beta1 schema for at least one year.

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

See the proposal section in [0048-schema-update-process](../0048-schema-update-process/README.md#proposal).

### Version Skew Strategy

<!--
If applicable, how will the component handle version skew with other
components? What are the guarantees? Make sure this is in the test plan.

Consider the following in developing a version skew strategy for this
proposal:
- Does this proposal involve coordinating behavior between components?
  - (i.e. the Zarf Agent and CLI? The init package and the CLI?)
-->

See the version skew strategy in [0048-schema-update-process](../0048-schema-update-process/README.md#version-skew-strategy).

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

- 2025-10-21: Proposal submitted
- 2026-02-12: Introduced Zarf Component Config, package templating changes, and Zarf services
- 2026-05-06: Simplified ZarfComponentConfig to one component per file; moved variants to alternatives

## Drawbacks

<!--
Why should this ZEP _not_ be implemented?
-->

### Component Import Reworks
Removing the ability to import components from packages directly, and instead requiring Zarf Component Config files, will require a sizable portion of the user base to rewrite files. This rewrite should leave users with a clearer directory structure, enhanced package validation, and a more intuitive import system.

Removing the ability to import from ZarfPackageConfig files will add some friction to standalone packages that are also imported. For instance, the [k3s sub-package](https://github.com/zarf-dev/zarf/blob/main/packages/distros/k3s/zarf.yaml) in the init package is deployable as a standalone package and imported by the init package. The proposed system would require creating a component config as well as a separate standalone k3s package that imports the component config to maintain the current structure. This drawback is deemed necessary to avoid packages that are only meant for import and not deployable as a standalone package. This has caused confusion among many users, and forcing creators to explicitly make a sub-package deployable will avoid this issue.

### Component Config

There is an implicit ordering in a zarf.yaml file: the first component in a list is installed, then the second, and so forth. By asking users to break apart their zarf.yaml files into Zarf Component Config files, they may lose this implicit ordering, and it could be more confusing to determine the order of components. 

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Component Config Schema

#### Variants in ZarfComponentConfig

A previous version of this proposal allowed a ZarfComponentConfig to declare either a `.component` field or a `.variants[]` field. The `.component` field was a single object representing a component importable by any package. The `.variants` field was a list of components where each entry had to specify a `.target` block (e.g. flavors, OSes, or architectures) to differentiate itself from the other entries. Zarf would error if two entries under `.variants` shared the same target. The `zarf component publish` command would have grown a `--all-variants` flag to publish every variant in one file at once.

This was rejected in favor of "exactly one component per file" to keep the mental model simple. With variants, a single file could expand into many components depending on flavor/OS/arch, and authors had to reason about which entry applied where. Forcing a 1:1 file-to-component mapping makes the import tree easy to follow at the cost of a few extra files for components with multiple targets.

#### List of Components

Another possibility for the [component config schema](#zarf-component-config-schema) was to have a list of components under a `.components[]` field, where each entry must specify a `.target` block. This was rejected since a major change in this system is that `ZarfComponentConfig` files represent a single component. The plural `.components[]` key would likely confuse users on this aspect.

#### Variants Extend Base Component

Another possibility for the [component config schema](#zarf-component-config-schema) is to have a single `.component` field that can be extended by a list of `.variants`. The `.component` field would be required, and could be imported or published as defined. It could also be extended using the `.variants` field. The logic for extending would exactly mirror the [component import logic](https://docs.zarf.dev/ref/components/#component-imports); the variant would import the base component.

This would be especially useful when there are multiple configurations of a chart, such as the example below. Each flavor prescribes its own values file and images, but otherwise is the same. A similar situation is seen in the [k3s sub-package](https://github.com/zarf-dev/zarf/blob/main/packages/distros/k3s/zarf.yaml) of the main Zarf repository. The only differences between the two k3s components are the files that vary by architecture.

```yaml
apiVersion: zarf.dev/v1beta1
kind: ZarfComponentConfig
metadata:
  name: grafana
component:
  - charts:
    - name: grafana-config
      namespace: grafana
      local:
        path: chart
      valuesFiles:
        - chart/values.yaml
variants:
  - target:
      flavor: upstream
    charts:
      - name: grafana
        valuesFiles:
          - values/upstream-values.yaml
    images:
      - name: docker.io/grafana/grafana:12.4.2

  - target:
      flavor: enterprise
    charts:
      - name: grafana
        valuesFiles:
          - values/enterprise-values.yaml
    images:
      - name: enterprise.corp.org/grafana/grafana:12.4.2
```

### Remote Component Templating

Remote components cannot be templated during import; this is a removed feature from its predecessor Skeleton packages. This allows Zarf to validate the component before it's published ([#4491](https://github.com/zarf-dev/zarf/issues/4491)) and is necessary since package templating now occurs before create. A potential alternative is a templated remote component where `zarf dev template oci://ghcr.io/<my-remote-component>` would download the component from OCI and template it. The user would then be able to import the component from their local directory. This was rejected because it adds complexity for a niche use case. This could be a future enhancement if the demand exists.

### Component Level Action Defaults

Action defaults could be set once at the component level rather than separately under each action set (`onCreate`, `onDeploy`, `onRemove`). This would reduce the schema's surface area. 

This was rejected. Create and deploy often run on separate hosts and have different jobs: `onCreate` actions typically pull files or load images as docker tars, while `onDeploy` actions typically run `kubectl` or stand up a cluster. Sharing defaults across that boundary creates an awkward mental model. The [example v1beta1 zarf.yaml](./zarf.yaml) is large because every action set has its own defaults block, but in practice actions are an escape hatch used sparingly. It is rare for a real component to define both `onCreate` and `onDeploy` actions.