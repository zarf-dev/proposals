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

**As** Jacquline **I want** to be able to specify values files for a Zarf package **so that** I can simplify package configurations and rely on my existing familiarity with Helm.

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

This will introduce a wholly new way to input values into Zarf that will live alongside the existing Variables, Constants and Templates for a time.  This feature will need to be clearly disambiguated from those features in documentation and if the feature gains traction and is accepted by the community a deprecation plan for the original Zarf Variables, Constants and Templates should be created.  Because this feature will be implemented alongside the existing featureset it should not introduce many breaking changes, though to assist with disambiguation the feature to map Zarf Variables to Helm Values should be deprecated and removed in favor of the new Zarf Values mapping.

This feature also could open up Zarf packages to being less declarative - especially if a package author opens up security-critical Helm values in their charts.  This should be clearly documented as a concern which should also recommend a policy engine to be used to enforce security-critical values within the cluster itself.

## Design Details

The new `values.files` field would be added to the `ZarfPackageConfig` schema and would accept a list of local values files or URLs to match the existing functionality of the `valuesFiles` key under `charts`.  This key would also follow the same composability logic as the `valuesFiles` key with additional parent values files merging with and overriding any common keys from children.  Since this is a global field, values would be merged regardless of component similar to how variables work today. The `values.schema` field would simply replace the child version if it were non-empty in the parent, similar to the `namespace` or `releaseName` fields in `charts` today.

Once all of the values (from the defaults in the `ZarfPackageConfig` and the set values files) are merged, the values would be passed to the Helm chart via the `values` field under a given `charts` entry.

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

For additional safety when implementing this feature we should look at adding fuzz testing to the existing unit / e2e tests to ensure that panics and other potential security issues are handled appropriately.  As mentioned above Helm does have some utility functions to help address some of these concerns but this would ensure that those functions were being used properly within Zarf and any additional concerns from Zarf's additional functionality were handled correctly as well.

##### Unit tests

Variable interfaces and libraries should be updated to ensure that interfaces are properly handled as opposed to strings.

##### e2e tests

Additional E2E tests should be added to ensure that `zarf-config` interfaces and `-f`/`--values` are passed through appropriately to Helm on chart install / upgrade.

### Graduation Criteria

Pending review / community input these changes could be made either with known breakages (pending perceived impact of the --set changes) or as a feature flag for a series of releases eventually becoming the default behavior.

### Upgrade / Downgrade Strategy

NA - There would be no upgrade / downgrade of cluster installed components

### Version Skew Strategy

This proposal doesn't impact how Zarf's Agent and CLI interact so no changes would be needed there. If a package that contained the `values` field was deployed with an older version of the Zarf CLI the values would simply be ignored.

## Implementation History

2025-03-31: Initial version of this document.

## Drawbacks

This will require a lot more design work to ensure that this new feature has a solid user experience and is well integrated with the rest of Zarf's features.

## Alternatives

We could patch the existing Variables/Constants/Templates paradigm to align more with Helm paradigms but this would not address the other issues that exist with the Variables/Constants/Templates paradigm.  Features like `autoIndent` and `setVariables` have always been limiting to users and creating a new way to set values will allow us to design something that is more user friendly (especially since many Zarf users are also Helm users).

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
