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

Several fields in the v1alpha1 ZarfPackageConfig can be restructured to provide a more intuitive experience. Other fields have a poor user experience and add unnecessary overhead to Zarf; these fields will be removed. A new schema version, v1beta1, provides the space to make these changes. 

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

Zarf will determine the schema of the package definition using the top level `apiVersion` field. `apiVersion` already exists as a top level field in the Zarf package config schema. If `apiVersion` is not set then Zarf will assume it is a v1alpha1 package. Users will be able to automatically upgrade their package to the v1beta1 schema by running `zarf dev convert`. 

The v1beta1 schema will remove, restructure, and rename several fields.

### Removed fields without replacement

These fields will error when `zarf dev convert` is run and recommend an alternative method to achieve the desired behavior. 

- `.components.[x].group` will be removed. Users will be recommended to use `components[x].only.flavor` instead.
- `.components.[x].dataInjections` will be removed. There will be a guide in Zarf's documentation for alternatives. See [#3926](https://github.com/zarf-dev/zarf/issues/3926). 
- `.components.[x].charts.[x].variables` will be removed. Its successor is [Zarf values](../0021-zarf-values/), but there will be no automated migration with `zarf dev convert`.

### Removed fields with automated replacement

`zarf dev convert` will automatically migrate these fields.

