# ZEP-0053: Sigstore Bundle Format Migration and AirGap Verification Strategy

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
- Define explicit connectivity profiles (airgap, online, hybrid) with sensible defaults
- Enable optional online verification with transparency log when connectivity is available
- Maintain backward compatibility with offline keypair signing
- Supporting keyless (OIDC-based) signing with online and private signing infrastructure

### Non-Goals

- Implementing custom transparency log infrastructure within Zarf
- Modifying the Sigstore bundle specification itself
- Supporting legacy signature format indefinitely (will be deprecated)

## Proposal

### Overview

This proposal introduces three explicit connectivity profiles for Zarf package signing and verification:

1. **Airgap Profile** (default): No network connectivity required or used
2. **Online Profile**: Full connectivity to Sigstore infrastructure (transparency logs, certificate authorities)
3. **Hybrid Profile**: Package creation using public good sigstore infrastructure and native offline verification embedded into zarf.

The Sigstore bundle format will become the default and only supported format, with a clear deprecation timeline for the legacy format.

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

As an enterprise running a private Sigstore instance for internal packages, I need to use a custom trusted root and private Fulcio/Rekor instances for signing and verification.

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

### Connectivity Profiles

#### Airgap Profile (Default)

**Purpose**: Complete offline operation without any network dependencies.

**Signing Behavior**:
- Supports keypair-based signing only (local key files or local HSM)
- Supports cloud KMS if accessible without internet (e.g., AWS KMS via VPC endpoint)
- Does NOT upload to transparency log (`tlogUpload = false`)
- Does NOT contact Fulcio (no keyless signing support)
- Creates minimal Sigstore bundle with signature only

**Verification Behavior**:
- Uses embedded Sigstore trusted root (fetched at Zarf build time via TUF)
- Operates in offline mode (`Profile = offline`)
- Skips transparency log verification (`IgnoreTlog = true`)
- Skips SCT verification (`IgnoreSCT = true`)
- No network calls during verification

**Default Configuration**:
```go
// Airgap Profile Defaults
SignBlobOptions{
    KeyOpts: options.KeyOpts{
        NewBundleFormat: true,
        SkipConfirmation: false,
    },
    Timeout: 3 * time.Minute,
}
// tlogUpload = false (hardcoded)
// No Fulcio/Rekor URLs configured
// Note: `Offline` is being deprecated for bundle behaviors - see the [related issue](https://github.com/sigstore/cosign/pull/4457)

VerifyBlobOptions{
    KeyOpts: options.KeyOpts{
        NewBundleFormat: true,
    },
    CertVerifyOptions: options.CertVerifyOptions{
        IgnoreSCT: true,
    },
    IgnoreTlog: true,
    Timeout: 3 * time.Minute,
}
```

**Use Cases**:
- Airgapped networks
- Disconnected systems
- Environments with strict network egress controls
- Maximum portability scenarios

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
- Verifies SCT if present (`IgnoreSCT = false`)
- May make network calls to verify log inclusion

**Default Configuration**:
```go
// Online Profile Defaults
SignBlobOptions{
    KeyOpts: options.KeyOpts{
        NewBundleFormat: true,
        OIDCIssuer: "https://oauth2.sigstore.dev/auth",
        OIDCClientID: "sigstore",
        FulcioURL: "https://fulcio.sigstore.dev",
        RekorURL: "https://rekor.sigstore.dev",
    },
    Timeout: 3 * time.Minute,
}
// tlogUpload = true

VerifyBlobOptions{
    KeyOpts: options.KeyOpts{
        NewBundleFormat: true,
        RekorURL: "https://rekor.sigstore.dev",
    },
    CertVerifyOptions: options.CertVerifyOptions{
        IgnoreSCT: false,
    },
    Offline: false,
    IgnoreTlog: false,
    Timeout: 3 * time.Minute,
}
```

**Use Cases**:
- Connected cloud environments
- CI/CD pipelines with keyless signing
- Public package repositories
- Maximum security with public transparency

#### Hybrid Profile

**Purpose**: Enable oneline creation to be coupled with offline verification

**Signing Behavior**:
- Same as online profile
- Creates offline-verifiable bundles

**Verification Behavior**:
- Same as offline profile
- Uses embedded or custom trusted root

