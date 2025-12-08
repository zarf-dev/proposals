# ZEP-0053: Sigstore Bundle Format Migration and Verification Strategy

## Summary

This proposal outlines the migration from Zarf's legacy signature format (`zarf.yaml.sig`) to the standardized Sigstore bundle format (`zarf.bundle.sig`) as the default signing mechanism. It establishes connectivity profiles for signing and verification operations, ensuring Zarf maintains its airgap-first philosophy while supporting online verification scenarios when connectivity is available.

## Motivation

Zarf currently supports the legacy signature format via asymmetrical keypairs, with incomplete handling of various connectivity scenarios and signing strategies:

1. **Keypair-only support**: Signing currently only supports asymmentrical keypairs
2. **Missing online verification path**: No support for transparency log verification when connectivity is available
3. **Lack of keyless signing support**: Does not support keyless signing/verification
4. **Lack of portability**: Public key distribution adds an additional artifact for every unique keypair that must be retrieved - zarf does not handle this natively

These gaps create confusion for users deploying Zarf in different environments and may lead to security misconfigurations.

### Goals

- Migrate to Sigstore bundle format as the default signature format
- Define explicit connectivity profiles (offline, online) with sensible defaults
- Enable optional online verification with transparency log when connectivity is available
- Maintain backward compatibility with offline keypair signing
- Supporting keyless (OIDC-based) signing with online and private signing infrastructure

### Non-Goals

- Implementing a chain of custody for Zarf package verification
- Implementing custom transparency log infrastructure within Zarf
- Modifying the Sigstore bundle specification itself
- Supporting legacy signature format storage indefinitely (will be deprecated)

## Proposal

### Overview

This proposal introduces connectivity profiles for Zarf package signing and verification:

1. **Offline Profile** (default): No network connectivity required or used
2. **Online Profile**: Full connectivity to Sigstore infrastructure (transparency logs, certificate authorities)

The Sigstore bundle format will become the default and only supported format for inclusion of signed-material in the package, with a clear deprecation timeline for the legacy format. Zarf will retain the ability to verify packages using the legacy signature format for backwards compatibility.

