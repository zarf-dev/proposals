# ZEP-0058: Direct Exec Mode for Component Actions

## Summary

Introduce an opt-in `noShell: true` field on component actions that executes commands directly via `exec` syscall rather than wrapping them in `sh -e -c`. This enables hardened deployment environments where no shell binary is present, reducing attack surface for programmatic and unattended Zarf operations.

## Motivation

Zarf currently wraps all action commands in a shell invocation (`sh -e -c "<cmd>"`). This is a reasonable default for interactive and general-purpose usage where users expect shell features (globbing, piping, variable expansion). However, it creates problems for hardened deployment patterns:

1. **Mandatory shell dependency**: Any component with actions (e.g., `git-server` running `./zarf internal update-gitea-pvc`) requires `/bin/sh` to be present in the execution environment. This forces operators to ship a shell in containers that would otherwise be minimal (`scratch` or distroless).

2. **Unnecessary attack surface**: A shell binary in a production container enables shell injection, interactive access via `kubectl exec`, and pivot opportunities for an attacker who gains code execution. In air-gapped and security-sensitive environments, this is unacceptable.

3. **Implicit behavior**: Commands that are simple binary invocations (no pipes, no globbing, no variable expansion) gain nothing from shell wrapping but inherit all its risks.

Zarf is increasingly used in automated, unattended deployment pipelines (Kubernetes Jobs, GitOps controllers, CI/CD) where the execution environment should be as locked down as possible. The current architecture prevents this.

### Goals

- Allow action authors to opt out of shell wrapping on a per-action basis.
- Enable Zarf to operate in shell-free environments when all actions in a package use direct exec mode.
- Maintain full backward compatibility; shell wrapping remains the default.
- Provide clear error messages when `noShell` is used with commands that require shell features.

### Non-Goals

- Removing shell support or changing the default behavior for existing packages.
- Implementing a full command parser that handles quoting, escaping, or complex argument splitting equivalent to a shell.
- Addressing shell availability for `wait` actions (these do not invoke shell commands).

## Proposal

### User Stories

#### Story 1: Hardened Initializer Container

As a platform operator deploying Zarf in an air-gapped environment via a Kubernetes Job, I want to run the initializer from a `scratch`-based container image containing only the Zarf binary and the init package. Today this fails because the `git-server` component's actions require `/bin/sh`. With `noShell: true`, the init package's internal actions can execute directly, eliminating the need for a shell binary in the container.

#### Story 2: Security-Compliant Package Authoring

As a package author building for environments that enforce CIS benchmarks or STIG compliance, I want to ensure my package does not require or invoke a shell at any point during deployment. With direct exec mode, I can author actions that invoke binaries directly and guarantee no shell interpreter is involved in my deployment pipeline.

#### Story 3: Programmatic Zarf Usage

As an operator running Zarf programmatically (embedded in controllers, operators, or automation tooling), I want to ensure that action execution is deterministic and free from shell interpretation side effects. Direct exec mode provides this guarantee.

### API Change

Add a `noShell` field to `ZarfComponentAction`:

```go
type ZarfComponentAction struct {
    // ...existing fields...

    // (cmd only) Execute the command directly without shell wrapping.
    // When true, the cmd string is split on whitespace into a binary and arguments,
    // then invoked directly via exec. Shell features (pipes, globbing, variable
    // expansion) are not available.
    NoShell *bool `json:"noShell,omitempty"`
}
```

Add a corresponding field to `ZarfComponentActionDefaults` to allow setting this at the action-set level:

```go
type ZarfComponentActionDefaults struct {
    // ...existing fields...

    // (cmd only) Execute all commands in this action set directly without shell wrapping.
    NoShell bool `json:"noShell,omitempty"`
}
```

### Schema Example

```yaml
components:
  - name: git-server
    actions:
      onDeploy:
        defaults:
          noShell: true
        before:
          - cmd: ./zarf internal update-gitea-pvc
          - cmd: ./zarf internal update-gitea-pvc --rollback
```

Per-action override (opt back into shell for a specific action):

```yaml
components:
  - name: example
    actions:
      onDeploy:
        defaults:
          noShell: true
        before:
          - cmd: ./binary --flag value
          - cmd: echo "this needs a shell" | grep shell
            noShell: false
```

### Execution Semantics

When `noShell: true` is active:

1. The `cmd` string is split into tokens using Go's `strings.Fields()` (splits on whitespace).
2. The first token is the binary path.
3. Remaining tokens are passed as arguments.
4. The command is executed directly via `exec.Command(binary, args...)`.
5. No shell metacharacters are interpreted (`|`, `>`, `<`, `*`, `$`, etc.).
6. If the binary is not found, the error should clearly indicate that the command was run without a shell.

### Risks and Mitigations

