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

# ZEP-0039: Enhance YOLO Experience

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
<!-- /toc -->

## Summary

This ZEP proposes a streamlined method of publishing Yolo and Airgap packages based on a single `Zarf.yaml` file. The intention is to provide a means of distributing a single configured package regardless of whether or not the destination environment is behind an airgap, similar to how packages for different architectures are distributed.

## Motivation

As a user of Zarf and a maintainer of a large package which may or may not be distributed to internet-connected or airgapped environments, I would like a way to distribute this large package without managing multiple `zarf.yaml` files. The zarf package I maintain includes a few Helm Charts which are deployed with Flux. Using Flux, users are able to add additional values.yaml files to these Helm charts at deploy time. The airgap package is 6GB in size, which is completely unecessary for my internet-connected users, who only require the Yolo package containing the Helm charts.

I would like use Zarf as a tool to create packages for both of these environments automatically, however the experience so far has been a lot of work and scripting to generate multiple `zarf.yaml` files and ensure they are correctly formatted and accurately implement the package I am trying to distribute. A process which in practice has not been foolproof.

Zarf should support Yolo packages natively, and in my opinion as a standard part of the packaging process and OCI image construction.

[kubernetes slack]: https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1754333786975669

### Goals

- Update the zarf package create process to build a Yolo package alongside the airgap package natively
- Allow users to specify Yolo-Only and Airgap-Only conditions/predicates for components in their `zarf.yaml`
- Remove `yolo: true/false` property from `zarf.yaml` schema (since this should be a toggle on package create/deploy)

### Non-Goals

- Integrating new Yolo features in the deploy process. In my opinion, this proposal should only apply to the creation of new packages as Zarf already has a `zarf dev deploy` command to emulate the deployment of local packages, and creating a Yolo package can be deployed already.
- Change the name `YOLO Mode` to something indicating production ready, internet-connected environments. That change should be a separate proposal and take feedback from usage of features in this proposal.

## Proposal

The most significant change to Zarf would be how it handles packaging at create, as once this optional feature is enabled, after creating a standard zarf package, it would remove all images from and create an additional Yolo package. Using optimized OCI layers, this Yolo package could be a subset of the airgap package layers, or the Airgap package could use the Yolo layers as a base. Additionally, users taking advantage of this feature should be able to specify whether or not an action, file, manifest, etc. should only be included in the Yolo package, the Airgap package, or both. A default naming convention should probably be used to standardize the name of the packages, similar to the package architecture.

### User Stories (Optional)

- As a package maintainer, I would like my release pipeline to generate zarf packages for both airgap and internet-connected automatically

- As a user, I would like more zarf packages to be available in Yolo-mode when I do not require airgap features

- As someone who thinks Zarf is pretty damn awesome, I would like to use it to manage my deployments to not just my airgap environments, but my cloud environments as well.

- As a user deploying into an internet-connected environment, I don't want to waste bandwidth downloading 6GB worth of images because the package maintainers didn't automatically build and publish a Yolo package.

### Risks and Mitigations

- Removing the `yolo` property would be a breaking change. An alternative to removing the property would be to disable this proposed feature if that flag exists.

- There may be some inherint differences in how packages are able to be deployed between airgap and connected environments, which may require constant feature updates to this feature to support.

## Design Details

- Schema changes
- CLI Flag

### `ZarfPackageConfig` Schema Changes

1. Make the `yolo: bool` property of zarf.yaml accept string values of "true" and "false" (or enable zarf template variables to accept non-string values). This will allow me to use zarf package templates to create either a Yolo package or a regular zarf package.

```
kind: ZarfPackageConfig
metadata:
  name: dynamic-packaging
  description: Build packages in both Yolo and Airgap mode with a single config
  version: '###ZARF_PKG_TMPL_CUSTOM_VERSION###'
  yolo: '###ZARF_PKG_TMPL_YOLO###'
```

**--OR--**

Get rid of the the `yolo: bool` metadata property all together. I think this makes the most sense, as users should be taking advantage of OCI layering or tagging or something similar to multi-architecture builds but for Yolo environments. This probably deserves its own section to discuss.

2. Add a Yolo property to `only` so that components are only included when a zarf package is either Yolo or not Yolo.

```
components:
  - name: Yolo-extras
    required: true
    description: "This component is only included in Yolo packages."
    only:
      yolo: true
```

3. Enable Yolo-only variables. Sometimes I need a variable defined with Yolo mode but not for normal airgap packages. An obvious example would be ImagePullSecrets, since Zarf agent will handle that for all packages, but in Yolo-mode, this could be handled separately. Setting an ImagePullSecret variable lets me require a secret at deploy time, and being able to specify that the variable is only required on Yolo packages lets me define all variables regardless of package type.

```
constants:
  - name: YOLO_LABEL
    description: "Label to apply to resources in Yolo packages"
    only:
        yolo: true
  - name: AIRGAP_LABEL
    description: "Label to apply to resources in airgap packages"
    only:
        yolo: false

variables:
  - name: IMAGE_PULL_SECRET
    description: "Used only with Yolo-mode. Any airgap packages that reference this variable will probably fail."
    only:
        yolo: true
```

4. Enable Yolo-only actions. If a component's `only` property supported a yolo-only value, I could still run into a situation where a component that is to be included in a either yolo or airgap-only mode contains a specific action (or action set) which may only be required exclusively in Yolo-mode.

```
actions:
  onDeploy:
    after:
    - cmd: echo "This command is for Yolo-only packages"
       only:
         yolo: true
```

