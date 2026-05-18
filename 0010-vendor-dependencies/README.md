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
  - [Graduation Criteria](#graduation-criteria)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
  - [Version Skew Strategy](#version-skew-strategy)
- [Implementation History](#implementation-history)
- [Drawbacks](#drawbacks)
- [Alternatives](#alternatives)
- [Followup work](#followup-work)
<!-- /toc -->

## Summary

Vendoring external dependencies involves copying third-party library code
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

- Replace go mod with a different tool.

## Proposal

The Go community has long debated the merits of vendoring dependencies. Research
reveals compelling arguments on both sides. This section explores these arguments
and explains why vendoring was chosen for the Zarf project.

### Pros

1. **Reproducible and Consistent Builds**

   Vendoring supports [hermetic builds](https://sre.google/sre-book/release-engineering/#hermetic-builds-nqslhnid),
   ensuring consistent and predictable outcomes.

2. **Self-contained Builds**

   Vendoring allows disconnected, self-contained builds immediately after cloning
   the repository, with only the Go compiler required.

3. **Faster CI/CD Build Times**

   Local dependencies eliminate the need to fetch external dependencies, speeding
   up builds.

4. **Resilience to External Change**

   A local copy of dependencies mitigates risks from upstream issues, such as projects
   being deleted or altered unexpectedly.

5. **Auditability and Control**

   Vendored dependencies make it easier to audit code and monitor license changes,
   preventing unwanted changes like license incompatibility, excessive growth of
   dependencies.

### Cons

1. **Increased repository size**

   Vendoring adds to repository's size. However, disk space is inexpensive, and many
   large projects (eg. [Kubernetes](https://github.com/kubernetes/kubernetes/), with
   a `vendor/` directory ~70MB) have benefited from improving their dependency management.
   For Zarf, vendoring would increase the repository size by approximately 470MB.
   This will affect the overall repository size and the `.git/` folder which keeps
   the git configuration and full history. The amount of changes over time is the
   primary increase motivator of `.git/` directory.

2. **Maintenance Overhead**

   Developers must run `go mod tidy` and `go mod vendor` when modifying dependencies.
   However, existing checks ensure consistency in `go.mod` and `go.sum`, and these
   can be extended to validate the `vendor/` directory.

3. **Cluttered Repository History**

   Vendored updates may add noise to the commit history. However, Zarf's history
   already includes [multiple entries from dependabot](https://github.com/zarf-dev/zarf/commits?author=dependabot%5Bbot%5D),
   so the additional entries should not pose significant issues.

### User Stories (Optional)

#### Story 1

As a developer, I want to have all project dependencies available in the repository,
so that I can build the project without needing a working internet connection.

#### Story 2

As a developer, I want to review and audit the dependency code in the `vendor/` directory,
so that I can identify potential security or licensing issues.

### Risks and Mitigations

- Risk: `vendor/` directory can silently go out of sync if a developer forgets
  to run `go mod vendor` after updating `go.mod`.

  Mitigation: introduce CI check, see [Design details #4](#design-details).

- Risk: vendoring code into the repository means malicious or vulnerable code is
  committed directly.

  Mitigation: CodeQL scanning [Design Detail #3](#design-details) and dependency
  auditing via Dependabot [Design details #2](#design-details).

- Risk: Repository growth over time: the 470MB initial size will compound with
  every dependency update in git history.

  Mitigation: Audit existing dependencies and identify which ones we could get
  rid of, or replace.

## Design Details

The necessary changes to implement this enhancement include:

1. Include all dependent libraries by running `go mod tidy` and `go mod vendor`
   to populate `vendor/` directory.
2. Update [Dependabot configuration](https://github.com/zarf-dev/zarf/blob/main/.github/dependabot.yaml)
   to [enable support for vendored dependencies](https://docs.github.com/en/code-security/dependabot/working-with-dependabot/dependabot-options-reference#vendor--).
3. Modify the [Codeql configuration](https://github.com/zarf-dev/zarf/blob/main/.github/codeql.yaml)
   to ignore the `vendor/` directory in scans.
4. Adjust [go mod checker](https://github.com/zarf-dev/zarf/blob/main/.github/workflows/check-go-mod.yml)
   to validate the `vendor/` directory instead of solely relying on `go.mod` and
   `go.sum` files.

No changes are required to the build process, as the Go compiler [automatically uses](https://go.dev/ref/mod#vendoring)
the `vendor/` directory if `vendor/modules.txt` is present and consistent with `go.mod`.

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
- 2026-04-15: Document updated
- 2026-05-13: Document updated

## Drawbacks

The reasons why this enhancement should not be implemented are presented in the
[Cons](#cons) section, along with an explanation of how each problem is being addressed.

## Alternatives

Several alternatives were considered for this particular problem. All of them
are discussed below, along with the reason for the rejection.

### Local cache or mirror

First alternative discussed was setting up a local cache or mirror for all dependencies.
This option was rejected because it requires additional infrastructure to be built
and maintained, as well as adjustments to developer configurations. Furthermore,
it does not align with any of the [goals](#goals).

### Optional vendoring

Vendoring can be made optional, but it would have to apply to entire repository,
not on a per-developer basis. This isn't much different from the current situation,
where any developer can locally vendor dependencies and make sure they are not
accidentally committed. This option was rejected because the whole point of this
document is to address problems identified in the [goals](#goals) section.

### Vendor on release

We could vendor only on per-releases basis. This would mean that the main repository
would never contain or track `vendor/`, only the released versions would have the
`vendor/` directory associated with them. This option was rejected because although
it resolves the problem of hermetic and disconnected builds, it does not allow
dependency auditing and tracking on a daily basis.

## Followup work

During discussions about dependency management we identified two additional tasks
that should be addressed as a followup from this work.

1. Verify unwanted dependencies ([#4894](https://github.com/zarf-dev/zarf/issues/4894)).
2. Audit dependencies ([#4985](https://github.com/zarf-dev/zarf/issues/4895))
