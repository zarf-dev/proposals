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

This ZEP proposes to introduce a new way to provide configuration to Zarf Packages that is more in line with the Helm values paradigm.  This proposal would provide a new settable interface that uses `map[string]interface{}` instead of the current variable interface of `map[string]string`.  These values would also be able to be mapped directly to Helm chart values, and would interact with other Zarf features where Zarf Variables are used today through Go templating (i.e. `actions`, `manifests` and `files`).

This ZEP supercedes [ZEP-0015](0015-helm-value-style-variable-interfaces/README.md).

## Motivation

The motivation for this centers around the long-lived desire to have Zarf Variables work more like Helm Values, which users in the Zarf community are generally more familiar with.  This proposal seeks to rethink what Variables should be given that desire and take the chance to also improve the user experience around value configuration within Zarf.  This proposal would allow values to be a full `interface{}` rather than simply `string`s and would also allow those values to be mapped directly to Helm chart values and templated in Zarf actions in a more Helm-like way. Pull Request [#2132](https://github.com/zarf-dev/zarf/pull/2131) initially allowed Zarf Variables to be directly passed as Helm values, but this always had limitations due to the design of Zarf Variables being geared toward string templating.  Over time, there has been desire from the Zarf community to treat Zarf Variables more like Helm Values [[1](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1706175082741539)], [[2](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1702400472208839)], [[3](https://kubernetes.slack.com/archives/C03B6BJAUJ3/p1706299407255849?thread_ts=1706175134.116329&cid=C03B6BJAUJ3)] and reevaluating variables entirely allows us to define more clearly how these values should work across Zarf's featureset (including non-Helm features like `actions` and `files`).

### Goals

- Design a replacement for Zarf Variables for members of the community familiar with Helm
- Provide more flexibility for what values can be and how they can be used
- Integrate this new design to work across the features where Zarf Variables work today

### Non-Goals

- Entirely remove Zarf Variables, Constants and Templates
  - (this may be desireable once the new design is solidified, but Zarf Variables enable some functionality that we may want to preserve (since they are template-driven))

## Proposal

The proposed solution is to add a new `values` global field to the Zarf package configuration that will accept a list of values files to serve as package defaults as well as an optional schema file for validating the values provided.  These fields would follow existing Zarf compose conventions and would also map into Helm charts with a new `values` field under `charts`. The Zarf configuration itself would also change to allow Go templating of values in Zarf actions instead of being injected into the environment like Zarf Variables are today.  Zarf `files` and `manifests` would also optionally implement Go templating to be able to take advantage of values as well.

To set these values new `package.[deploy|remove].values` configuration options would be added to the Viper config and a new `-f`/`--values` flag would be added to the CLI to allow values files to be specified on `zarf package deploy`, `zarf package remove` and `zarf dev deploy`.  For now, the `--set` flag would remain as it is for Zarf Variables though eventually we may want to deprecate it and align to the [Helm `--set` syntax](https://helm.sh/docs/intro/using_helm/#the-format-and-limitations-of---set) with the values specified setting Zarf Values instead of Zarf Variables.  Zarf `actions` would also add a new `setValues` field that would allow setting values from an action similar to `setVariables`.

### User Stories (Optional)

#### Story 1

**As** a platform engineer **I want** to be able to specify values files for a Zarf package **so that** I can simplify package configurations and rely on my existing familiarity with Helm.

**Given** I have a Zarf Package with a Helm value override in a chart
```yaml
# zarf.yaml
components:
  - name: my-component
    charts:
      - name: mychart
        version: 0.1.0
        namespace: zarf
        localPath: chart
        valuesFiles:
          - values.yaml
        values:
          # option 1: 
          - sourcePath: my-component.resources
            targetPath: resources
          # option 2: 
          - my-component.resources: resources
          # option 3:
          my-component:
            resources: resources
```
**When** I deploy that package with a `zarf-config.yaml` like the below* or by specifying `-f values.yaml`:
```yaml
# zarf-config.yaml
package:
  deploy:
    values:
      - values.yaml
```
**And** I have a `values.yaml` like the below:
```yaml
# values.yaml
my-component:
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

---

**Given** I have a Zarf Package with top-level `values` and Go templating inside an action
```yaml
# zarf.yaml
values:
  files:
    - values-defaults.yaml
  schema: values.schema.json

components:
  - name: my-component
    actions:
      onDeploy:
        before:
          - cmd: "echo \"{{ .Values.my-component.resources.limits.memory }}\""
```
**And** it was created with the following `values-defaults.yaml` file:
```yaml
# values-defaults.yaml
my-component:
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
**When** I deploy that package without setting any values
**Then** Zarf will template the action and output `128Mi`

---

**Given** I have a Zarf Package with a setValues action
```yaml
# zarf.yaml
components:
  - name: my-component
    actions:
      onDeploy:
        before:
          - cmd: "echo '{ \"memory\": \"256Mi\", \"cpu\": \"200m\" }'"
            setValues:
              - path: my-component.resources.limits
                type: json
                
```
**When** I deploy that package with a `zarf-config.yaml` like the below* or by specifying `-f values.yaml`:
```yaml
# zarf-config.yaml
package:
  deploy:
    values:
      - values.yaml
```
**And** I have a `values.yaml` like the below:
```yaml
# values-defaults.yaml
my-component:
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
**Then** Zarf will initially use the values provided for any components before the one with `setValues`
**And Then** apply the values provided in the `setValues` action for any following components based on the actions lifecycle.

> [!NOTE]
> *This would apply to all `zarf-config` formats not just YAML

---

**Given** I have a Zarf Package with top-level `values` and Go templating inside of a file and/or manifest (with templating enabled)
```yaml
# zarf.yaml
values:
  files:
    - values-defaults.yaml
  schema: values.schema.json

components:
  - name: my-component
    manifests:
      - name: my-deployment
        namespace: my-namespace
        files:
          - my-deployment.yaml
        template: true
    files:
      - source: my-deployment.yaml
        target: my-out-deployment.yaml
        template: true
```
```yaml
# my-deployment.yaml
kind: Deployment
metadata:
  name: my-deployment
  namespace: my-namespace
spec:
  template:
    spec:
      containers:
        - name: my-container
          resources:
            {{ .Values.my-component.resources | toYaml }}
```
**And** it was created with the following `values-defaults.yaml` file:
```yaml
# values-defaults.yaml
my-component:
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
**When** I deploy that package without setting any values
**Then** Zarf will template the file and manifest with the resources given

### Risks and Mitigations

This will introduce a wholly new way to input values into Zarf that will live alongside the existing Variables, Constants and Templates for now.  Because of this, the feature will need to be clearly disambiguated from Variables/Constants/Templates in documentation and, while this feature should not introduce many breaking changes being implemented alongside the existing featureset, the feature to map Zarf Variables to Helm Values should be deprecated and removed in favor of the new Zarf Values mapping to assist with disambiguation.  If the feature gains traction and is accepted by the community, a deprecation plan for the original Zarf Variables/Constants/Templates should be created.  Likely this plan would not break `charts.variables` in existing packages and would simply prevent furutre packages from using this feature.

This feature also could open up Zarf packages to being less declarative - especially if a package author opens up security-critical Helm values in their charts.  This caveat should be clearly documented as a concern which should also recommend a policy engine be used to enforce security-critical values within the cluster itself.

Because we will be using more `interface{}` types, we should also look into the security implications of this feature and ensure that this is well tested and that we utilize some of Helm's existing protections against `nil` maps and other potential security issues with this feature.

This proposal also adds to the concept of Zarf `onDeploy` actions and creates another way to execute arbitrary bash commands on the host (depending on how the package creator implemented Zarf Values and their Go templates).  If this feature is to replace Zarf variables however, using Values in actions is still needed, and examples exist in the wild where Helm templates alone are not sufficient to provide the desired functionality for a package.  One example being the GitLab Runner UDS Package that creates a runner token through the GitHub API - this requires pulling a registration token from an existing secret (which is possible today with Helm templates), but then this token is used to register the runner with the GitLab API.  This requires making an HTTP request which Helm cannot help with requiring onDeploy actions to wire this in. References: [GitLab Runner Config Chart Values](https://github.com/defenseunicorns/uds-package-gitlab-runner/blob/d2b573bdbed12ac2aafd52082f1b9ea84b213439/chart/values.yaml#L9), [GitLab Runner Token `onDeploy` action](https://github.com/defenseunicorns/uds-package-gitlab-runner/blob/d2b573bdbed12ac2aafd52082f1b9ea84b213439/common/zarf.yaml#L34).  This will need to be mitigated with documentation and it may be desireable to implement a form of `shellcheck` to `zarf dev lint` to look for areas where this might be an issue.  Users would be able to control the shape of input values via the `values.schema` field and Zarf should halt a deployment if a bad value is provided.  Users could also pass user input through the `env` field in actions for some additional protection.

Zarf `files` and `manifests` may already contain Go templates that we would not want to trample on. Previous behavior wrapped `manifests` in an additional `{{ ... }}` so that Helm would not erroneously see the templates inside a manifest and mess with them [[Ref](https://github.com/zarf-dev/zarf/blob/e51ea928f58e24d2558e679d1905254b1f3ae7cd/src/internal/packager/helm/common.go#L107)]. To mitigate this we could change template delims (i.e. `###{{ .Values.hello }}###`) but this may not be familiar to Helm users and it may be useful to be able to have a longer term toggle for templating in these use cases.  To solve this this proposal would add a `template: true` field to `files` and `manifests` (defaulting to `false` to start but this can change as the feature gets maturity)

## Design Details

The new `values.files` field would be added to the `ZarfPackageConfig` schema and would accept a list of local values files or URLs, matching the existing functionality of the `valuesFiles` key under `charts`.  This key would also follow the same composability logic as the `valuesFiles` key with additional parent values files merging with and overriding any common keys from children imports.  Since this is a global field, values would be merged regardless of component, similar to how variables work today.  The `values.schema` field would simply replace the child version if it were non-empty in the parent, similar to the `namespace` or `releaseName` fields in `charts` today.

Each of the referenced `values.files` would be included inside the created Zarf package and stored inside of the tarball or at an OCI path so that they would travel with the package and be available to the user on deploy.  On deploy, the values (from the defaults in the `ZarfPackageConfig` and the user set values files) would be merged with set values overriding the defaults.  The resulting values would be passed to the Helm chart via the `values` field under a given `charts` entry and any actions which contained a go template (i.e. `{{ .Values.component.resources | toYaml }}`) would be templated prior to execution. This templating would work in all fields within an action definition including `cmd` and `wait` actions.

This feature would also implement a `setValues` field for actions that would act as a replacement for the existing `setVariables` field.  This field would take a values path and would set that path to the output of the command.  This field would also have `type` defined on it, though instead of `file`, it would take `string`, `json`, or `yaml` and then handle the standard output of the command according to that format.

Zarf Values would also be available to onRemove actions if they existed in `package.remove.values` or were set with `-f`/`--values`.  These would template actions and be able to use `setValues` but would not map to Helm charts.

An additional `zarf dev generate-values` command will also be added to generate a sample values file based on the merging of the default values specified in the `ZarfPackageConfig` of a given package.  This will assist deploy users in knowing what values they are able to set within a given package.

Zarf Values files themselves will only accept `yaml` as a format.  The list of values files in a zarf config file under `package.[deploy|remove].values` will exist in all of Zarf's config formats, but the values files that the list references will only support the YAML format.

Zarf Values will _not_ be available to onCreate actions since Zarf Variables are not available there either and there may be confusion around when a value is set (`onCreate` or `onDeploy`).

If multiple Zarf values use the same target Helm value path in a chart, the last value set will be the one that is used.  This is similar to how Zarf compose rules for `valuesFiles` work today. i.e.

```yaml
components:
  - name: my-component
    charts:
      - name: mychart
        ...
        values:
          - sourcePath: my-component.resources
            targetPath: resources
          - sourcePath: other-component.resources
            targetPath: resources # this wins
```

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

As mentioned above, for additional safety when implementing the elements of this feature that interact with the `map[string]interface{}` types, we should look at adding fuzz testing to the existing unit / e2e tests to ensure that panics and other potential security issues are handled appropriately.  Helm does have some utility functions to help address some of these concerns but this would ensure that those functions were being used properly within Zarf and any additional concerns from Zarf's additional functionality were handled correctly as well.

##### Unit tests

Values interfaces and libraries should be updated to ensure that interfaces are properly passed to charts and templated in actions.

##### e2e tests

Additional E2E tests should be added to ensure that `zarf-config` values files and `-f`/`--values` are passed through appropriately to Helm on chart install / upgrade.

### Graduation Criteria

Pending review / community input these changes could be replace the existing Variables/Constants/Templates with this new Values feature.  This would require evaluating the features still in use by the community and creating suitable alternatives for them in either the values themselves or the Go templating of the Zarf package definition.

### Upgrade / Downgrade Strategy

NA - There would be no upgrade / downgrade of cluster installed components

### Version Skew Strategy

This proposal doesn't impact how Zarf's Agent and CLI interact so no changes would be needed there, however if a package that contained the `values` field was deployed with an older version of the Zarf CLI the values would simply be ignored.

## Implementation History

2025-03-31: Initial version of this document.

## Drawbacks

This feature will require a lot of design work to ensure that it has a solid user experience and is well integrated with the rest of Zarf's features.  Go templating is also a large change that was avoided initially to reduce potential conflicts with Helm's templating - while relatively safe to use in the Zarf Package definition it would be difficult to extend this templating further down if that were desired.

## Alternatives

We could patch the existing Variables/Constants/Templates paradigm to align more with Helm paradigms but this would not address the other issues that exist with Variables/Constants/Templates.  Features like `autoIndent` and `setVariables` have always been limiting to users and creating a new way to set values will allow us to design something that is more user friendly (especially since many Zarf users are also Helm users).

## Infrastructure Needed (Optional)

NA - This change requires no additional infrastructure as it is internal to Zarf's operation.