Configuration will be exposed for including a Sigstore [Trusted Root](https://docs.sigstore.dev/about/security/#sigstores-trust-root) or the use of the Public Good Sigstore instance Trusted Root embedded in Zarf to enable verification of signed packages without any additional artifacts.

### User Stories

#### Story 1: Offline Sign - Offline Verify

As an operator deploying Zarf in an airgapped network, I need to verify package signatures without any external network connectivity while maintaining full cryptographic verification of package authenticity. This maintains backwards compatibility with existing keypair signing.

**Solution**: Use airgap profile (default) with keypair signing.

```bash
# Developer (online environment) - sign package
zarf package sign zarf-package-app-amd64.tar.zst --signing-key cosign.key

# Operator (airgap environment) - verify
zarf package verify zarf-package-app-amd64.tar.zst --key cosign.pub
```

#### Story 2: Online Sign - Offline Verify

As a package author using Zarf in a connected environment, I want to leverage the Sigstore public good infrastructure to provide additional assurance that package signatures are publicly recorded and auditable.

**Solution**: Use online profile with embedded Trusted Root verification.

TODO: validate required inputs
```bash
# Developer - sign with transparency log upload
zarf package sign zarf-package-app-amd64.tar.zst --signing-key cosign.key --profile online

# Operator - verify with embedded trusted root - Do we need the public key?
zarf package verify zarf-package-app-amd64.tar.zst
```

```bash
# Developer - sign with keyless
zarf package sign zarf-package-app-amd64.tar.zst --oidc-issuer='https://oauth2.sigstore.dev/auth' --profile online

# Operator - verify with embedded trusted root - no external connectivity made
zarf package verify zarf-package-app-amd64.tar.zst
```

#### Story 3: Online Sign - Online Verify - Private Sigstore Infrastructure

As an enterprise running a private Sigstore instance for internal packages, I want to use a custom trusted root and private Fulcio/Rekor instances for signing and verification.

**Solution**: Use online profile for signing and online verify with internal URLs or custom trusted root

```bash

# Sign with private infrastructure
zarf package sign zarf-package-app-amd64.tar.zst \
  --oidc-issuer https://oauth2.internal.company.com/auth \
  --rekor-url https://rekor.internal.company.com \
  --profile online

# Verify with private infrastructure in an "online" profile
Zarf package verify zarf-package-app-amd64.tar.zst \
  --rekor-url https://rekor.internal.company.com \
  --fulcio-url https://fulcio.internal.company.com \
  --profile online

# or verify in the default offline stance with a custom trusted-root
# Create custom trusted root
cosign trusted-root create \
  --rekor-url https://rekor.internal.company.com \
  --fulcio-url https://fulcio.internal.company.com \
  --output custom_trusted_root.json

# Verify with custom trusted root
zarf package verify zarf-package-app-amd64.tar.zst \
  --trusted-root custom_trusted_root.json
```

### Risks and Mitigations

#### Risk 1: Embedded Trusted Root Expiration

**Risk**: If embedded trusted root is not updated, verification may fail due to expired keys.

**Mitigation**:
- Explicit support for including updated Trusted Roots from the command line
- Pre-release checklist includes trusted root update verification
- Monitor Sigstore for key rotations
- Renovate monitoring for updates to the Trusted Root repository
- Documentation on manual trusted root updates

#### Risk 2: Bundle Format Incomplete in Offline Mode

**Risk**: Sigstore bundle may require fields that cannot be populated offline.

**Mitigation**:
- Validate minimal bundle structure works with cosign verification given offline keypair signing and verification
- Document minimal vs. complete bundle differences

#### Risk 3: Profile Confusion and Misconfiguration

**Risk**: Users may choose wrong profile or misconfigure connectivity settings.

**Mitigation**:
- Default to offline profile (safest, most aligned with Zarf mission)
- Validation of profile compatibility with flags
- Examples in documentation for each profile

## Design Details

### Connectivity Profiles

Connectivity Profiles will allow Zarf to abstract orchestrating what configurations are made to enable/disable network connectivity.

#### Offline Profile (Default)

**Purpose**: Complete offline operation without any network dependencies.

**Signing Behavior**:
- Supports keypair-based signing only
- Supports cloud KMS if accessible without internet (e.g., AWS KMS via VPC endpoint)
- Does NOT upload to transparency log (`tlogUpload = false`) - therefor does NOT contact Fulcio
- Does NOT contact Fulcio (no keyless signing support)

**Verification Behavior**:
- Uses embedded Sigstore trusted root (fetched at Zarf build time via The Update Framework (TUF))
- Operates in offline mode (`Profile = offline`)
- Skips transparency log verification (`IgnoreTlog = true`)
- Skips signed certificate timestamp verification (`IgnoreSCT = true`)

#### Online Profile

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

### CLI Changes

#### New `--profile` Flag

Add profile flag to signing and verification commands:

```bash
# Signing commands
zarf package sign <package> --signing-key cosign.key [--profile offline|online]

# Verification commands
zarf package verify <package> [--profile offline|online]
```

Default profile: `offline`

#### Cosign Optional Flags

The use of [Cosign](https://docs.sigstore.dev/cosign/signing/overview/) will require additional flags for Signing and Verification.

A subset of flags will be included on `zarf package sign` and `zarf package verify` to ensure that each signing and verification workflow can execute accordingly.

##### `zarf package sign` Flags

The following flags from `cosign sign-blob` will be available:

**Key & Certificate Management:**
- `--key` - Path to private key file, or KMS URI

**Fulcio Integration (Online Profile):**
- `--fulcio-url` - Address of Sigstore PKI server (default: `https://fulcio.sigstore.dev`)
- `--fulcio-auth-flow` - Fulcio interactive OAuth2 flow for certificate
- `--identity-token` - Identity token for certificate from Fulcio
- `--oidc-client-id` - OIDC client ID (default: `sigstore`)
- `--oidc-issuer` - OIDC provider to issue ID token
- `--oidc-provider` - Specify OIDC provider (spiffe, google, github-actions, etc.)

**Rekor Integration (Online Profile):**
- `--rekor-url` - Address of Rekor transparency log server (default: `https://rekor.sigstore.dev`)

**Timestamp & Security:**
- `--timestamp-server-url` - URL to RFC3161 timestamp server
- `--trusted-root` - Optional path to TrustedRoot JSON file

**Hardware Security:**
- `--sk` - Use a hardware security key
- `--slot` - Security key slot to use (default: `signature`)

##### `zarf package verify` Flags

The following flags from `cosign verify-blob` will be available:

**Key & Certificate Verification:**
- `--key` - Path to public key file, KMS URI, or Kubernetes Secret
- `--certificate` - Path to public certificate
- `--certificate-chain` - CA certificates in PEM format for chain building
- `--ca-roots` - Bundle of CA certificates for certificate chain
- `--ca-intermediates` - Intermediate CA certificates in PEM format (used with `--ca-roots`)

**Certificate Identity Verification:**
- `--certificate-identity` - Expected identity in Fulcio certificate
- `--certificate-identity-regexp` - Regular expression alternative for identity matching
- `--certificate-oidc-issuer` - Expected OIDC issuer in certificate
- `--certificate-oidc-issuer-regexp` - Regular expression alternative for issuer matching

**Transparency Log & Timestamp Verification:**
- `--rekor-url` - Address of Rekor transparency log server (default: `https://rekor.sigstore.dev`)
- `--timestamp-certificate-chain` - PEM-encoded certificate chain for RFC3161 timestamp authority

**Trust:**
- `--trusted-root` - Path to Sigstore TrustedRoot JSON file
- `--private-infrastructure` - Skip transparency log verification for private deployments

**Hardware Security:**
- `--sk` - Use hardware security key
- `--slot` - Security key slot (authentication, signature, card-authentication, key-management)

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
- Provide tooling to re-sign legacy packages

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
- Profile framework implemented and tested
    - `--profile` flag available on signing and verification workflows
    - Offline profile works as default
- Bundle signing implemented
    - Bundle format always created
    - Deprecate legacy signature
- Backwards Compatibility
    - Fallback to legacy signature when bundle is not present
- Online profile implemented
    - Online-sign and online-verify available
    - Keyless-signing availability
- Documentation available
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
- Comprehensive documentation published
    - Architecture Documentation

**Exit Criteria**:
- All test suites passing (unit, E2E, integration)
- At least 2 releases with bundle signing
- User feedback incorporated
- No critical bugs in bundle verification

#### GA

**Criteria**:
- All profiles stable and well-tested
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

**Scenario**: Package signed with Zarf bundle format , verified with Zarf legacy format

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

### Increased Complexity

The introduction of profiles adds a new concept for users to understand. Users must now choose between offline and online profiles rather than having a single signing/verification workflow.

**Counter-argument**: The complexity is unavoidable given Zarf's diverse deployment scenarios. Explicit profiles are clearer than implicit behavior based on flag combinations.

### Dependency on Sigstore Infrastructure

Online profile creates a dependency on Sigstore infrastructure (Rekor, Fulcio) which could become a single point of failure.

**Counter-argument**: Offline profile has zero dependencies. Online profile is opt-in for users who specifically want transparency. Private Sigstore deployments are supported.

## Alternatives

### Alternative 1: No Profiles, Flag-Based Configuration

**Description**: Use individual flags (`--ignore-tlog`, `--tlog-upload`) instead of profiles.

**Pros**:
- Maximum flexibility for advanced users
- No new concepts to learn
- Granular control over each setting

**Cons**:
- Easy to misconfigure (many flags to set correctly)
- No validation of flag combinations
- Unclear what flags are needed for specific scenarios
- Difficult to document all permutations
- High cognitive load for users

**Rejection Reason**: Profiles provide better user experience for common scenarios while still allowing flag overrides for advanced use cases.

### Alternative 2: Always Use Online Mode

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