| Risk                                                              | Mitigation                                                                                                                                                                                                  |
| ----------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Users set `noShell: true` on commands that rely on shell features | Validation warning at `zarf package create` time if cmd contains shell metacharacters and `noShell` is set. Clear error message at runtime.                                                                 |
| Whitespace splitting is insufficient for complex arguments        | Document that arguments with spaces are not supported in `noShell` mode. For complex cases, users should create a wrapper binary or use shell mode. Future enhancement could support an `args` array field. |
| Breaking init package for environments that still have a shell    | No breaking change; this is opt-in. The init package can adopt `noShell` for its internal actions without affecting user-authored packages.                                                                 |

## Design Details

### Implementation

The change is localized to `src/pkg/packager/actions/actions.go` in the `actionRun` function:

```go
func actionRun(ctx context.Context, cfg v1alpha1.ZarfComponentActionDefaults, action v1alpha1.ZarfComponentAction) (string, string, error) {
    l := logger.From(ctx)
    start := time.Now()

    execCfg := exec.Config{
        Env:   cfg.Env,
        Dir:   cfg.Dir,
        Print: !cfg.Mute,
    }

    noShell := cfg.NoShell
    if action.NoShell != nil {
        noShell = *action.NoShell
    }

    var stdout, stderr string
    var err error

    if noShell {
        parts := strings.Fields(action.Cmd)
        if len(parts) == 0 {
            return "", "", errors.New("noShell action has empty cmd")
        }
        l.Debug("running command (direct exec)", "bin", parts[0], "args", parts[1:])
        stdout, stderr, err = exec.CmdWithContext(ctx, execCfg, parts[0], parts[1:]...)
    } else {
        shell, shellArgs := exec.GetOSShell(cfg.Shell)
        l.Debug("running command", "shell", shell, "cmd", action.Cmd)
        stdout, stderr, err = exec.CmdWithContext(ctx, execCfg, shell, append(shellArgs, action.Cmd)...)
    }

    if !cfg.Mute {
        l.Debug("command complete", "stdout", stdout, "stderr", stderr, "duration", time.Since(start))
    }
    return stdout, stderr, err
}
```

No changes are needed to `exec.CmdWithContext` itself since it already supports arbitrary binary+args invocation.

### Test Plan

- **Unit tests**: Verify `actionRun` with `noShell: true` executes binaries directly without invoking a shell.
- **Unit tests**: Verify `noShell` per-action override takes precedence over defaults.
- **Integration test**: Deploy a component with `noShell: true` actions in an environment without `/bin/sh` (e.g., a distroless container) and verify successful execution.
- **Integration test**: Verify that shell metacharacters in a `noShell` command are passed literally (not interpreted).
- **Negative test**: Verify clear error when binary is not found in `noShell` mode.

### Graduation Criteria

#### Alpha

- `noShell` field added to schema (behind feature flag if desired).
- Unit and integration tests passing.
- Documentation updated with usage examples and limitations.

#### Beta

- Init package internal actions (`update-gitea-pvc`, etc.) migrated to use `noShell: true`.
- Validation warning for shell metacharacters in `noShell` commands during `zarf package create`.
- Feedback collected from operators running in hardened environments.

#### Stable

- Feature flag removed (if one was used in alpha).
- Consider adding an `args` array field as an alternative to whitespace splitting for complex argument cases.

### Upgrade / Downgrade Strategy

- **Upgrade**: No action required. Existing packages without `noShell` continue to use shell wrapping. The field is additive.
- **Downgrade**: Packages authored with `noShell: true` will fail schema validation on older Zarf versions. This is expected and documented.

### Version Skew Strategy

No cluster-side components are affected. The change is entirely in the Zarf CLI binary's action execution path.

## Drawbacks

- Adds another field to the action schema, increasing surface area.
- `strings.Fields()` splitting is limited compared to proper shell argument parsing (no quoting support). This is intentional to keep the feature simple and predictable, but may surprise users who expect `"quoted args"` to work.
- Requires init package changes to benefit the hardened-container use case (though these changes are trivial).

## Alternatives

### 1. Use `shell.linux: ""` to mean "no shell"

Overloading the empty string to mean "disable shell" would be a breaking change since the current behavior treats empty as "use default (sh)". It also conflates "shell preference" with "shell presence".

### 2. Require a shell in all environments

This is the status quo. It forces operators to ship shell binaries in containers that otherwise have no need for them, expanding the attack surface unnecessarily.

### 3. Add an `args` array field instead of `noShell`

```yaml
- bin: ./zarf
  args: ["internal", "update-gitea-pvc"]
```

This is more explicit but represents a larger schema change and diverges from the current `cmd`-string-based approach. It could be considered as a future enhancement on top of `noShell`.

### 4. Ship a minimal shell (busybox) in hardened containers

This is the current workaround. It works but defeats the purpose of running a minimal container: any shell binary enables interactive access, shell injection, and expands the trust boundary. It is a compromise, not a solution.

## Infrastructure Needed

None. The change is entirely within the existing Zarf codebase and CI infrastructure.

## Implementation History

- 2026-06-29: Initial proposal drafted.
