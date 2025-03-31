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

# ZEP-0021: Zarf Values

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

This ZEP proposes to introduce a new way to provide configuration to Zarf packages that is more in line with the Helm values paradigm.  This proposal would provide a new settable interface that uses `map[string]interface{}` instead of the current variable interface of `map[string]string`.  These values would also be able to be mapped directly to Helm chart values, and would introduce Go templating to the Zarf package configuration to support non-Helm features like `actions`.

This ZEP supercedes [ZEP-0015](0015-helm-value-style-variable-interfaces/README.md).

## Motivation

The motivation for this centers around the desire to align Zarf Variables closer to Helm Values which users in the Zarf community are generally already familiar with.  This proposal would also provide more flexibility for what values could be set to since the Zarf Values would be a full `interface{}` rather than simply `string`s. Pull Request [#2132](https://github.com/zarf-dev/zarf/pull/2131) initially allowed Zarf Variables to be directly passed as Helm values, but this always had limitations due to the design of Zarf Variables being geared toward string templating.  Over time, there has been desire from the Zarf community to treat Zarf Variables more like Helm Values [[1](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1706175082741539)], [[2](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1702400472208839)], [[3](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1706299407255849?thread_ts=1706175134.116329&cid=C03B6BJAUJ3)] and reevaluating variables entirely allows us to define more clearly how these values should work across Zarf's featureset (including non-Helm features like `actions`).

### Goals

- Design a replacement for Zarf Variables for members of the community familiar with Helm
- Provide more flexibility for what values can be and how they can be used
- Integrate this new design to work across Zarf's featureset

### Non-Goals

- Entirely remove Zarf Variables, Constants and Templates
  - (this may be desireable once the new design is solidified, but Zarf Variables enable some functionality that we may want to preserve (since they are template-driven))

## Proposal

The proposed solution is to add a new `values` global field to the Zarf package configuration that will accept a list of values files to serve as package defaults as well as an optional schema file for validating the values provided.  These fields would follow existing Zarf compose conventions and would map into Helm charts with a new `values` field under `charts`. The Zarf configuration itself would also change to allow Go templating of values in Zarf actions instead of being injected into the environment like Zarf Variables are today.

To set these values a new `package.deploy.values` configuration option would be added to the Viper config and a new `-f`/`--values` flag would be added to the CLI to allow values files to be specified on `zarf package deploy` and `zarf dev deploy`.  For now, the `--set` flag would remain as it is for Zarf Variables though eventually we may want to deprecate it in the future and align to the [Helm `--set` syntax](https://helm.sh/docs/intro/using_helm/#the-format-and-limitations-of---set) with the values specified setting Zarf Values instead of Zarf Variables.

### User Stories (Optional)

#### Story 1

**As** Jacquline **I want** to be able to set full Helm value objects in a Zarf config variable **so that** I can simplify package configurations and rely on my existing familiarity with Helm.

**Given** I have a Zarf Package with a Helm value override in a chart
```yaml
components:
  - name: component
    charts:
      - name: mychart
        version: 0.1.0
        namespace: zarf
        localPath: chart
        valuesFiles:
          - values.yaml
        values:
          - key: component.resources
            path: resources
```
**When** I deploy that package with a `zarf-config.yaml` like the below* or by specifying `-f values.yaml`:
```yaml
package:
  deploy:
    values:
      - values.yaml
```
**And** I have a `values.yaml` like the below:
```yaml
component:
  resources:
    limits:
      memory: 128Mi
      cpu: 100m
    requests:
      memory: 64Mi
      cpu: 100m

other-component:
  disabled: true
```
**Then** Zarf will apply the entire interface to the Helm chart override

> [!NOTE]
> *This would apply to all `zarf-config` formats not just YAML

### Risks and Mitigations

This will require some additional processing of the `zarf-config` files to allow them to be processed properly because of Viper's long-running [case-insensitive keys issue](https://github.com/spf13/viper/issues/1014).  This could be performed similar to UDS CLI's implementation of this feature and / or by deprecating some of the existing less-used formats for configuration with the Zarf team tentatively looking to deprecate and remove config file options besides yaml and toml given they are not used much in the community but still would take time to support.

The `--set` syntax will change somewhat how variables are interpreted on the CLI (i.e. `--set VAR=100` will no longer represent `"100"` and instead will just be `100` internally). For `###` templates this will not be a breaking change and can be mitigated by simply representing the value as a string for backwards compatibility.  Existing Helm Overrides in Zarf `charts` may experience breakages from this however and this should be noted upon release.  We could add additional opt-in flags but this would add complexity and is likely not desirable for the expected impact of this smaller break (given Helm overrides have not seen much use yet).

As we implement these changes there are risks around opening a `string` to an `interface{}` and we should strongly look at adopting many of the [Helm helpers](https://github.com/helm/helm/blob/main/pkg/chartutil/values.go#L71) from their `chartutil` package to ensure that potential security and stability issues are minimized.  Also called out below we should implement fuzz testing to catch unanticipated issues and provide an additional layer of assurance to the implementation.

## Design Details

The new `values.files` field would 

`values.files` and `schema` would follow the current component composability logic where additional values files from parent Zarf packages would layer in and merge with those of the children.  The schema from the parent would replace that of the child.

`setValues` 

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

For additional safety when implementing this feature we should look at adding fuzz testing to the existing unit / e2e tests to ensure that panics and other potential security issues are handled appropriately.  As mentioned above Helm does have some utility functions to help address some of these concerns but this would ensure that those functions were being used properly within Zarf and any additional concerns from Zarf's additional functionality were handled correctly as well.

##### Unit tests

Variable interfaces and libraries should be updated to ensure that interfaces are properly handled as opposed to strings.

##### e2e tests

Additional E2E tests should be added to ensure that `zarf-config` interfaces and Helm-style `--set`s are passed through appropriately to Helm on chart install / upgrade.

### Graduation Criteria

Pending review / community input these changes could be made either with known breakages (pending perceived impact of the --set changes) or as a feature flag for a series of releases eventually becoming the default behavior.

### Upgrade / Downgrade Strategy

NA - There would be no upgrade / downgrade of cluster installed components

### Version Skew Strategy

NA - This proposal doesn't impact how Zarf's components interact

## Implementation History

2025-02-03: Initial version of this document.

## Drawbacks

This will introduce a breakage that will need to be mitigated for existing users either expressly as a breaking change or with a feature flag.  It also pushes out/builds upon the existing Variables/Constants/Templates paradigm which has been in need of some rethinking for a while.  This doesn't preclude doing that work eventually but it does build upon / patch a paradigm that will likely need to be rethought.

## Alternatives

We could instead spend more effort redesigning the entire Variables/Constants/Templates paradigm to simplify this for Zarf users as mentioned in the drawbacks above.  This would likely look like collapsing Templates/Constants together and changing Variables to Values or something similar to wholly align them with Helm paradigms (since that is what many Zarf users would be familiar with).

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