**Default Configuration**:
```go
// Hybrid Profile Defaults
SignBlobOptions{
    // Same as airgap profile
}

VerifyBlobOptions{
    KeyOpts: options.KeyOpts{
        NewBundleFormat: true,
        RekorURL: "https://rekor.sigstore.dev",
    },
    CertVerifyOptions: options.CertVerifyOptions{
        IgnoreSCT: true, // SCT still skipped (not in bundle)
    },
    IgnoreTlog: false,
    Timeout: 3 * time.Minute,
}
```

**Use Cases**:
- Development environments and published artifacts on the internet - delivered to airgapped environments.

### Supported Signing/Verification Permutations

The following table documents all supported permutations across profiles:

| Signing Method | Profile | Network Required | Transparency Log | Bundle Completeness | Verification Support |
|----------------|---------|------------------|------------------|---------------------|---------------------|
| Keypair (local file) | Offline | No | Not uploaded | Minimal (signature only) | Offline with embedded root |
| Keypair (local file) | Online | Optional | Uploaded if available | Complete with tlog entry | Online or offline |
| Keypair (local file) | Hybrid | No | Not uploaded | Minimal | Offline with optional tlog check |
| Keypair (local HSM) | Offline | No | Not uploaded | Minimal | Offline with embedded root |
| Keypair (cloud KMS) | Offline | Yes (to KMS) | Not uploaded | Minimal | Offline with embedded root |
| Keypair (cloud KMS) | Online | Yes | Uploaded | Complete | Online or offline |
| Keyless (OIDC) | Online | Yes | Uploaded | Complete with cert chain | Online or offline with cert validation |
| Keyless (OIDC) | Offline | N/A | N/A | **NOT SUPPORTED** | **NOT SUPPORTED** |

**Key Findings**:
- Offline profile cannot support keyless signing (requires Fulcio by design)
- All profiles can be verified offline if using keypair signing or via embedded or referenced trusted root
- Online profile provides maximum transparency but requires connectivity
- Cloud KMS can work in airgap if accessible without internet (private cloud)

### CLI Changes

#### New `--profile` Flag

Add profile flag to signing and verification commands:

```bash
# Signing commands
zarf package sign <package> --signing-key cosign.key [--profile offline|online|hybrid]

# Verification commands
zarf package verify <package> [--profile offline|online|hybrid]
```

Default profile: `offline`

### Trusted Root Management

#### Embedded Trusted Root Update Process

The embedded trusted root should be updated regularly to include new Sigstore keys and prevent expiration issues.

**Update Schedule**:
- Checked before major releases (required)
- Automatically with tooling (recommended)
- When Sigstore announces key rotations (required)

#### Custom Trusted Root for Private Deployments

Users could fork and build the Zarf binary while also embedding their own custom Trusted Roots for distrbution.

In doing so they would leverage the portability that comes with embedding it into the binary.

### Migration Path from Legacy Format

#### Phase 1: Dual Format Support (Current State)

**Timeline**: Current - v0.67.0

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

**Timeline**: v0.67.0 - v0.69.0

**Behavior**:
- Only `zarf.bundle.sig` created during signing
- Legacy format verification still supported for backward compatibility
- Loud deprecation warnings when verifying legacy signatures

**Actions**:
- Remove legacy signature creation from signing operations
- Update all documentation to reference bundle format only

### Risks and Mitigations

#### Risk 1: Embedded Trusted Root Expiration

**Risk**: If embedded trusted root is not updated, verification may fail due to expired keys.

**Mitigation**:
- Explicit support for including updated Trusted Roots from the command line
- Pre-release checklist includes trusted root update verification
- Monitor Sigstore announcements for key rotations
- Renovate monitoring for updates to the Trusted Root repository
- Documentation on manual trusted root updates

#### Risk 2: Bundle Format Incomplete in Offline Mode

**Risk**: Sigstore bundle may require fields that cannot be populated offline.

**Mitigation**:
- Validate minimal bundle structure works with cosign verification given offline keypair signing and verification
- Document minimal vs. complete bundle differences
- Contribute upstream to Sigstore if minimal bundle support is lacking

#### Risk 3: Profile Confusion and Misconfiguration

**Risk**: Users may choose wrong profile or misconfigure connectivity settings.

**Mitigation**:
- Default to offline profile (safest, most aligned with Zarf mission)
- Validation of profile compatibility with flags
- Examples in documentation for each profile

#### Risk 4: Transparency Log Service Availability

**Risk**: Online profile depends on public Sigstore infrastructure availability.

**Mitigation**:
- Timeouts and retry logic for network operations
- Option to disable transparency log even in online profile

## Design Details

### Implementation Phases

#### Phase 1: Profile Framework 

