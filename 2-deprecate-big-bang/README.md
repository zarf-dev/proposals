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
- [X] **Fill out this file as best you can.**
  Focus on the "Summary" and "Motivation" sections first. If you've already discussed
  the idea with the Technical Steering Committee, this part should be easier.
- [X] **Create a PR for this ZEP.**
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

# ZEP-1: Deprecate Big Bang Extension

- [Summary](#summary)
- [Motivation](#motivation)
  - [Goals](#goals)
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

## Summary

This ZEP proposes removing the Big Bang extension in Zarf. The Big Bang extension simplifies deploying the [Big Bang platform](https://p1.dso.mil/services/big-bang), but it adds complexity to the codebase. Zarf will focus on solving the general air-gap Kubernetes problem rather than a Department of Defense (DoD) platform specific use case. The proposal introduces a new repo, `generate-big-bang-zarf-package`, which will contain a tool that generates a zarf.yaml file with the necessary components for Big Bang deployment.

## Motivation

### Background - why does the Big Bang extension exist

Zarf simplifies Kubernetes deployments in the air-gap. The initial Zarf use cases targeted air-gapped environments within the DoD. The creators of Zarf were heavily involved in creating the DoD platform, Big Bang. 

Big Bang is a helm chart of helm charts. The default Big Bang deployment contains forty different images across fourteen different repositories. A user of Zarf creating a package to deploy Big Bang without the Big Bang extension needs to include every image and git repository within Big Bang. Each image and repo has a version that changes with each Big Bang release, and images & repos change depending on the helm values. Manually finding these versions every release is time-consuming and arduous.

To simplify the deployment of Big Bang within Zarf the Big Bang extension was created. This allowed users to create their Big Bang component in as little as three lines of code - version is the only required field. Below is the extensions key with a value for every sub-key.

```yaml
extensions:
bigbang:
    version: 2.34.0
      skipFlux: false
    repo: https://repo1.dso.mil/big-bang/bigbang.git
    FluxPatchFiles:
      - config/flux-patch.yaml
    valuesFiles:
      - config/values.yaml
```

During `zarf package create` the Big Bang extension is processed so that the created package includes everything needed for a Big Bang deployment. The zarf.yaml within the package tar contains a component with all the necessary images, repos, manifests and actions. To view the full component created by the Big Bang extension see [#2875](https://github.com/zarf-dev/zarf/issues/2875).

### Problems with the Big Bang extension

The extension does simplify deploying Big Bang in Zarf, but this comes with downsides.

- The extension must be considered during package create, skeleton package create, and component compose increasing complexity in the codebase.
- Zarf performs implicit actions that are not visible to the user, such as adding a custom values file with Kyverno policies.
- Big Bang relies on images stored in [Iron Bank](https://p1.dso.mil/services/iron-bank), a container registry run by the DoD. Unfortunately, Iron Bank has frequent outages which causes flakes in Zarf’s test suite.

A notable reason to move away from offering Big Bang support in Zarf is to focus on being an air-gap Kubernetes tool, rather than a DoD deployment tool. The Big Bang support exists because Zarf was created by the DoD contractor Defense Unicorns. Now Zarf has been donated to Open SSF and it should focus on what is best for the community at large.


### Goals

- Remove all references to Big Bang within Zarf.

## Proposal

The proposed solution is to create a new go project, [defenseunicorns/generate-big-bang-zarf-package](https://github.com/defenseunicorns/generate-big-bang-zarf-package), focused solely on generating a big bang package. The command `generate-big-bang-zarf-package` will accept have one argument, version, and accept the following command line flags: `values-file-manifests`, `skipFlux`, and `repo`. The `fluxPatchFiles` flag will not have an equivalent, however, users will be able to edit their Flux component to deploy it with patch files after the generate command has run. Below is the helper text for the command: 

```bash
Generates a zarf.yaml file and the associated manifests necessary to create a Zarf package that deploys Big Bang 
Usage:
  generate-big-bang-zarf-package [ Version ] [flags]

Examples:
zarf dev generate big bang --version 2.3.4 –skip-flux=false --values-file-manifests =istio-values.yaml

Flags:
  --skip-flux bool           Whether or not to create a flux component (default false)
  --repo string   	         Override repo to pull Big Bang from instead of Repo One.
  --values-file-manifests    A comma separated list of configmap or secret manifests to pass to the Big Bang Helm Release

```

This command will create a zarf.yaml file with a component that fully deploys Big Bang for the provided version. Zarf will generate manifests and place them in a manifests folder. Any files submitted with the `values-files-manifests` flag will be added to the manifests list as well. The component will follow the following structure: 

```
- Manifests
  - Flux Git Source Manifest
  - Flux Helm Release Manifest
  - Zarf custom values file
  - User submitted custom values files
- Images
  - All the Big Bang images
- Git Repositories
  - All the Big Bang repos
- Health checks
  - A health check for each Helm Release
```

If a zarf.yaml file already exists `generate-big-bang-zarf-package` will instead generate a file called `zarf-<uid>.yaml`. This will prevent Zarf from overwriting any existing Zarf files while providing a convenient way to copy and paste images and repos from one file to the other.

### User Stories (Optional)

#### Story 1

A user wants to deploy Big Bang with Zarf. They run `generate-big-bang-zarf-package 2.34.0` without an existing `zarf.yaml` and get the necessary manifests and zarf.yaml so that they are ready to run `zarf package create`.

#### Story 2

An existing deployer of Big Bang with Zarf wants to update to a new version of Big Bang. They already have a `zarf.yaml` that deploys Big Bang alongside other components - shown below. They run `generate-big-bang-zarf-package 2.34.0` which creates a `bigbang-zarf.yaml` so that their existing `zarf.yaml` is not overridden. The user then copies and pastes the images and repos from the `bigbang-zarf.yaml` to their existing `zarf.yaml`.

```yaml
- name: BigBang
  description: Big Bang component - not showing the other keys for brevity
- name: my-app
  description: My app that is deployed on top of Big Bang
  manifests:
    my-manifest.yaml
  images:
    - my-image:latest
```

### Risks and Mitigations

This deprecation removes and simplifies the Zarf package create flow by removing extensions. There are no security implications.

The proposed UX will be reviewed and tested by real users of the Big Bang extension to ensure it provides a suitable replacement.

## Design Details

### User Values File Changes
When a user runs `generate-big-bang-zarf-package 2.34.0 –values-file-manifests=my-values-file.yaml` the values files are expected to be Kubernetes secrets manifests to work with Flux Helm Releases. These manifests will both be added to the Zarf Package and referenced in the generated Helm Release under the [valuesFrom](https://fluxcd.io/flux/components/helm/helmreleases/#values-references) key. This differs from the Big Bang extension which expects traditional Helm values files and converts them to Kubernetes secret manifests during zarf package create. If a user submits a values file that is not in the below format, the generate command will fail and they will be prompted to fix. 

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: bb-neuvector-vals
  namespace: bigbang
stringData:
  values.yaml: |
  neuvector:
  	values:
    	k3s:
       enabled: true
```

### Flux component

The Big Bang extension deploys Flux within the bigbang component. The generate command will instead create a zarf.yaml with separate components for flux and bigbang so the deployments are clearly differentiated.

### Subsequent Generate Big Bang runs 

When a `zarf.yaml` file already exists in the current working directory `zarf dev generate bigbang 2.34.0` will create a file called `zarf-<uid>.yaml` to not delete user configuration on their `zarf.yaml` file. Existing manifests (Flux Git Repository, Flux Helm Repository, and Zarf credentials values manifest) will be replaced since user updates to these files are not anticipated, and were not possible with the Big Bang extension. 

### Test Plan

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

##### Prerequisite testing updates

NA

##### Unit tests

There must be unit tests to ensure the command generates the expected Zarf package. These tests must be run with and without values files to confirm the generated package reflects the requested Big Bang release. 

##### e2e tests

NA

### Graduation Criteria

As this feature is being moved into it's own repository it will not have a typical graduation criteria. The release process will be as simple as possible. There will be no tags created on the repository. The tool is intended to be installed with `go install github.com/defenseunicorns-partnerships/generate-big-bang-zarf-package@latest`. Since `go install @latest` grabs the default branch, a new release occurs every time a PR is merged to main.

### Upgrade / Downgrade Strategy

NA

### Version Skew Strategy

NA

## Implementation History

<!--
Major milestones in the lifecycle of a ZEP should be tracked in this section.
Major milestones might include:
- the `Summary` and `Motivation` sections being merged, signaling acceptance of the ZEP
- the `Proposal` section being merged, signaling agreement on a proposed design
- the date implementation started
- the first Kubernetes release where an initial version of the ZEP was available
- the version of Kubernetes where the ZEP graduated to general availability
- when the ZEP was retired or superseded
-->

## Drawbacks

A separate git repository increases administrative burden such as upgrading dependencies, creating the CI and release process, and managing permissions. This is somewhat mitigated by the simple release process we planned in [Graduation Criteria](#Graduation-Criteria). Additionally, as this functionality is moved out of Zarf organization, we expect new feature work to be driven by the community. 

## Alternatives

### Build more functionality into the find-images command

Zarf could enhance the `zarf dev find-images` functionality to work with [Flux Helm Releases](https://fluxcd.io/flux/components/helm/helmreleases/). This would help with the most difficult part of updating Big Bang manually - finding all of the images. Finding images using Helm Releases is not always possible. Helm Releases are specified using the `sourceRef` key - example below. This key points to data in the cluster since find-images is run on manifests, not on cluster resources, the `sourceRef` data is not available.

Zarf could scan each manifest for potential Flux Helm sources, then match those sources with the Helm Releases. This would have to be done recursively since those Helm Releases could point to other Helm Releases. This strategy is similar to how the Big Bang extension works, except the Big Bang extension gets to use hard coded values for the Big Bang Helm source.

We considered this option but decided not to go with it for a few reasons. First, the above strategy would not be 100% reliable since there’s no guarantee users will deploy their sources and releases in the same component. Additionally, there have been no requests from the community to use Flux Helm Releases to find images. The general solution is more complex since Zarf can't rely on hard coding Big Bang values or assume that the Flux Helm Sources will be Git sources. Lastly, this solution does not solve finding repos so Zarf would either lose that functionality or we would have to introduce another new feature.

```yaml
apiVersion: helm.toolkit.fluxcd.io/v2beta1
kind: HelmRelease
metadata:
 name: generic-helm-release
spec:
 chart:
   spec:
 	chart: ./chart
 	sourceRef:
   	kind: GitRepository
   	name: my-git-repo
```

### Deprecate the Big Bang extension without an alternative

While the Big Bang extension is not part of the core functionality of Zarf we wanted to avoid alienating users by not providing an alternative. Given the low level of effort expected to maintain the `generate-big-bang-zarf-package` repository, we believe the proposed solution to be a good compromise. 

We estimated a low level of effort based on the relative infrequency and size of changes throughout the history of the [src/extensions/bigbang](https://github.com/zarf-dev/zarf/commits/main/src/extensions/bigbang) folder. The proposed changes will make maintenance easier still. 