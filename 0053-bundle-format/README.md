# ZEP-0053: Sigstore Bundle Format Migration and Verification Strategy

## Summary

This proposal outlines enhancements to the Zarf package signing and verification lifecycle in parity with [Cosign](https://github.com/sigstore/cosign). This includes the migration from Zarf's legacy signature format (`zarf.yaml.sig`) to the standardized Sigstore bundle format (`zarf.bundle.sig`) as the default signing mechanism. In doing so this creates more options around Zarf package provenance while ensuring that default behaviors support airgapped or otherwise security-critical environments.

## Motivation

Zarf currently supports the legacy signature format via asymmetrical keypairs, with incomplete handling of various connectivity scenarios and signing strategies:

1. **Keypair-only support**: Signing currently only supports asymmentrical keypairs
2. **Missing online verification path**: No support for transparency log verification when connectivity is available
3. **Lack of keyless signing support**: Does not support keyless signing/verification
4. **Lack of portability**: Public key distribution adds an additional artifact for every unique keypair that must be retrieved - zarf does not handle this natively

These gaps create confusion for users deploying Zarf in different environments across many Zarf packages and may lead to security misconfigurations, denial of deployment, or otherwise reduced trust in the package delivery process.

### Goals

- Migrate to [Sigstore bundle format](https://docs.sigstore.dev/about/bundle/) as the default signature format
- Supporting keyless (OIDC-based) signing with online and private signing infrastructure
- Enable optional online verification with transparency log when connectivity is available
- Enable offline verification for multiple signing strategies
- Maintain backward compatibility with offline keypair signing

### Non-Goals

- Implementing a chain of custody for Zarf package verification
- Implementing custom transparency log infrastructure within Zarf
- Modifying the Sigstore bundle specification itself
- Supporting legacy signature format storage indefinitely (will be deprecated)
- Support for `cosign verify` direct verification of a package

## Proposal

### Overview

The [Sigstore bundle format](https://docs.sigstore.dev/about/bundle/) will become the default and only supported format for inclusion of signed-material in the package, with a clear deprecation timeline for the legacy format. Zarf will retain the ability to verify packages using the legacy signature format for backwards compatibility.

> A Sigstore bundle is everything required to verify a signature on an artifact. This is satisfied by the Verification Material _and_ signature Content.

Configuration will be exposed for including a Sigstore [Trusted Root](https://docs.sigstore.dev/about/security/#sigstores-trust-root) or the use of the Public Good Sigstore instance Trusted Root embedded in Zarf to enable verification of signed packages without any additional artifacts. The embedded Trusted Root is not meant to be the sole authoritative verification material for Zarf; rather it serves as the default verification material to enable the core signing and verification workflow while allowing for any other Trusted Root to be passed in for verification.

Given a Bundle and a Trusted Root - users will have everything required to perform verification without any connectivity required. 

### User Stories

#### Story 1: Offline Sign - Offline Verify

As an operator deploying Zarf in an airgapped network, I need to verify package signatures without any external network connectivity while maintaining full cryptographic verification of package authenticity. This maintains backwards compatibility with existing keypair signing.

**Solution**: Package Sign by default will sign/verify without network connectivity.

```bash
# Developer (airgap environment) - sign package
zarf package sign zarf-package-app-amd64.tar.zst --signing-key cosign.key

# Operator (airgap environment) - verify
zarf package verify zarf-package-app-amd64.tar.zst --key cosign.pub
```

#### Story 2: Online Sign - Offline Verify

As a package author using Zarf in a connected environment, I want to leverage the Sigstore public good infrastructure to provide additional assurance that package signatures are publicly recorded and auditable.

**Solution**: Use sigstore public-good infrastructure with embedded Trusted Root verification.

```bash
# Developer - sign with keyless
zarf package sign zarf-package-app-amd64.tar.zst --oidc-issuer='https://oauth2.sigstore.dev/auth' --sigstore-public-defaults

# Operator - verify with embedded trusted root - no external connectivity made
zarf package verify zarf-package-app-amd64.tar.zst

# Additionally verify with identity verification
zarf package verify zarf-package-app-amd64.tar.zst \
  --certificate-identity user@example.com \
  --certificate-oidc-issuer https://oauth2.sigstore.dev/auth
```

#### Story 3: Online Sign - Online Verify - Private Sigstore Infrastructure

As an enterprise running a private Sigstore instance for internal packages, I want to use a custom trusted root and private Fulcio/Rekor instances for signing and verification.

**Solution**: Use online profile for signing and online verify with internal URLs or custom trusted root

```bash

# Sign with private infrastructure
zarf package sign zarf-package-app-amd64.tar.zst \
  --oidc-issuer https://oauth2.internal.company.com/auth \
  --rekor-url https://rekor.internal.company.com \
  --fulcio-url https://fulcio.internal.company.com

# Verify with private infrastructure in an "online" profile
zarf package verify zarf-package-app-amd64.tar.zst \
  --rekor-url https://rekor.internal.company.com \
  --fulcio-url https://fulcio.internal.company.com

# or verify in the default offline stance with a custom trusted-root
# Create custom trusted root
zarf tools trusted-root create \
    --fulcio="url=https://fulcio.sigstore.dev,certificate-chain=/path/to/fulcio.pem,end-time=2025-01-01T00:00:00Z" \
    --rekor="url=https://rekor.sigstore.dev,public-key=/path/to/rekor.pub,start-time=2024-01-01T00:00:00Z" \
  --output custom_trusted_root.json

# Verify with custom trusted root
zarf package verify zarf-package-app-amd64.tar.zst \
  --trusted-root custom_trusted_root.json
```

#### Story 4: Trusted Root retrieval and Utilization

As an operator deploying zarf packages, I want the ability to easily retrieve the latest Trusted Root from defined infrastructure (public or private) in order to circumvent any potential gaps in availability of updates for newer trusted roots.

```bash

# Retrieve the latest trusted root (mirrors cosign trusted-root create)
zarf tools trusted-root create --with-default-services --out trusted-root.json

# Verify with the retrieved trusted root
zarf package verify zarf-package-app-amd64.tar.zst \
  --trusted-root trusted-root.json
```

### Risks and Mitigations

#### Risk 1: Embedded Trusted Root Expiration

**Risk**: If embedded trusted root is not updated, verification may fail due to expired keys.

**Mitigation**:
- Explicit support for retrieving and including updated Trusted Roots from the command line
- Pre-release checklist includes trusted root update verification
- Monitor Sigstore for key rotations
- Renovate monitoring for updates to the Trusted Root repository
- Documentation on manual trusted root updates

#### Risk 2: Compromised Trusted Root

**Risk**: If the trusted root is compromised, verification may fail or be otherwise inaccurate.

**Mitigation**:
- Evaluate upstream (or other administrators of the infrastructure) for impact and response
- Evaluate impact to packages signed for any required re-signing
- Retrieve a new trusted root with `zarf tools trusted-root create` once available
- Use said trusted root until a Zarf binary is released using the updated trusted root (if using the public good instance)

#### Risk 3: Bundle Format Incomplete in Offline Mode

**Risk**: Sigstore bundle may require fields that cannot be populated offline.

**Mitigation**:
- Validate minimal bundle structure works with cosign verification given offline keypair signing and verification
- Document minimal vs. complete bundle differences

## Design Details

### Connectivity Requirements

Zarf should be aware of connectivity requirements - whether the connectivity is present or not. As such this section outlines explicit awareness of the requirements whereby the default stance attempts to ensure no additional network connectivity while still presenting optionality for connected signing and verification workflows. 

#### Default Stance

**Purpose**: Complete offline operation without any network dependencies.

**Signing Behavior**:
- Supports keypair-based signing only
- Supports KMS if accessible
- Does NOT upload to transparency log (`tlogUpload = false`) - therefore does NOT contact Rekor
- Does NOT contact Fulcio (no keyless signing support)

**Verification Behavior**:
- Uses embedded Sigstore trusted root (fetched via The Update Framework (TUF) and embedded into the Zarf binary/SDK)
- Skips transparency log verification (`IgnoreTlog = true`)
- Skips signed certificate timestamp verification (`IgnoreSCT = true`)

#### Online Options

**Purpose**: Full connectivity to Sigstore infrastructure (public/private) with transparency log support.

**Signing Behavior**:
- Supports all signing methods (keypair, keyless, cloud KMS)
- Uploads to transparency log (`tlogUpload = true`)
- Contacts Fulcio for keyless signing certificates
- Creates complete Sigstore bundle with transparency log entries

**Verification Behavior**:
- Uses embedded or custom trusted root
- Operates in online mode (`Profile = online`)
- Verifies transparency log entries (`IgnoreTlog = false`)
- Verifies signed certificate timestamp if present (`IgnoreSCT = false`)

**Use Cases**:
- Connected cloud environments
- CI/CD pipelines with keyless signing
- Public package repositories
- Maximum security with public transparency
- Public/Private Sigstore infrastructure

### CLI Changes

#### Flags for Cosign Parity

The use of [Cosign](https://docs.sigstore.dev/cosign/signing/overview/) will require additional flags for Signing and Verification.

`zarf package sign|verify` will include all required flags in order to support parity of operations with `cosign sign-blob|verify-blob`. 

Defaults for these flags will ensure configuration is not implicitly pointing to resolvable infrastructure such that someone may unintentionally make a network call. Helper flags will be included such as `--sigstore-public-defaults` to easily configure for opt-in Sigstore public-good infrastructure utilization. 

These flags will not be added to other `zarf package` commands given the surface area of available configuration and required synchronization with `Cosign`.

### Trusted Root Management

#### Embedded Trusted Root Update Process

The embedded trusted root should be updated regularly to include new Sigstore keys and prevent expiration issues. Zarf will utilize The Update Framework (TUF) to establish provenance in the retrieval and update process for the embedded data.

**Update Schedule**:
- Checked before releases
- Automatically with tooling

#### Embedded Trusted Root SDK Experience

The Zarf Cosign utilities in the SDK offer generic implementation around `sign-blob` and `verify-blob`. The embedded trusted root is available for consumption by default while also allowing SDK consumers to override with their own embedded trusted root by passing it as a referenced path. 

### Migration Path from Legacy Format

#### Phase 1: Dual Format Support

**Behavior**:
- Both `zarf.bundle.sig` and `zarf.yaml.sig` created during signing
- Verification checks bundle first, falls back to legacy
- Deprecation warnings logged when legacy format is used
- All new packages signed with bundle format

**Actions**:
- Document bundle format in user tutorials
- Add migration guide to documentation
- Provide tooling to re-sign pre-existing packages signed with legacy format signature

#### Phase 2: Bundle Format Default, Legacy Deprecated

**Behavior**:
- Only `zarf.bundle.sig` created during signing
- Legacy format verification still supported for backward compatibility
- Warnings when verifying legacy signatures

**Actions**:
- Remove legacy signature creation from signing operations
- Update all documentation to reference bundle format only

### Test Plan

#### Unit Tests

Unit tests will validate profile selection (defaults, overrides, validation), signing operations across all profiles and methods (keypair), verification behaviors (offline with embedded/custom roots, online with transparency logs, signature validation (invalid signatures, tampered packages), and trusted root management (loading, overrides, error handling, priority).

#### E2E Tests

E2E tests will cover workflows for each profile: offline (keypair generation, signing, offline verification, deployment), online (transparency log upload and verification, bundle structure validation), custom trusted roots (private Sigstore infrastructure), and keyless signing (OIDC-based signing with certificate validation).

### Graduation Criteria

#### Alpha

**Criteria**:
- Bundle signing implemented
    - Bundle format always created
    - Deprecate legacy signature
- Embedded Trusted Root implemented
    - Retrieval
    - Embedding & Development Process
    - Helper Functions
- Backwards Compatibility
    - Fallback to legacy signature when bundle is not present
- Documentation available
    - Development & Update Processes
    - Command Line Documentation
    - Tutorials

**Exit Criteria**:
- All unit tests passing
- Basic E2E test for offline profile (default) passing
- No critical bugs reported in profile selection

#### Beta

**Criteria**:
- Automated trusted root updates
    - Updates to release process
    - Detection & alerting of rotation to trusted root
- Connected Signing & Verification
    - Keyless sign/verify availability
- Comprehensive documentation published
    - Architecture Documentation

**Exit Criteria**:
- All test suites passing (unit, E2E, integration)
- At least 2 releases with bundle signing
- User feedback incorporated
- No critical bugs in bundle verification

#### GA

**Criteria**:
- Production usage across multiple organizations
- Documentation complete and accurate
- Migration period complete (6+ months)

**Exit Criteria**:
- Zero critical bugs in signing/verification for 2+ releases
- Community feedback positive
- All migration tooling validated in production
- Performance characteristics documented

### Upgrade/Downgrade Strategy

#### Upgrading to Bundle Format

**From Previous Versions** (legacy signatures):

1. **Upgrade Zarf**: Install version supporting bundle format
2. **Sign and Verify new Packages**: Bundle format included in the package
3. **Verify Legacy Packages**: Verification still works
4. **Migrate Existing Packages** (optional but recommended):
   ```bash
   zarf package sign old-package.tar.zst \
     --signing-key cosign.key
   ```

### Version Skew Strategy

#### Signing Version vs. Verification Version

**Scenario**: Package signed with Zarf bundle format, verified with Zarf legacy format

**Result**: Works (The legacy signature exists alongside the bundle - deprecated until removal)

**Scenario**: Package signed with Zarf legacy format, verified with Zarf bundle format

**Result**: Works (Will fallback to the legacy signature when bundle is not present)

#### Trusted Root Version Skew

**Scenario**: Package signed with trusted root from Jan 2025, verified with Zarf binary from Dec 2024

**Result**: May fail if Jan 2025 root contains rotated keys not in Dec 2024 root

**Mitigation**:
- Update Zarf regularly to get latest embedded trusted root
- Use custom trusted root if specific version required - introduce helper command if needed
- Monitor Sigstore announcements for key rotations

## Implementation History

- 2025-12-01: ZEP-0053 Ready for Review

## Drawbacks

### Dependency on Sigstore Infrastructure

Online profile creates a dependency on Sigstore infrastructure (Rekor, Fulcio) which could become a single point of failure.

**Counter-argument**: Offline profile has zero dependencies. Online profile is opt-in for users who specifically want transparency. Private Sigstore deployments are supported.

## Alternatives

### Alternative 1: Use Sigstore-go directly for signing/verification

**Description**: Integrate the [sigstore-go](https://github.com/sigstore/sigstore-go) library directly rather than wrapping Cosign CLI. Sigstore-go is a minimal dependency library designed specifically as an API for Sigstore signing and verification, offering programmatic access to bundle creation/verification without Cosign's broader feature set and dependencies.

**Pros**:
- Minimal API
- Smaller dependency surface

**Cons**:
- Overhead of matching parity with Cosign

**Rejection Reason**: Cosign provides intuitive and expected functionality and comprehensive signing/verification support out-of-the-box. Users of Zarf will be more familiar with Cosign over the underlying Sigstore-go library. Use of Cosign will be familiar to those who have been working in the cloud native supply chain ecosystem for artifacts that Zarf packages.

### Alternative 2: Connectivity Profiles for configuration

**Description**: Specify connectivity "profiles" that capture "online" and "offline" defaults for cosign option configurations such that all cosign flags are not exposed to end-users. 

**Pros**:
- Lower barriers to entry
- Reducing options exposed to CLI end users

**Cons**:
- Potential confusion for permutations of online/offline and asymmetrical keypairs (IE KMS)
- Less explicit configuration for CLI end users

**Rejection Reason**: Profiles create some confusion over the permutations of execution scenarios versus parity with Cosign configuration. Additionally Profiles can be implemented outside of the proposal if deemed necessary. 

### Alternative 3: Always Use Online Mode

**Description**: Default to online verification with transparency log, make offline mode opt-in.

**Pros**:
- Maximum security with transparency
- Aligns with Sigstore's vision
- Public audit trail for all signatures

**Cons**:
- Breaks Zarf's airgap-first philosophy
- Creates network dependency by default

**Rejection Reason**: Violates Zarf's core mission of supporting airgapped deployments. Default must work offline.

## Infrastructure Needed

### Testing Infrastructure

**CI/CD Infrastructure**:
- Automated trusted root updates (Renovate or other actions)

### Production Infrastructure

**No infrastructure required** for Zarf itself. Users choosing online profile will need:

**Sigstore**:
- Available at https://rekor.sigstore.dev and https://fulcio.sigstore.dev or private Sigstore instances.