**Scope**:
- Implement `--profile` flag for signing and verification commands
- Add profile validation logic
- Create profile configuration structures
- Update default options based on profile selection

**Deliverables**:
- Profile enum and configuration structures
- CLI flag parsing and validation
- Profile-specific default option builders
- Unit tests for profile selection logic

#### Phase 2: Bundle-Only Signing

**Scope**:
- Keep legacy verification support for backward compatibility
- Update all signing operations to prioritize bundle creation

**Deliverables**:
- Signing operations produce a `zarf.bundle.sig`

#### Phase 3: Enhanced Verification

**Scope**:
- Implement online profile verification with transparency log
- Add hybrid profile with graceful fallback
- Add transparency log entry validation

**Deliverables**:
- Online transparency log verification working
- Hybrid profile with network error tolerance
- Comprehensive tests for all profiles
- Performance testing for online verification

#### Phase : Trusted Root Automation

**Scope**:
- Support for automated updated to Trusted Root
- Pre-release trusted root update checks
- Monitoring and alerting for Sigstore announcements

**Deliverables**:
- Automated trusted root update workflow
- CI checks for trusted root freshness
- Documentation for manual updates
- Alerting for key rotations

### Test Plan

#### Unit Tests

Unit tests will validate profile selection (defaults, overrides, validation), signing operations across all profiles and methods (keypair), verification behaviors (offline with embedded/custom roots, online with transparency logs, hybrid with/without network), signature validation (invalid signatures, tampered packages), and trusted root management (loading, overrides, error handling, priority).

#### E2E Tests

E2E tests will cover complete workflows for each profile: offline (keypair generation, signing, offline verification, deployment), online (transparency log upload and verification, bundle structure validation), hybrid (network availability toggling, graceful fallback), custom trusted roots (private Sigstore infrastructure), and keyless signing (OIDC-based signing with certificate validation).

### Graduation Criteria

#### Alpha

**Criteria**:
- Profile framework implemented and tested
- `--profile` flag available on signing and verification commands
- Offline profile works as default
- Bundle signing implemented
- Online profiles implemented
- Documentation draft available

**Exit Criteria**:
- All unit tests passing
- Basic E2E test for offline profile passing
- No critical bugs reported in profile selection

#### Beta

**Criteria**:
- Automated trusted root updates working
- Comprehensive documentation published

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

**From v0.66.x or earlier** (legacy signatures):

1. **Upgrade Zarf**: Install v0.67.0 or later
2. **Continue Using Legacy Packages**: Verification still works
3. **New Packages Automatically Use Bundle**: No action needed
4. **Migrate Existing Packages** (optional but recommended):
   ```bash
   zarf package sign old-package.tar.zst \
     --signing-key cosign.key
   ```

### Version Skew Strategy

#### Signing Version vs. Verification Version

**Scenario**: Package signed with Zarf v0.66.0, verified with Zarf v0.44.0

**Result**: Works (The legacy signature exists alongside the bundle)

**Scenario**: Package signed with Zarf v0.65.0 (legacy), verified with Zarf v0.67.0

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

Online profile creates a dependency on public Sigstore infrastructure (Rekor, Fulcio) which could become a single point of failure.

**Counter-argument**: Airgap profile (default) has zero dependencies. Online profile is opt-in for users who specifically want transparency. Private Sigstore deployments are supported.

## Alternatives

### Alternative 1: No Profiles, Flag-Based Configuration

**Description**: Use individual flags (`--ignore-tlog`, `--tlog-upload`) instead of profiles.

**Pros**:
- Maximum flexibility for advanced users
- No new concepts to learn
- Granular control over each setting

**Cons**:
- Easy to misconfigure (8+ flags to set correctly)
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
- Performance impact from network calls
- Unusable in disconnected environments without flag changes

**Rejection Reason**: Violates Zarf's core mission of supporting airgapped deployments. Default must work offline.

## Infrastructure Needed (Optional)

### Development Infrastructure

**No additional infrastructure required** - all development can use existing Zarf development environment.

### Testing Infrastructure

**CI/CD Infrastructure**:
- GitHub Actions for automated trusted root updates (already available)
- Scheduled workflow for monthly updates
- PR creation automation (already available via GitHub API)

### Production Infrastructure

**No infrastructure required** for Zarf itself. Users choosing online profile will need:

**Public Sigstore** (default for online profile):
- Provided by Sigstore project (no cost)
- Available at https://rekor.sigstore.dev and https://fulcio.sigstore.dev