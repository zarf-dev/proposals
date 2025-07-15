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

# ZEP-NNNN: Your short, descriptive title

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

This ZEP proposes adding image signature verification capabilities to Zarf for `images` that are signed with Cosign signatures. The proposal introduces a configurable mechanism that allows package creators to define registry-specific verification rules using glob pattern matching. This enhancement would enable Zarf to verify the authenticity and integrity of container images during package creation by checking their cryptographic signatures against specified certificates or public keys.

By implementing this feature, Zarf would provide stronger security guarantees for air-gapped environments where supply chain security is critical and would provide a mechanism to catch issues early in the package development process. The proposal includes support for both public key infrastructure and keyless verification methods, accommodating various security postures and compliance requirements across different operational environments.

## Motivation

Container image signature verification is a critical component of supply chain security. In air-gapped environments where Zarf deploys packages, ensuring the integrity and authenticity of container images is particularly important since traditional online verification methods are unavailable. Currently, Zarf lacks native support for verifying signatures on container images, which creates a security gap that must be filled with other tools.  Configuring those tools in the airgap (such as Kyverno) can also be complex because the tools must be configured to work offline - configuring tools like cosign manually on package create can also be cumbersome due to the number of images some zarf packages contain.  This feature would also help align with industry best practices for container security as outlined by organizations like the Cloud Native Computing Foundation (CNCF) and the Open Source Security Foundation (OpenSSF).

### Goals

- Implement container image signature verification for `images` being pulled into Zarf packages
- Provide a flexible configuration mechanism that supports each of the various signing methods supported by Cosign

### Non-Goals

- Support other signature verification systems beyond Cosign (such as Notary or Docker Content Trust)
- Automatic signature verification for common registries/configurations

## Proposal

This proposal introduces a new configuration option for Zarf package creation that allows package creators to specify Cosign verification options for container images based on registry glob pattern matching to apply specific verification rules for different registries.

When a container image is pulled from a registry that matches a configured pattern, Zarf would verify its signature using the specified verification options before including it in the final package.  These options would be mapped to Cosign's CLI options to provide a familiar interface to users and allow for the same verification options to be used in both Zarf as in Cosign.

### User Stories (Optional)

#### Story 1: Platform Engineer Using Iron Bank Images

As a platform engineer I want to validate container image signatures on package creation so that I can ensure the integrity of the images before they leave for my air-gapped environment.

**Given** I have a Zarf Package images I would like to verify
```yaml
# zarf.yaml
components:
  - name: my-component
    required: true
    images:
      - registry1.dso.mil/ironbank/opensource/zarf-dev/zarf/zarf-agent:v0.58.0
      - ghcr.io/stefanprodan/podinfo:6.5.0
```
**When** I create that package with a `zarf-config.yaml` like the below:
```yaml
# zarf-config.yaml
package:
  create:
    cosignOpts:
      "registry1.dso.mil/ironbank/*":
        certificate-chain: cosign-ca-bundle.pem
        cert: cosign-certificate.pem
      "ghcr.io/stefanprodan/*":
        certificate-identity-regexp: "^https://github.com/stefanprodan/podinfo.*$"
        certificate-oidc-issuer: https://token.actions.githubusercontent.com
```
**Then** Zarf will use cosign to validate the most specific glob pattern match during image pulls
**And** Zarf will fail the package creation if the image signature verification fails

### Risks and Mitigations

**Performance Impact**: Signature verification adds an additional step to the image pull process, which could impact performance during package creation and deployment. To mitigate this, verification is opt-in and will be performed concurrently where possible.

**UX Complexity**: Because of the number of options available in Cosign, the configuration options could become complex and may require additional documentation to help users understand the available options.  Error messages on failures will also need to be clear and actionable calling out the specific image that failed verification along with the specific reason for the failure from cosign.

## Design Details

### Configuration Structure

The image signature verification feature will be implemented by adding a new `cosignOpts` field to the `package.create` section of the Zarf package configuration. This field will contain a map of glob patterns to verification options.

```yaml
package:
  create:
    cosignOpts:
      "registry1.dso.mil/ironbank/*":
        certificate-chain: cosign-ca-bundle.pem
        cert: cosign-certificate.pem
      "ghcr.io/stefanprodan/*":
        certificate-identity-regexp: "^https://github.com/stefanprodan/podinfo.*$"
        certificate-oidc-issuer: https://token.actions.githubusercontent.com
```

The glob patterns will be matched against the full image reference (including the registry, repository, and tag) to determine which verification options to apply. If multiple patterns match an image, the most specific pattern (determined by the number of characters matched and wildcards contained) will be used.

For example for the image `registry1.dso.mil/ironbank/zarf/zarf-agent:v0.58.0`, with the following patterns:

- `registry1.dso.mil/ironbank/*`
- `registry1.dso.mil/ironbank/zarf/*`
- `registry1.dso.mil/ironbank/zarf/zarf-agent`

The most specific pattern will be `registry1.dso.mil/ironbank/zarf/zarf-agent` because it contains the fewest wildcards and matches the most characters of the image reference.

```golang
func findMostSpecificPattern(imageRef string, patterns []string) string {
    var bestPattern string
    var highestScore int

    for _, pattern := range patterns {
        if matchesPattern(imageRef, pattern) {
            // Calculate specificity score
            wildcardCount := strings.Count(pattern, "*")
            exactCharCount := len(pattern) - wildcardCount
            
            score := exactCharCount*10 - wildcardCount*5 // This can be further tuned during implementation
            
            if score > highestScore {
                highestScore = score
                bestPattern = pattern
            }
        }
    }
    
    return bestPattern
}
```

### Verification Process

During package creation, when Zarf pulls a container image, it will check if the image matches any of the configured patterns. If a match is found, Zarf will verify the image's signature using the specified verification options before including it in the package.

The verification process will use the Cosign library to perform the actual verification mapping the configured verification options to the Cosign CLI options.

If verification fails, Zarf will provide a clear error message indicating which image failed verification and why.

### CLI Integration

The `cosignOpts` configuration will also be accessible through the Zarf CLI using the `--cosignOpts` flag, which will accept a JSON string containing the verification options:

```bash
zarf package create . --cosignOpts='{"registry.dso.mil/*": {"certificate-chain": "cosign-ca-bundle.pem", "cert": "cosign-certificate.pem"}}'
```

This allows users to specify verification options without modifying the Zarf configuration file in scripts or pipelines.

### Test Plan

The image signature verification feature will be tested using a combination of unit tests and end-to-end tests to ensure it functions correctly and reliably.

#### Unit Tests

- Test the pattern matching logic to ensure that glob patterns correctly match image references
- Test the verification option parsing to ensure that configuration is correctly interpreted

#### End-to-End Tests

- Test the full package creation workflow with signed images from public registries
- Test verification with different types of verification options (certificate-based, oidc-based, etc.)
- Test verification with images from different registries to ensure that pattern matching works correctly

[X] I/we understand the owners of the involved components may require updates to
existing tests to make this code solid enough prior to committing the changes necessary
to implement this proposal.

### Graduation Criteria

Pending review / community input these changes would move out of alpha and become a part of the stable Zarf configuration schema.

### Upgrade / Downgrade Strategy

This feature adds a new optional configuration field and does not modify any existing behavior, so no special upgrade or downgrade strategy is required. Existing packages will continue to work without modification.

To make use of the new feature, package creators will need to add the `cosignOpts` field to their Zarf package create configuration. This can be done without affecting the rest of the configuration.

If a user downgrades to a version of Zarf that doesn't support image signature verification, packages that include the `cosignOpts` field will still work, but the verification options will be ignored.

### Version Skew Strategy

This feature is primarily implemented in the Zarf CLI and does not involve coordination with other components like the Zarf Agent. The verification process happens during package creation, which is handled by the CLI.

The feature does not modify the package format or any interfaces between components, so version skew is not a concern. Packages created with signature verification enabled will be compatible with all versions of Zarf that support the current package format.

## Implementation History

- 2025-07-14: ZEP created and initial draft submitted

## Drawbacks

- Adds some additional complexity to the Zarf codebase and increases the maintenance burden to keep up with cosign options
- May impact performance during package creation to verify images
- Requires users to understand and correctly configure signature verification

## Alternatives

### Alternative 1: External Verification Tool

One alternative approach would be to rely on an external tool for signature verification, such as the Cosign CLI itself. Users would run Cosign to verify images before creating Zarf packages.

**Rejected because**: This approach would require users to manually verify each image and would not integrate well with the Zarf workflow.  It also requires users to install and manage the Cosign CLI in their environment.

### Alternative 2: Verification During Deployment Only

Another approach would be to only verify signatures during deployment, rather than during package creation with a tool like Kyverno or by modifying the Zarf Agent.

**Rejected because**: Verifying during package creation ensures that only verified images are included in the package, providing feedback to the user as early as possible that something is wrong.  Package creation also usually occurs in a connected environment making online verification more practical.

### Alternative 3: Registry-Based Verification

A third approach would be to rely on registry-based verification, where the registry itself enforces signature verification policies.

**Rejected because**: Not all registries support signature verification, and usually requires an enterprise license even for those that do.  It would also force a workflow where Zarf would need to be locked to a given set of registries which creates additional user friction to ensure they have the right images in the right registries.

## Infrastructure Needed (Optional)

No additional infrastructure is needed for this proposal.
