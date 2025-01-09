# ZEP-10: Vendoring external dependencies

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

Vendoring external dependencies involves copying third-party libraries code
directly into the project under a `vendor/` directory. This document outlines
the advantages and disadvantages of this approach, aiming to support the adoption
of vendoring for the Zarf project.

## Motivation

Zarf is [advertised](https://docs.zarf.dev/) as:

> a free and open-source tool that enables declarative creation & distribution
> of software into air-gapped/constrained/standalone environments.

Its primary objective is to facilitate software delivery to disconnected environments.
Consequently, it is logical for Zarf itself to support building in such environments.

### Goals

- Enable hermetic (reproducible and consistent) builds.
- Support builds in disconnected environments.
- Provide tighter control over dependencies.

### Non-Goals

- Modify existing dependency management tool.

## Proposal

The Go community has long debated the merits of vendoring dependencies. Research
reveals compelling arguments on both sides. This section explores these arguments
and explains why vendoring was chosen for the Zarf project.

#### Pros

1. **Reproducible and Consistent Builds.**

   Vendoring supports [hermetic builds](https://sre.google/sre-book/release-engineering/#hermetic-builds-nqslhnid),
   ensuring consistent and predictable outcomes.

2. **Self-contained Builds.**

   Vendoring allows disconnected, self-contained builds immediately after cloning
   the repository, with only the Go compiler required.

3. **Faster CI/CD Builds Times.**

  Local dependencies eliminate the need to fetch external dependencies, speeding
  up builds.


4. **Resilience to External Change.**

   A local copy of dependencies mitigates risks from upstream issues, such as projects
   being deleted or altered unexpectedly.

5. **Auditability and Control.**

   Vendored dependencies make it easier to audit code and monitor license changes,
   preventing unwanted changes like license incompatibility, excessive dependencies
   growth.

### Cons

1. **Increased repository size.**

   Vendoring adds to repository's size. However, disk space is inexpensive, and many
   large projects (eg. [Kubernetes](https://github.com/kubernetes/kubernetes/), with
   a `vendor/` directory ~70MB) have benefited from improving their dependency management.
   For Zarf, vendoring would increase the repository size by approximately 410MB.

2. **Maintenance Overhead.**

   Developers must run `go mod tidy` and `go mod vendor` when modifying dependencies.
   However, existing checks ensure consistency in `go.mod` and `go.sum`, and these
   can be extended to validate the `vendor/` directory.

3. **Cluttered Repository History.**

   Vendored updates may add noise to the commit history. However, Zarf's history
   already includes [multiple entries from dependabot](https://github.com/zarf-dev/zarf/commits?author=dependabot%5Bbot%5D),
   so the additional entires should not pose significant issues.


### User Stories (Optional)

#### Story 1

As a developer, I want to have all project dependencies available in the repository,
so that I can build the project without needing an internet connection.

#### Story 2

As a developer, I want to review and audit the dependency code in the `vendor/` directory,
so that I can identify potential security or licensing issues.

### Risks and Mitigations

N/A

## Design Details

The necessary changes to implement this enhancement include:

1. Include all dependant libraries by running `go mod tidy` and `go mod vendor`
   to populate `vendor/` directory.
2. Update [Dependabot configuration](https://github.com/zarf-dev/zarf/blob/main/.github/dependabot.yaml)
   to [enable support for vendored dependencies](https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference#vendor--).
3. Modify the [Codeql configuration](https://github.com/zarf-dev/zarf/blob/main/.github/codeql.yaml)
   to ignore the `vendor/` directory in scans.
4. Adjust [go mod checker](https://github.com/zarf-dev/zarf/blob/main/.github/workflows/check-go-mod.yml)
   to validate the `vendor/` directory instead of solely relying on `go.mod` and
   `go.sum` files.

No changes are required to the build process, as the Go compiler [automatically uses](https://go.dev/ref/mod#go-mod-file-go)
the `vendor/` directory if `vendor/modules.txt` is present and consistent with `go.mod`.

## OPEN QUESTION

**What to do with [zarf/hack/schema] sub-project?**

The simplest approach seems to be to leverage the idea of [multi-module workspaces](https://go.dev/doc/tutorial/workspaces)
and implement that instead of maintaining separate `vendor/` directories for each
of the projects.

### Test Plan

[x] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

### Graduation Criteria

This feature will be implemented directly in a stable state, as it primarily impacts
the development workflow and poses minimal risk.

### Upgrade / Downgrade Strategy

N/A

### Version Skew Strategy

N/A

## Implementation History

- 2025-01-09: Document created

## Drawbacks

The reasons why this enhancement should not be implemented are presented in the
[Cons](#cons) section, along with explanation how each problem is being addressed.

## Alternatives

Alternative approaches considered for this particular problem included setting
up a cache or mirror. The option was rejected because it requires additional
infrastructure to be built and maintained, as well as adjustments to developer
configurations. Furthermore, it does not align with any of the [goals](#goals).

## Infrastructure Needed (Optional)

N/A