- `.components.[x].actions.[onAny].onSuccess` will be removed. Any `onSuccess` actions will be appended to the `actions.[onAny].after` list.
- `.components[x].actions.[onAny].setVariable` will be removed. This field is already deprecated and will be migrated to the existing field `.components[x].actions.[onAny].setVariables`.
- `.components.[x].scripts` will be removed. This field is already deprecated and will be migrated to the existing `.components.[x].actions`. 
- `.metadata` fields `image`, `source`, `documentation`, `url`, `authors`, `vendors` will be removed. `zarf dev convert` will move these fields under `.metadata.annotations`, which is a generic map of strings.
- `.components[x].actions.[onAny].wait.cluster` will receive a new required sub field, `.apiVersion`. During conversion `.apiVersion` will be added to the object but kept empty. Users will be warned that they must fill this field out, otherwise create will error. 
- `.components.[x].healthChecks` will be removed and appended to `.components.[x].actions.[onAny].wait.cluster`. This will be accompanied by a behavior change in `zarf tools wait-for` to perform kstatus style readiness checks when `.wait.cluster.condition` is empty. See [Zarf Tools wait-for Changes](#zarf-tools-wait-for-changes).
- `.component.[x].charts` will be restructured to move fields into different sub-objects depending on the method of consuming the chart. See [Helm Chart Changes](#zarf-helm-chart-changes)

### Renamed fields

`zarf dev convert` will automatically migrate these fields.

- `.metadata.aggregateChecksum` will move to `.build.aggregateChecksum`
- `.metadata.yolo` will be renamed to `.metadata.airgap`. `airgap` will default to true
- `.components[x].required` will be renamed to `.components[x].optional`. `optional` will default to false. Since `required` currently defaults to false, components now default to being required by default.
- `noWait` will be renamed to `wait`. `wait` will default to true. This change will happen on both `.components.[x].manifests` and `.components.[x].charts`.
- `.components.[x].actions.[default/onAny].maxRetries` will be renamed to `.components.[x].actions.[default/onAny].retries`
- `.components.[x].actions.[default/onAny].maxTotalSeconds` will be renamed to `.components.[x].actions.[default/onAny].timeout`, which must be in a [Go recognized duration string format](https://pkg.go.dev/time#ParseDuration)

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

I want to upgrade to the v1beta1 schema so I run `zarf dev convert`, which produces:

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
        helm:
          url: https://stefanprodan.github.io/podinfo
          name: podinfo  # Changed from `repoName`
          version: 6.4.0
        releaseName: cool-release-name
        valuesFiles:
          - values.yaml
```

#### Story 2

As a user of health checks in my package, I have the following `zarf.yaml` in v1alpha1:

```yaml
kind: ZarfPackageConfig
metadata:
  name: health-checks
  description: Deploys a simple pod to test health checks

components:
  - name: health-checks
    required: true
    manifests:
      - name: ready-pod
        namespace: health-checks
        noWait: true
        files:
          - ready-pod.yaml
    healthChecks:
      - name: ready-pod
        namespace: health-checks
        apiVersion: v1
        kind: Pod
```

I want to move to the latest schema so I run `zarf dev convert`, which produces:

```yaml
apiVersion: v1beta1
kind: ZarfPackageConfig
metadata:
  name: health-checks
  description: Deploys a simple pod to test health checks

components:
  - name: health-checks
    optional: false  # Changed from `required: true`
    manifests:
      - name: ready-pod
        namespace: health-checks
        wait: false  # Changed from `noWait: true`
        files:
          - ready-pod.yaml
    actions:
      onDeploy:
        after:
          - wait:
              cluster:
                kind: Pod
                name: ready-pod
                namespace: health-checks
                apiVersion: v1
                # condition is empty, so Kstatus will be used for readiness check
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

The alpha field `.components.[x].charts.[x].variables` has seen significant adoption and there will be no automatic conversion to it's replacement Zarf values. There must be documentation on how users can utilize Zarf values as an alternative to chart variables. 

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

### Zarf Helm Chart Changes

The ZarfChart object will be restructured to match the code block below. Exactly one of sub-objects `helm`, `git`, `oci`, or `local` must exist for each `components.[x].charts`. The fields `localPath`, `gitPath`, `URL`, and `repoName` will be removed from the top level of `components.[x].charts`. See [#2245](https://github.com/defenseunicorns/zarf/issues/2245).

During conversion, Zarf will detect the method of consuming the chart and create the proper sub-objects. If a git repo is used then `@` + the `.version` value will be appended to `.gitRepoSource.URL`. This is consistent with the current Zarf behavior. 

Zarf uses the top level `version` field to determine where in the package layout file structure it will place charts. This makes the field necessary for deploy, and therefore it must be carried over using the strategy defined in the removed fields section of [0048](https://github.com/zarf-dev/proposals/pull/49/files). Newer versions of Zarf will ensure that Zarf works whether or not `version` is set. Packages created with the v1beta1 schema will leave `version` empty, and therefore not work with earlier versions of Zarf. When support is dropped for v1alpha1 packages the `version` field will be dropped entirely. Note, this process is applied to internal conversion so that there is no change in behavior when v1alpha1 packages use  function signatures that contain v1beta1 objects. `zarf dev convert` will simply move the top level `version` field to the right sub object, or drop it when not applicable. 

```go
// ZarfChart defines a helm chart to be deployed.
type ZarfChart struct {
	// The name of the chart within Zarf; note that this must be unique and does not need to be the same as the name in the chart repo.
	Name string `json:"name"`
  // The version of the chart. This field is removed for the schema, but kept as a backwards compatibility shim so v1alpha1 packages can be converted to v1beta1
  version string
	// The Helm repo where the chart is stored
	Helm HelmRepoSource `json:"helm,omitempty"`
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

#### Zarf Tools wait-for Changes

`zarf tools wait-for` is the underlying engine to `.wait.cluster`. Currently, `zarf tools wait-for` shells out to `zarf tools kubectl wait`. In the future, `wait-for` will optionally accept an API version alongside the resource kind. If API version is set, and condition is empty then kstatus will be used as `wait-for`'s engine. 

v1alpha1 packages do not have `.apiVersion` as a sub field under `.wait.cluster`, so they will always use the existing engine, avoiding breaking changes. v1beta1 packages will require that `apiVersion` is set.

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

There will be e2e tests for `zarf dev convert` from a v1alpha1 definition to a v1beta1 definition.

There will be e2e tests for creating, deploying, and publishing a v1beta1 package. As the schema nears towards GA, existing tests will shift to use the v1beta1 schema.

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

- Alpha: fields are subject to change or rename. No backwards compatibility guarantees.
- Beta: Fields will not change in a way that is not fully backwards compatible.
- GA: Users have provided feedback that the new schema improves the UX. Examples and tests in Zarf shift to using the v1beta1 schema.

Deprecation:
- This schema will likely be deprecated one day in favor of a v1 schema. It will not be deprecated until after the next schema version generally available. Once deprecated, Zarf will still support the v1beta1 schema for at least one year.

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

See proposal in ZEP-0048

### Version Skew Strategy

<!--
If applicable, how will the component handle version skew with other
components? What are the guarantees? Make sure this is in the test plan.

Consider the following in developing a version skew strategy for this
proposal:
- Does this proposal involve coordinating behavior between components?
  - (i.e. the Zarf Agent and CLI? The init package and the CLI?)
-->

See version skew strategy in ZEP-0048

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
