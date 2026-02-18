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

Several fields in the v1alpha1 ZarfPackageConfig can be restructured to provide a more intuitive experience. Other fields that have a poor user experience and add unnecessary overhead to Zarf should be removed. A new schema version, v1beta1, provides the opportunity to make these changes. 

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

- Discuss how the Zarf codebase will shift to handle multiple API versions. This is detailed in [0048-schema-upgrade-process](https://github.com/zarf-dev/proposals/pull/49)

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

Zarf will determine the schema of the package definition using the existing optional field `apiVersion`. If `apiVersion` is not set, then Zarf will assume it is a v1alpha1 package. Users will be able to automatically upgrade their package to the v1beta1 schema by running `zarf dev upgrade-schema`. `apiVersion` will be a required field in v1beta1. 

The v1beta1 schema will remove, replace, and rename several fields. View this [zarf.yaml](zarf.yaml) to see a package definition with reasonable values for each key. 

### Schema changes

#### Removed Fields

If a package has these fields defined then `zarf dev upgrade-schema` will error and print a recommendation for an alternative.

- `.components.[x].group` will be removed. Users will be recommended to use `components[x].only.flavor` instead.     
- `.components.[x].dataInjections` will be removed. There will be a guide in Zarf's documentation for alternatives. See [#3926](https://github.com/zarf-dev/zarf/issues/3926). 
- `.components.[x].charts.[x].variables` will be removed. Its successor is [Zarf values](../0021-zarf-values/), but there will be no automated migration with `zarf dev upgrade-schema`.
- `.components.[x].default` will be removed. It set the default option for groups and (y/n) interactive prompts for optional components. Groups are removed, and we've generally seen the user base shift away from optional components. 
- `.metadata.yolo` will be removed. Its successor will be connected deployments [#4580](https://github.com/zarf-dev/zarf/issues/4580)
- `.components.[x].import.name` will be removed given that component imports will be changed. See [ZarfComponentConfig](#zarfcomponentconfig)

#### Replaced / Restructured Fields

`zarf dev upgrade-schema` will automatically migrate these fields.

- `.components.[x].actions.[onAny].onSuccess` will be removed. Any `onSuccess` actions will be appended to the `actions.[onAny].after` list.
- `.components[x].actions.[onAny].setVariable` will be removed. This field is already deprecated and will be migrated to the existing field `.components[x].actions.[onAny].setVariables`.
- `.components.[x].scripts` will be removed. This field is already deprecated and will be migrated to the existing field `.components.[x].actions`. 
- `.metadata` fields `image`, `source`, `documentation`, `url`, `authors`, `vendors` will be removed. `zarf dev upgrade-schema` will move these fields under `.metadata.annotations`, which is a generic map of strings.
- `.components.[x].healthChecks` will be removed and appended to `.components.[x].actions.onDeploy.After.wait.cluster`. This will be accompanied by a behavior change in `zarf tools wait-for` to perform kstatus style readiness checks when `.wait.cluster.condition` is empty. See [Zarf Tools wait-for Changes](#zarf-tools-wait-for-changes).
- `.component.[x].charts` will be restructured to move fields into different sub-objects depending on the method of consuming the chart. See [Helm Chart Changes](#zarf-helm-chart-changes)
- `.component.[x].images` will move from a list of strings to a list of objects. The ZarfImage object will have a required field, `name`, and an optional enum, `source`. Allowed values for `source` will be `daemon` and `registry`. Zarf will no longer fall back to pull images from the Docker Daemon.

#### Renamed Fields

`zarf dev upgrade-schema` will automatically migrate these fields.

- `.metadata.aggregateChecksum` will move to `.build.aggregateChecksum`.
- `.components[x].required` will be renamed to `.components[x].optional`. `optional` will default to false. Since `required` currently defaults to false, components will now default to being required.
- `noWait` will be renamed to `wait`. `wait` will default to true. This change will happen on both `.components.[x].manifests` and `.components.[x].charts`.
- `.components.[x].actions.[default/onAny].maxRetries` will be renamed to `.components.[x].actions.[default/onAny].retries`.
- `.components.[x].actions.[default/onAny].maxTotalSeconds` will be renamed to `.components.[x].actions.[default/onAny].timeout`, which must be in a [Go recognized duration string format](https://pkg.go.dev/time#ParseDuration).

### New Fields

- `.components[x].features` will be introduced to avoid magic names in Init package components. See [Zarf Features](#zarf-features) for more details.

### Behavior Changes

#### Wait Changes

There will be a behavior change in `.components[x].actions.[onAny].wait.cluster`. In the v1alpha1 ZarfPackageConfig when `.cluster.condition` is empty Zarf will wait until the resource exists. In the v1beta1 schema, when `.cluster.condition` is empty Zarf will wait for the resource to be ready using kstatus readiness checks. 

#### Zarf Features

In the v1alpha1 schema, Zarf looks at init component names to determine when to run certain init logic. For instance, the injector is always run when an init component has a name "zarf-seed-registry". These magical names have caused confusion for custom init package creators [#4528](https://github.com/zarf-dev/zarf/issues/4528) and leave little room for configurability. 

There will be a new "features" key on components that should make the inherent coupling between the init package and the Zarf CLI more transparent. It'll also allow for setting specific properties using Zarf values. For instance, a user will be able to set tolerations for the injector dynamically on deploy by setting `.features.injector.values.tolerations` to `".injector.tolerations"`. The registry and agent features don't allow setting specific values, as those features already have Helm charts. There will be validation that ensures that Features are only used in packages that are `Kind: ZarfInitConfig`. This validation will run after the import chain is resolved. 

View the full schema in [Zarf Features Schema](#zarf-features-schema). There will not be a separate schema for `ZarfInitConfig` and `ZarfPackageConfig` objects to avoid complexity given Zarf Features are the only difference.

```yaml
- name: zarf-seed-registry
  features:
    isRegistry: true
    injector:
      enabled: true
      values:
        tolerations: ".injector.tolerations"
```

### ZarfComponentConfig

The v1beta1 APIVersion will introduce a new `Kind` alongside ZarfPackageConfig called ZarfComponentConfig. ZarfComponentConfig files will allow declaring a component to be imported from other packages. It will have its own schema, and this schema will be verified on create and publish. ZarfComponentConfigs will be importable only from v1beta1 packages. Components from other ZarfPackageConfigs will not be importable in v1beta1 packages.

A ZarfComponentConfig must define exactly one of `component` or `variants`. The `component` field is a single object representing a component that is always importable. The `variants` field is a list of components where each entry must specify the `.only` key to define when that variant applies (e.g. flavors, OSs, or architectures). If the `.only` key has the same value for two variants, the user will receive an error. View the schema of this object in [design details](#zarf-component-config-schema).

The component in a ZarfComponentConfig will be able to import another ZarfComponentConfig. Cyclical imports will error. ZarfComponentConfig files will not have a default filename such as zarf.yaml. This will encourage users to give their files descriptive names and help encourage a flatter directory structure as users will not default to having a new folder for each component. ZarfComponentConfigs will be able to define their own values and valuesSchema.

The `.import.path` field will not accept directories; users will give the filepath to the ZarfComponentConfig file they are importing.

The `zarf dev` commands that accept a directory containing a `zarf.yaml`, lint, inspect, and find-images, will accept component config files. For instance, `zarf dev inspect definition my-component-config.yaml`.

#### Remote Components

Skeleton packages will be replaced by remote components. Instead of publishing an entire package, users will be able to publish a ZarfComponentConfig. This component will behave similarly to Skeleton packages in that local resources will be published alongside it, while remote resources will be pulled at create time.

Remote components will be published using the new command `zarf component publish <component-file>`. This command will have the flags `--flavor` and `--all-variants`. When `--all-variants` is used, all variants will be published regardless of their `.only` block. If the `.component` block is supplied instead of a `.variants` block, `--all-variants` will have no effect. 

Unlike Skeleton packages, which are published with unresolved templates, remote components must be fully templated before publishing. See [Package Templates](#package-templates) for more detail.

### Package Templates

The Zarf v1alpha1 schema allows for package templates during create using the ###ZARF_PKG_TMPL_*### format. This format will be replaced in the v1beta1 schema with Go templating. Additionally, instead of templating on create, a new command `zarf dev template` will be introduced. This command will take in a zarf.tpl.yaml file, and will output a zarf.gen.yaml file based on the go templating result. The command will accept a flag `--set` to set templates and a flag `--set-file` which will accept a values style file to define templates.

The `.gen` extension will be used to easily discern between generated and included packages. It will also make it simple to ignore these files within Git repositories. When `zarf package create`, or any other relevant command, is run on a directory, it will first look for a `zarf.yaml`, then fall back to a `zarf.gen.yaml`.  

`zarf dev template` will have logic to follow local component imports. If the `.import.path` points to a file called `<base>.tpl.yaml` Zarf will template the file and edit the value of `import.path` to be `<base>.gen.yaml`. Users that prefer to template in separate steps may set their import path to `<base>.gen.yaml`. Zarf will template imports after the current file is finished templating, so a user will be able to template the value of `.import.path` into a `<base>.tpl.yaml` file and Zarf will template the given file.

Package templates will be required to have a value; otherwise the command will fail. 

The delimiter for Go templates during `zarf dev template` will be `[[ ]]`. This will separate package templates from the standard Go template delimiter `{{ }}` which are used during on-deploy actions.

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
        repoName: podinfo
        releaseName: cool-release-name
        valuesFiles:
          - values.yaml
```

I want to upgrade to the v1beta1 schema, so I run `zarf dev upgrade-schema`, which produces:

```yaml
apiVersion: v1beta1
kind: ZarfPackageConfig
metadata:
  name: helm-charts
  description: Example showcasing multiple ways to deploy helm charts
  version: 0.0.1

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
        helmRepo:
          url: https://stefanprodan.github.io/podinfo
          name: podinfo  # Changed from `repoName`
          version: 6.4.0
        releaseName: cool-release-name
        valuesFiles:
          - values.yaml
```

#### Story 3

As a package creator, I have a logging component that I maintain locally and want to use a monitoring component published by my team to an OCI registry. I combine both into a single v1beta1 package.

First, I define my local logging component in `logging.yaml`:

```yaml
apiVersion: v1beta1
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
apiVersion: v1beta1
kind: ZarfComponentConfig
metadata:
  name: monitoring
  version: 1.0.0
component:
  charts:
    - name: kube-prometheus-stack
      namespace: monitoring
      helmRepo:
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
apiVersion: v1beta1
kind: ZarfPackageConfig
metadata:
  name: observability
  description: Combines logging and monitoring into a single package

components:
  - name: logging
    import:
      path: logging.yaml
  - name: monitoring
    import:
      url: oci://ghcr.io/my-org/components/monitoring:1.0.0
```

I can then create my package as usual:

```bash
zarf package create
```

#### Story 4

As a package creator, I want to template image references and metadata into my package at build time. I write a `zarf.tpl.yaml` that uses Go templates with the `[[ ]]` delimiter:

```yaml
apiVersion: v1beta1
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
apiVersion: v1beta1
kind: ZarfPackageConfig
metadata:
  name: my-app
  description: "my-app personal"

components:
  - name: my-app
    charts:
      - name: my-app
        namespace: my-app
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

The field `.components.[x].dataInjections` will be removed without a direct replacement in the schema. There must be documentation to present to users so they know what alternatives they can use to achieve a similar result. 

The alpha field `.components.[x].charts.[x].variables` has seen significant adoption and there will be no automatic conversion to its replacement Zarf values. There must be documentation on how users can utilize Zarf values as an alternative to chart variables. 

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

### Zarf Helm Chart Changes

The ZarfChart object will be restructured to match the code block below. Exactly one of sub-objects `helmRepo`, `git`, `oci`, or `local` is required for each entry in `components.[x].charts`. The fields `localPath`, `gitPath`, `URL`, and `repoName` will be removed from the top level of `components.[x].charts`. See [#2245](https://github.com/defenseunicorns/zarf/issues/2245).

During conversion, Zarf will detect the method of consuming the chart and create the proper sub-objects. If a git repo is used, then `@` + the `.version` value will be appended to `.gitRepoSource.URL`. This is consistent with the current Zarf behavior. 

Zarf uses the top level `version` field to determine where in the package layout file structure it will place charts. This makes the field necessary for deploy, and therefore it must be carried over using the strategy defined in the removed fields section of [0048](https://github.com/zarf-dev/proposals/pull/49/files). Newer versions of Zarf will ensure that Zarf works whether or not `version` is set. Packages created with the v1beta1 schema will leave `version` empty, and therefore will not work with earlier versions of Zarf. When support is dropped for v1alpha1 packages, the `version` field will be dropped entirely. Note, this process is applied to internal conversion so that there is no change in behavior when v1alpha1 packages use function signatures that contain v1beta1 objects. `zarf dev upgrade-schema` will simply move the top level `version` field to the right sub object, or drop it when not applicable. 

```go
// ZarfChart defines a helm chart to be deployed.
type ZarfChart struct {
	// The name of the chart within Zarf; note that this must be unique and does not need to be the same as the name in the chart repo.
	Name string `json:"name"`
  // The version of the chart. This field is removed for the schema, but kept as a backwards compatibility shim so v1alpha1 packages can be converted to v1beta1
  version string
	// The Helm repo where the chart is stored
	HelmRepo HelmRepoSource `json:"helmRepo,omitempty"`
	// The Git repo where the chart is stored
	Git GitRepoSource `json:"git,omitempty"`
	// The local path where the chart is stored
	Local LocalRepoSource `json:"local,omitempty"`
	// The OCI registry where the chart is stored
	OCI OCISource `json:"oci,omitempty"`
	// The namespace to deploy the chart to.
	Namespace string `json:"namespace,omitempty"`
	// The name of the Helm release to create (defaults to the Zarf name of the chart).
	ReleaseName string `json:"releaseName,omitempty"`
	// Whether to not wait for chart resources to be ready before continuing.
	Wait *bool `json:"wait,omitempty"`
	// List of local values file paths or remote URLs to include in the package; these will be merged together when deployed.
	ValuesFiles []string `json:"valuesFiles,omitempty"`
  // [alpha] List of values sources to their Helm override target
	Values []ZarfChartValue `json:"values,omitempty"`
}

// HelmRepoSource represents a Helm chart stored in a Helm repository.
type HelmRepoSource struct {
	// The name of a chart within a Helm repository
	RepoName string `json:"repoName,omitempty"`
	// The URL of the chart repository where the helm chart is stored.
	URL string `json:"url"`
  // The version of the chart to deploy; for git-based charts this is also the tag of the git repo by default (when not using the '@' syntax for 'repos').
	Version string `json:"version"`
}

// GitRepoSource represents a Helm chart stored in a Git repository.
type GitRepoSource struct {
	// The URL of the git repository where the helm chart is stored.
	URL string `json:"url"`
	// The sub directory to the chart within a git repo.
	Path string `json:"path,omitempty"`
}

// LocalRepoSource represents a Helm chart stored locally.
type LocalRepoSource struct {
	// The path to a local chart's folder or .tgz archive.
	Path string `json:"path"`
}

// OCISource represents a Helm chart stored in an OCI registry.
type OCISource struct {
	// The URL of the OCI registry where the helm chart is stored.
	URL     string `json:"url"`
	Version string `json:"version"`
}
```

### Zarf Component Config Schema

The schema for the Zarf component config will look like so:

```go
// ComponentConfig is the top-level structure of a Zarf component config file.
type ComponentConfig struct {
	// The API version of the component config.
	APIVersion string `json:"apiVersion,omitempty" jsonschema:"enum=zarf.dev/v1beta1"`
	// The kind of component config.
	Kind ZarfPackageKind `json:"kind" jsonschema:"enum=ZarfComponentConfig,default=ZarfComponentConfig"`
	// Component metadata.
	Metadata ZarfComponentMetadata `json:"metadata"`
  // Exactly one of Component or Variants must be set.
	// A single component definition that applies in all contexts.
	Component *Component `json:"component,omitempty"`
	// A list of component variants, each with a distinct .only filter. Use this when the
	// component has different definitions for different flavors, OSs, or architectures.
	Variants []Variant `json:"variants,omitempty"`
	// Values imports Zarf values files for templating and overriding Helm values.
	Values ZarfValues `json:"values,omitempty"`
	// Zarf-generated publish data for the component config.
	PublishData ComponentPublishData `json:"publishData,omitempty"`
}

// Component is a reduced component definition used in component configs.
type Component struct {
	// Import a component from another Zarf component config.
	Import ZarfComponentImport `json:"import,omitempty"`
	// Kubernetes manifests to be included in a generated Helm chart on package deploy.
	Manifests []ZarfManifest `json:"manifests,omitempty"`
	// Helm charts to install during package deploy.
	Charts []ZarfChart `json:"charts,omitempty"`
	// Files or folders to place on disk during package deployment.
	Files []ZarfFile `json:"files,omitempty"`
	// List of OCI images to include in the package.
	Images []ZarfImage `json:"images,omitempty"`
	// List of Tar files of images to bring into the package.
	ImageArchives []ImageArchive `json:"imageArchives,omitempty"`
	// List of git repos to include in the package.
	Repos []string `json:"repos,omitempty"`
	// Custom commands to run at various stages of a package lifecycle.
	Actions ZarfComponentActions `json:"actions,omitempty"`
  // Features of the Zarf CLI 
  Features ZarfComponentFeatures `json:features,omitempty"`
}

// Variant is a component definition with a required filter for when it applies.
type Variant struct {
	Component
	// Filter when this variant is included in package creation or deployment.
	Only ZarfComponentOnlyTarget `json:"only"`
}

// ZarfComponentMetadata holds metadata about a component config.
type ZarfComponentMetadata struct {
	// Name to identify this component config.
	Name string `json:"name" jsonschema:"pattern=^[a-z0-9][a-z0-9\\-]*$"`
	// Additional information about this component config.
	Description string `json:"description,omitempty"`
	// Generic string to track the component config version.
	Version string `json:"version,omitempty"`
	// Annotations contains arbitrary metadata about the component config.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// ComponentPublishData is written during publish to track details of the component config.
type ComponentPublishData struct {
	// The version of Zarf used to build this component config.
	ZarfVersion string `json:"zarfVersion"`
	// Any migrations that have been run on this component config.
	Migrations []string `json:"migrations,omitempty"`
	// Requirements for specific package operations.
	VersionRequirements []VersionRequirement `json:"versionRequirements,omitempty"`
}
```

### Zarf Features Schema

The schema for Zarf Features:

```go
type ZarfComponentFeatures struct {                                                                                                                             
  IsRegistry bool       `json:"isRegistry,omitempty"`
  Injector   *Injector  `json:"injector,omitempty"`
  IsAgent    bool       `json:"isAgent,omitempty"`
}                                                                                                                                                      
                                                                                                                                                                  
type Injector struct {
  Enabled bool             `json:"enabled"`
  Values  *InjectorValues  `json:"values,omitempty"`
}

type InjectorValues struct {
  Tolerations string `json:"tolerations,omitempty"`
}
```

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

There will be e2e tests for creating, deploying, and publishing a v1beta1 package. As the schema is nears GA, existing tests will shift to use the v1beta1 schema.

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

The v1beta1 schema will not have an alpha/beta/GA phase. Creating a package with the v1beta1 schema will initially be behind a feature flag for at least two releases after the v1beta1 schema is introduced. During this time, the v1beta1 schema may remove or rename fields. Once the feature flag is enabled by default, there will be no removed or renamed fields until the next schema version. 

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

See proposal in [ZEP-0048](https://github.com/zarf-dev/proposals/issues/48).

### Version Skew Strategy

<!--
If applicable, how will the component handle version skew with other
components? What are the guarantees? Make sure this is in the test plan.

Consider the following in developing a version skew strategy for this
proposal:
- Does this proposal involve coordinating behavior between components?
  - (i.e. the Zarf Agent and CLI? The init package and the CLI?)
-->

See version skew strategy in [ZEP-0048](https://github.com/zarf-dev/proposals/issues/48).

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
- 2025-02-12: Introduce Zarf Component Config, package templating changes, and Zarf features

## Drawbacks

<!--
Why should this ZEP _not_ be implemented?
-->

### Component Import Reworks
Removing the ability to import components from packages directly, and instead requiring Zarf Component Config files, will require a sizable portion of the user base to rewrite files. We believe this is a worthwhile tradeoff as this re-write should leave users with a clearer directory structure, enhanced package validation, and a more intuitive import system.  

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->

### Component Import Schema

Another possibility for the [component imports schema](#component-import-schema) instead of allowing for one of `.component` or `.variants[]` was to simply have a list of components. The list of components would allow for multiple entries, so long as each entry had a `.only` block. This was rejected since a major change in this system is that `ZarfComponentConfig` files represent a single component. The list key `.components[]` would likely confuse users on this aspect. Separate keys for `.component` and `.variants[]` also allows for builtin schema validation, requiring the `.only` key with `.variants[]` but not with `.component`. 