5. Enable Yolo-only valuesFiles. If a chart's `valuesFiles` property has a list item which is an object instead of a string, parse the path from a `path` property and determine inclusion based on `only` property of the object.

```
valuesFiles:
  - ./path/to/values.yaml
  - path: ./path/to/yolo/values.yaml
    only:
      yolo: true
```

---

The result of these proposed changes could result in a Zarf package config of the following:

```
kind: ZarfPackageConfig
metadata:
  name: dynamic-packaging
  description: Build packages in both Yolo and Airgap mode with a single config
  version: '###ZARF_PKG_TMPL_CUSTOM_VERSION###'
  yolo: '###ZARF_PKG_TMPL_YOLO###'

constants:
  - name: YOLO_LABEL
    description: "Label to apply to resources in Yolo packages"
    only:
        yolo: true
  - name: AIRGAP_LABEL
    description: "Label to apply to resources in airgap packages"
    only:
        yolo: false

variables:
  - name: IMAGE_PULL_SECRET
    description: "Used only with Yolo-mode. Any airgap packages that reference this variable will probably fail."
    only:
        yolo: true

components:
  - name: Yolo-extras
    required: true
    description: "This component is only included in Yolo packages."
    only:
      yolo: true
    charts:
      - name: chart
        namespace: zarf
        version: 0.0.1
        localPath: ./path/to/chart
        only:
          yolo: true
        valuesFiles:
          - ./path/to/values.yaml
          - path: ./path/to/yolo/values.yaml
            only:
              yolo: true
        variables:
          - description: The chart values
            name: VALUES
            path: "$"
            only:
              yolo: false
    actions:
      onDeploy:
        after:
        - cmd: echo "This command is for Airgap-only packages"
          only:
            yolo: false
        - wait:
          only:
            yolo: true
            network:
              address: "some address only reachable when connected to internet"

    images: []
```

The biggest concern here is the scope of each `only` property, which could be nested as we drill down into the package manifest, and may not be an intuitive solution for users. On the other hand, supporting the `only` property at arbitrary YAML paths does provide exceptional composability with regard to templating package builds and generating build pipelines, not just for Yolo/Airgap, but for multiple platforms, architectures, os, or any other arbitrary criteria.

Looking at the example above, what happens when a component is marked `yolo`, but a chart within the component is marked `airgap`? Should zarf throw an error or a warning, or should zarf silently ignore that configuration (someone might have an odd but legitimate reason to configure something like that...)?

This also begs the question, what should Zarfs approach be to managing the `only` filter on any other objects or any future objects which get added to the schema? Should this filter be optional and implemented on a per feature basis, or should it be required with generalized unit tests?

### Package Create Options

There are multiple considerations for build-time CLI options depending on whether or not the `metadata.yolo` property is removed or improved. If the property is retained, then there isn't a required change at build-time, as using `--set` with a tempalte variable would suffice.

TODO: If it is removed, or the decision is made to expand the options for creating multi-architecture style packages for airgap/yolo environments, there will be way more options to consider and risks it would entail. This section should be dedicated for that discussion.

#### OCI Layers?

#### Multi-Architecture?

#### Multiple tagged images?

#### Yolo Flavor support?

### Handling Images

#### Yolo deployments of Airgap packages and vice versa

When a package is destined for a Yolo environment, there shouldn't be a requirement to remove all images/data. If I already have an airgap package, I should be able to deploy it to a Yolo environment and have it just ignore pushing images/data, or push images/data if the environment has zarf initialized. Similarly, If a package has no image or data injections, it should be able to deploy to airgap or yolo.

I guess the question I have is; what are the scenerios where an airgap package would fail in a Yolo environment or vice versa? Is it possible to simply capture these edge cases and provide options to work around them? There may be more nuanced issues I'm not aware of, but I wanted to pose the question and suggest such an explanation be added to future documentation on this feature.

#### Handle remote vs local images when yolo is specified

Assume I have a zarf package that is configured with images which are accessible over the internet. With an airgap package, I build the package locally and we are done. With a Yolo package, (assuming images are ignored), those image URLs would simply not be mutated. Since the images are remotely available to the environment, k8s should still be able to pull them down.

Now consider some of the images are only available locally or are private. Zarf would be able to package a normal airgap package, but fail in a yolo mode. Would it be worth discussing how zarf can be made aware of these differences?

Consider for example a mixed environment, 3 local/private images and 3 public images included in a zarf package. A Yolo-only environment would always fail, because those packages need to be reachable. Airgap works as expected. But is there a use case for a yolo deployment to a zarf initialized, connected environment? In this case, zarf packages could include only private/local images, and any public images could be ignored and pulled at deploy time in the environment? This would requre zarf to be smart enough to determine this information at build-time.

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

- Manage Yolo and Airgap packages separately with two different `zarf.yaml` files. This is error prone creates issues with testing changes properly between Yolo and Airgap deployments. It opens the door to drift between versions. Using this approach either requires two declaritive `zarf.yaml` files with lots of duplication, or a declarative `zarf.yaml` that dynamically updates, and is therefore not declarative.
  [Slack suggestion](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1754485517547389?thread_ts=1754333786.975669&cid=C03B6BJAUJ3)

- Make `yolo` property of zarf schema accept a string instead of a boolean. This would allow a zarf template variable to dynamically change the yolo-mode at package create time. This is a decent enough workaround, but wouldn't enhance the capability of Zarf's Yolo-Mode experience. Additionally, it wouldn't provide opportunities to make yolo/airgap specific changes to components.
