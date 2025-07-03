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

# ZEP-0032: Initial IPv6 support

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

Currently a container runtime interface (CRI) connects to the default Zarf registry using a nodeport service on 127.0.0.1. Connecting to nodeport services on localhost is blocked by certain distros, IPV6, and NFTables. Zarf will introduce a hostNetwork / hostPort proxy daemonset to enable the default registry for these use cases.   

This ZEP proposes to implement support in Zarf for deploying into Kubernetes clusters configured with IPv6 networking and where IPv4 is not supported.

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

Kubeneretes 1.33 has made [NFTables](https://kubernetes.io/blog/2025/02/28/nftables-kube-proxy/) generally available. The NFTables designers have made the explicit choice to stop making Nodeport services accessible one 127.0.0.1 (https://kubernetes.io/docs/reference/networking/virtual-ips/#migrating-from-iptables-mode-to-nftables). NFtables are not on by default, however we can expect distros, especially secure or performance focused distros, to start adopting NFTables by defualt in the coming months or years. It's important for the Zarf registry to work by default in these distros. 

The current nodeport service solution does not work with IPV6. There is a mandate ([wayback machine link because white house site is flaky ATM](https://web.archive.org/web/20250116092323/https://www.whitehouse.gov/wp-content/uploads/2020/11/M-21-07.pdf)) for government agencys to migrate to IPv6 single stack by end of fiscal year (FY) 2025. Given how often Zarf is used in government environments it's important IPV6 is enabled. 

Zarf does not work by default on certain distros such as talso and OpenShift (I need to verify openshift works with IPV6)


### Goals


* Implement a short-term solution for using Zarf with its internal "seed and long-term" container image registries in an IPv6-only Kubernetes cluster.

### Non-Goals

<!--
What is out of scope for this ZEP? Listing non-goals helps to focus discussion
and make progress.
-->

* Remove current mechanism for bootstrapping Zarf in IPv4-only or IPv4/IPv6 dual-stack clusters (that is, using a `Service` of type `NodePort` and the "the route localnet hack"). At least in the short term

## Proposal

<!--
This is where you explain the specifics of the proposal. Provide enough detail
for reviewers to clearly understand what you're proposing, but avoid including
too many specifics like API designs or implementation details. Focus on the
desired outcome and how success will be measured. The "Design Details" section
below is for the real nitty-gritty.
-->

Introduce a new flag called `--registry-proxy` will be added to `zarf init` which changes how Zarf connects to the registry. When `--registry-proxy` is used Zarf will replace the nodeport service with a clusterIP serivce and a `DaemonSet` running a proxy on each node to forward the registry. The proxy will use `hostIP` and `hostPort` in IPV4 and dual IP stacks, and hostNetwork in IPV6 only clusters. The `DaemonSet` will have to run for both the injector and long lived registry. 

A user can run `--registry-proxy` during `zarf init` and their choice will be saved to the cluster and used on subsequent runs on `init`. If a user wants to switch back to the localhost nodeport solution they must run `zarf init --registry-proxy=false`. If a user runs `zarf init` without the `--registry-proxy` flag on an already initalized cluster, it will keep using the registry connect method that the cluster is currently using, whether that is the registry proxy or nodeport solution. 

In an IPv6-only Kubernetes cluster it is currently not possible to use a `Service` of type `NodePort` to expose the Zarf internal container image registries via the IPv6 loopback.

This means that during the initialization of a cluster, Zarf should use another mean to expose the Zarf injector registry; the proposal here is to deploy the Zarf injector registry using a `DaemonSet` resource and use the `HostNetwork` functionality in order to allow for access to the Kubertnetes container runtime on each cluster node.

After that, Zarf should deploy its "long term" registry. A basic socket proxy component is added and deployed again to all nodes via a `Daemonset` resource. It also makes use of the `HostNetwork` functionality; this allows to access the "long term" registry on all nodes. This proxy component image should be available from the Zarf injector registry.


### User Stories (Optional)

<!--
Detail the things that people will be able to do if this ZEP is implemented.
Include as much detail as possible so that people can understand the "how" of
the system. The goal here is to make this feel real for users without getting
bogged down.
-->

#### Story 1

As a administrator of a Kubernetes cluster configured in IPv6-only networking mode, I want to deploy the Zarf init package using the default in cluster registry so I run the `zarf init` with the `--registry-proxy` command line flag.

#### Story 2

As a administrator of a Kubernetes cluster I want to move to NFTable for the performance and security improvements. To enable NFTables in my Zarf cluster I run `zarf init --registry-proxy`

### Risks and Mitigations

<!--
What are the risks of this proposal, and how do we mitigate? Think broadly.
For example, consider both security and how this will impact the larger
Zarf ecosystem.

How will security be reviewed, and by whom?

How will UX be reviewed, and by whom?
-->

As the proxying workload uses a `DaemonSet` and the host networking stack, it should:
* include the minimal amount of binaries to prevent shell access.
* expose the registry only locally (similar to what the `NodePort` exposes).

Security risks:
- Network policies are not considered in the hostIP setup so if someone wanted to block certain namespaces from the Zarf registry they would no longer be able to. It will exist on every pod. FIXME: I need to verify this. 
- Increased attack vector, if someone were to gain access to the pod in the daemonset, they could break the registry or potentially forward malicious content. 
- Decreased monitoring: The daemonset pod will not be monitored by a sidecar / istio. FIXME: I need to verify this. Also does the isitio host mode change this?

- HostIP and hostPort can be used in the daemonset on IPV4, which will limit the connections to only those on the actual node, however since IPV6 does not support rewriting packets to ::1 then the hostPort strategy will not work. IPv6 will have to use hostNetwork instead. HostNetwork comes with a greater security risk as anyone who can connect to the node will have connection to the registry, this differs from hostPort where the call must come from the node itself.


## Design Details

<!--
This section should contain enough information that the specifics of your
change are understandable. This may include API specs (though not always
required) or even code snippets. If there's any ambiguity about HOW your
proposal will be implemented, this is the place to discuss that.
-->
Initially the proxying component will be based on an existing container image having the `socat` binary ([Alpine socat](https://hub.docker.com/r/alpine/socat)); this is a small and simple image.

`zarf init` should fail if both `--registry-proxy` and `--registry-url` are used.

### Test Plan

<!--
**Note:** *Not required until targeted at a release.*
The goal is to ensure that we don't accept proposals with inadequate testing.

All code is expected to have adequate tests (eventually with coverage
expectations). Please adhere to the [Zarf testing guidelines][testing-guidelines]
when drafting this test plan.

[testing-guidelines]: https://docs.zarf.dev/contribute/testing/
-->

For end-to-end testing, a new type of Kubernetes cluster should be made available, where IPv6 is the only networking stack available.


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

If an administrator with an existing Kubernetes cluster, configured with dual-stack networking, wants their cluster to use the IPv6 setup then he can run `zarf init --ipv6`.

If an administrator wants to stop using the IPv6 setup then he runs `zarf init` without the `--ipv6` flag and their cluster will go back to the IPv4 setup.

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

Code complexity: the init code (Golang and Helm template) needs to support two paths, one for IPv4 (the current solution) and IPv6.

## Alternatives

<!--
What other approaches did you consider, and why did you rule them out? These do
not need to be as detailed as the proposal, but should include enough
information to express the idea and why it was not acceptable.
-->
In a later stage, the proxy component could be replaced by a component similar to the Rust Zarf injector (or even the Zarf injector itself - a proxy based on the Rust Tokyo library - already part of the used libraries - is only a few lines of code).

## Infrastructure Needed (Optional)

<!--
Use this section if you need things from the project. Examples include a new repo,
cloud infrastructure for testing or GitHub details. Listing these here
allows the process to get these resources to be started right away.
-->

- ClusterIP works from the node, but cluster DNS does not. You cannot use http://zarf-registry.zarf.svc.cluster.local:5000/v2 from the node for example. 
- What if the service needs to be re-created and it's created with a different clusterIP
  - If an image is pushed to an OCI registry and the domain changes, the image doesn't change.
  - There should
- What about the automatic port forward to 127.0.0.1:31999 on commands like `zarf tools registry ls 127.0.0.1:31999/path`. Likely we will have a DNS name that replicates it. 
- I believe we'll have to make config for several different container runtime interfaces so we can give a cert without trusting the cert at the node level
- I need to evaluate the proxy method to the local host registry and see if it can be better secured
- I'll need to check the use of different signing keys
- I will need to test the difference that a proxy makes. For example, does a proxy make openshift work? Does a proxy make things work with NFtables instead of IP tables.  


The question we need to ask is do we need the users input on whether or not this is an IPV6 cluster. Maybe and maybe not. 
Here is currently all the ways that IPV6 is used
- It's used to decide whether or not the long living daemonset is spun up - use hostNetwork
- It's used to decide whether or not to spin up the daemonset injector - use hostNetwork
- It's used to decide if the docker registry service should be a nodeport or clusterIP - use hostNetwork
- It's used to configure the address of the localhost registry, whether we should use [::1] or 127.0.0.1, - this uses IPV6 and can be determined automatically


#### Notes on proxy solution

- There are two daemonsets in this solution. The first is the zarf-injector which is how the Zarf docker registry initially gets things spun up. The second is the long living daemonset which is how the registry moves to the new solution. Currently I have it working with the first daemonset but not the second. Likely because the first daemonset has the total loopback whereas I will need to separate out of the concept of IPV6 vs the concept of host network. 




## Risks

- One problem with this solution is spinning up new nodes. The proxy daemonset on new nodes won't have the required image since the injector daemonset will not have spun up. This can be solved with a zarf init, however it would be better if it could be solved automatically. One idea could be to have a mutating webhook spin up an injector when a new node enters the cluster. IMO, it's fine to force users to run `zarf init` when creating a new node while this feature is not yet in GA.


Practical risks:
- Some distros may disallow this
  - Kind - works
  - K3D - works
  - k3s - tbd
  - RKE2 - tbd
  - openshift - tbd
  - microk8s - tbd
  - talos - works, but need
  - k0s - almost certainly works
- Some CNIs may disallow host network or host port

