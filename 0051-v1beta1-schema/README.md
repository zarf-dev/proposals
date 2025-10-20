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

Several fields in the ZarfPackageConfig v1alpha1 can be restructured to provide a more intuitive experience. Some field in the v1alpha1 schema have a poor user experience and add overhead to Zarf, these will be removed. A new schema version, v1beta1, provides Zarf the space to make these changes. 

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

- Discuss how the codebase will change to handle a new schema version. This is detailed in 0048-schema-upgrade-process

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

The v1beta1 schema will remove or rename several fields.

- `.metadata.aggregateChecksum` will move to `.build.aggregateChecksum`
- `.metadata` fields `image`, `source`, `documentation`, `url`, `authors`, `vendors` -> will be removed. `zarf dev convert` will automatically add them as fields under `.metadata.annotations`.
- `.components[x].required` will be renamed to `.components[x].optional`. `optional` will default to false, this is a change in behavior since required defaults to false.
- `.components.[x].group` will be removed. Users are recommend to use `components[x].only.flavor` instead.
- `setVariable` will be removed. It can be automatically migrated to the existing field `setVariables`.  
- `scripts` will be removed. It can be automatically migrated to the existing field `actions`. 
- `noWait` will be renamed to `wait`. `wait` will default to true. This change will happen on both `.components.[x].manifests` and `components.[x].charts`
- `yolo` will be renamed to `airgap`. `airgap` will default to true
- `.components.[x].actions.[default/onAny].maxRetries` -> `.components.[x].actions.[default/onAny].retries`
- `.components.[x].actions.[default/onAny].maxTotalSeconds` -> `.components.[x].actions.[default/onAny].timeout`, which must be in a [Go recognized duration string format](https://pkg.go.dev/time#ParseDuration)
- `.component.[x].charts` will break off fields into different sub-objects depending on the method of consuming the chart. See [#2245](https://github.com/defenseunicorns/zarf/issues/2245). Exactly one of `helm`, `git`, `oci`, or `local` must exist for each `components.[x].charts`, and their objects look like below. The fields `localPath`, `gitPath`, `version`, `URL`, and `repoName` will all be removed from the top level of `components.[x].charts`. 
```yaml
- name: podinfo-repo-new
  helm:
    url: https://stefanprodan.github.io/podinfo
    name: podinfo # replaces repoName since it's only applicable for helm chart repositories
    version: 6.4.0

- name: podinfo-git-new
  git:
    url: https://stefanprodan.github.io/podinfo@6.4.0
    path: charts/podinfo
    # no version field, Zarf will use the version in the chart.yaml at that git tag

- name: podinfo-oci-new
  oci:
    url: oci://ghcr.io/stefanprodan/charts/podinfo
    version: 6.4.0 

- name: podinfo-local-same
  local:
   path: chart
  # no version field, use local chart.yaml version
```
- `.components.[x].healthChecks` will be removed in favor of changing the behavior of `.components.[x].actions.[onAny].wait.cluster` to use Kstatus when the `.wait.cluster.condition` is empty. `.wait.cluster` currently shells out to `kubectl wait`. Kstatus checks are generally preferred as the user doesn't need to set a condition, instead Kstatus has inherent knowledge of how to check the readiness of a resource. The advantages of `.wait.cluster` are that specific conditions can be set. This can be useful when readiness is not the desired state, or for certain CRDs that do not implement the fields for Kstatus readiness checks. The original behavior of `.wait.cluster` will be used when `.wait.cluster.condition` is set. 
  - Since Kstatus requires the API version, `apiVersion` will be added as a field to `.wait.cluster`.
  - `.healthChecks` always occur after deploy so `zarf dev convert` will migrate them to `.components[x].actions.onDeploy.After.wait.cluster`.
- `.components.[x].dataInjections` will be removed from the v1beta1 schema without replacement. See [#3926](https://github.com/zarf-dev/zarf/issues/3926). 
- `.components.[x].charts.[x].variables` will be removed. It's successor is [Zarf values](../0021-zarf-values/), but there will be no automated migration with `zarf dev convert`.
- `.components.[x].actions.[onAny].onSuccess` will be removed. Any onSuccess actions, will be migrated to the end of `actions.[onAny].after`.

In order for this schema to be applied, users must set `apiVersion` to `v1beta1`. If `apiVersion` is not set then Zarf will assume it is a v1alpha1 package. Users will be able to automatically upgrade their package to the v1beta1 schema by running `zarf dev convert`. 

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

The `healthChecks` field has been removed and replaced with an `actions.onDeploy.after` wait action that uses Kstatus for health checking when no explicit condition is set.

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

The fields `.components.[x].dataInjections` will be removed without a direct replacement in the schema. There must be documentation to present to users so they know what alternatives they can use achieve a similar result. 

The alpha field `.components.[x].charts.[x].variables` has seen significant adoption and we will not be able to automatically convert users to Zarf values with `zarf dev convert`. There should be documentation on how users can utilize Zarf values as an alternative to chart variables. 

## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->

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

There will be e2e tests for creating, deploying, and publishing a v1beta1 package. As the schema nears towards GA, the current v1alpha1

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
- GA: We've received feedback that all of are changes are an improvement. Examples and tests in Zarf shift to using the v1beta1 schema.

Deprecation:
- This schema will likely be deprecated one day in the future in favor of a v1 schema. It will not be deprecated until the next schema version is at least generally available. Once deprecated, Zarf will still support the v1beta1 schema for at least a year.

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
