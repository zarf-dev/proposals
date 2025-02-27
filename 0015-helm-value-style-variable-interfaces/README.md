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

# ZEP-0015: Helm Value Style Variable Interfaces

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

This ZEP proposes to expand variables to more than just `string` values and instead to accept an `interface{}` instead.  This would change `map[string]string` for variables into `map[string]interface{}` in addition to changing how variables are inputted and handled internally.

## Motivation

The motivation for this centers around aligning Zarf Variables closer to Helm Values which use the type `map[string]interface{}` over `map[string]string`.  Pull Request [#2132](https://github.com/zarf-dev/zarf/pull/2131) pulled this point more into focus by allowing Zarf variables to be directly passed as Helm values, and there has been desire from the Zarf community to treat Zarf Variables more like Helm Values [[1](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1706175082741539)], [[2](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1702400472208839)], [[3](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1706299407255849?thread_ts=1706175134.116329&cid=C03B6BJAUJ3)].

### Goals

- Simplify the use of Zarf Variables for members of the community familiar with Helm

### Non-Goals

- Entirely redesign Zarf Variables, Constants and Templates
  - (while this is a non-goal for _this_ proposal it is worth discussion if this would better serve the goal above)

## Proposal

The proposed solution is to change the internal tracking of Zarf Variables from a `map[string]string` to a `map[string]interface{}` as well as allowing the `map[string]interface{}` to be loaded from a `zarf-config` file (including all of the existing formats that the file accepts).

Additionally `--set` on the CLI would align to the [Helm `--set` syntax](https://helm.sh/docs/intro/using_helm/#the-format-and-limitations-of---set) with the values on the right-hand side of the `=` being handled as described there instead of as a `string`.

### User Stories (Optional)

#### Story 1

**As** Jacquline **I want** to be able to set full Helm value objects in a Zarf config variable **so that** I can simplify package configurations and rely on my existing familiarity with Helm.

**Given** I have a Zarf Package with a Helm value override in a chart
**When** I deploy that package with a `zarf-config.yaml` like the below*:
```yaml
package:
  deploy:
    set:
      MY_OVERRIDE_VARIABLE_NAME:
        key: value
        bool: true
        num: 100
        arr: []
        obj: {}
```
**Then** Zarf will apply the entire interface to the Helm chart override

> [!NOTE]
> *This would apply to all `zarf-config` formats not just YAML

#### Story 2

**As** Jacquline **I want** to be able to set full Helm value objects from the CLI **so that** I can simplify package configurations and rely on my existing familiarity with Helm.

**Given** I have a Zarf Package with a Helm value override in a chart
**When** I deploy that package with a `--set` like the below:
```yaml
zarf package deploy zarf-package-test.tar.zst --set MY_OVERRIDE_VARIABLE_NAME.key=value,MY_OVERRIDE_VARIABLE_NAME.bool=true,MY_OVERRIDE_VARIABLE_NAME.num=100,MY_OVERRIDE_VARIABLE_NAME.arr=[]
```
**Then** Zarf will apply the entire interface to the Helm chart override

> [!NOTE]
> *This should follow the [Helm `--set` syntax](https://helm.sh/docs/intro/using_helm/#the-format-and-limitations-of---set)

### Risks and Mitigations

This will require some additional processing of the `zarf-config` files to allow them to be processed properly because of Viper's long-running [case-insensitive keys issue](https://github.com/spf13/viper/issues/1014).  This could be performed similar to UDS CLI's implementation of this feature and / or by deprecating some of the existing less-used formats for configuration with the Zarf team tentatively looking to deprecate and remove config file options besides yaml and toml given they are not used much in the community but still would take time to support.

The `--set` syntax will change somewhat how variables are interpreted on the CLI (i.e. `--set VAR=100` will no longer represent `"100"` and instead will just be `100` internally). For `###` templates this will not be a breaking change and can be mitigated by simply representing the value as a string for backwards compatibility.  Existing Helm Overrides in Zarf `charts` may experience breakages from this however and this should be noted upon release.  We could add additional opt-in flags but this would add complexity and is likely not desirable for the expected impact of this smaller break (given Helm overrides have not seen much use yet).

As we implement these changes there are risks around opening a `string` to an `interface{}` and we should strongly look at adopting many of the [Helm helpers](https://github.com/helm/helm/blob/main/pkg/chartutil/values.go#L71) from their `chartutil` package to ensure that potential security and stability issues are minimized.  Also called out below we should implement fuzz testing to catch unanticipated issues and provide an additional layer of assurance to the implementation.

## Design Details

This design proposal seeks to keep changes to a minimum to align Zarf Variables with Helm Values with the largest changes being the change from `string` to `interface` and the changes to loading a `zarf-config` file and handling `--set` as described above.

This change will also impact other aspects of Zarf variables as described below:

- `default` values would now accept an interface - this would affect packages in the same way as zarf-config values for backwards compatibility
- `setVariables` would still set variables to strings with no additional processing - if more is desired here later (i.e. expanding the `type` field) that would be a separate ZEP

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
