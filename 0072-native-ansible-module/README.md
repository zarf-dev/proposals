# ZEP-0072: Native Zarf Ansible Module

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
  - [The Zarf Binary as the Module Entrypoint](#the-zarf-binary-as-the-module-entrypoint)
  - [Module Dispatch and Convention](#module-dispatch-and-convention)
    - [Two Dispatch Entry Points, One Registry](#two-dispatch-entry-points-one-registry)
    - [Why Not `zarf internal ansible-module`?](#why-not-zarf-internal-ansible-module)
  - [Inventory Projection and Thin Action Plugin](#inventory-projection-and-thin-action-plugin)
    - [Bypassing Standard Module Staging (`_execute_module`)](#bypassing-standard-module-staging-_execute_module)
    - [Argument Delivery When Bypassing `_execute_module()`](#argument-delivery-when-bypassing-_execute_module)
  - [Stdout Protection and WANT_JSON Contract](#stdout-protection-and-want_json-contract)
  - [Check Mode and Idempotency Signaling](#check-mode-and-idempotency-signaling)
  - [Live Phase Heartbeat and Execution Monitoring](#live-phase-heartbeat-and-execution-monitoring)
    - [Why Not Ansible's Built-In `async`/`poll`?](#why-not-ansibles-built-in-asyncpoll)
  - [Test Plan](#test-plan)
      - [Prerequisite testing updates](#prerequisite-testing-updates)
      - [Unit tests](#unit-tests)
      - [e2e tests](#e2e-tests)
  - [Graduation Criteria](#graduation-criteria)
  - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
  - [Version Skew Strategy](#version-skew-strategy)
  - [Collection Packaging and Distribution](#collection-packaging-and-distribution)
- [Implementation History](#implementation-history)
- [Drawbacks](#drawbacks)
- [Alternatives](#alternatives)
- [Future Work / Follow-Up ZEPs](#future-work--follow-up-zeps)
- [Infrastructure Needed (Optional)](#infrastructure-needed-optional)
<!-- /toc -->

## Summary

This ZEP proposes adding native Ansible integration to Zarf by having the compiled `zarf` binary itself serve as an Ansible module run on the management/controller node, paired with a companion Ansible collection containing thin action plugins. Instead of wrapping Zarf with external Python scripts or distributing binaries to every managed node, Zarf detects module invocations via `argv[0]` (e.g., `zarf_package_deploy`) and environment variable overrides. The module reads task parameters via Ansible's [WANT_JSON protocol](https://docs.ansible.com/ansible/latest/dev_guide/developing_program_flow_modules.html#non-native-want-json-modules), executes the underlying command pipeline via internal CLI argument synthesis without fork/exec overhead, shields stdout by redirecting unexpected terminal writes to stderr, and writes a single structured JSON response to file descriptor 1.

Ansible inventory resolution (`groups` and `hostvars`) is handled controller-side by a minimal Python action plugin that projects an allowlist of connection and `zarf_*` variables into JSON arguments, delegating cluster role mapping, schema validation, and deployment execution entirely to Go.

## Motivation

Many enterprise and air-gapped environments use Ansible as an infrastructure orchestrator to prepare hosts, manage networks, and coordinate deployments. Today, driving Zarf from Ansible requires wrapping the CLI in generic `ansible.builtin.command` or `ansible.builtin.shell` tasks, or maintaining ad-hoc Python wrapper modules. This introduces several major pain points:

1. **Interpreter Spread and Python Dependencies**: Air-gapped management VMs span OS distributions ranging from RHEL 8 (Python 3.6 platform-python) to modern RHEL and Ubuntu (Python 3.12+). Python wrappers must navigate fragile virtualenvs, missing PyPI packages, and interpreter incompatibilities in disconnected networks.
2. **Duplicate Fleet Definitions**: Operators must either maintain separate Zarf cluster/package definitions and Ansible inventories, or write bespoke templates in playbooks to translate inventory variables into Zarf CLI arguments.
3. **Loss of Native Ansible Semantics**: Shelling out loses structured changed/failed status, makes `--check` (dry run) implementation cumbersome, and pollutes play logs when terminal formatting or progress spinners write to stdout.
4. **Fragile Secret and Output Handling**: Shell commands risk leaking sensitive variables into process tables and log aggregators.

By adopting a single-binary module architecture, Zarf can be called natively as an Ansible module from the statically linked binary already staged on the management VM, honoring air-gapped constraints while providing first-class Ansible integration.

### Goals

- Allow the standard `zarf` binary to act as native Ansible modules for lifecycle actions (`zarf_package_deploy`, `zarf_package_remove`, `zarf_init`) without requiring Python on managed target nodes.
- Implement Ansible's `WANT_JSON` module protocol directly within Zarf, outputting clean, structured JSON and preserving process exit code 0 on handled module failures.
- Guard stdout by duplicating fd 1 and diverting any stray CLI logger/pterm writes to stderr to prevent corrupting Ansible's JSON parser.
- Map Ansible check mode (`--check` / `_ansible_check_mode`) directly to Zarf dry-run workflows.
- Provide a companion Ansible collection (`zarf_dev.zarf`, used as the working name throughout this ZEP; not a hard requirement and may be revisited during review) featuring a thin action plugin that projects Ansible `groups` and `hostvars` (under `zarf_*` prefixes) into module parameters.
- Provide honest idempotency signaling (`changed: true/false`, with `complete`, `partial`, or `unknown` signal clarity) derived from component/phase execution outcomes.
- Provide live execution observability during long-running tasks via an out-of-band heartbeat status file monitored by the action plugin and rendered using Ansible's `Display` subsystem.

### Non-Goals

- Distributing the `zarf` binary to every managed node in the cluster or having Ansible execute tasks per-node in parallel across the cluster. Zarf remains an orchestrator running from the management node.
- Writing heavy Python business logic or inventory translation in Python. All inventory resolution logic, schema validation, and execution rules live in Go.
- Supporting packaging commands (like `zarf package create`) as Ansible modules. Package creation is a build artifact step belonging in CI pipelines, not fleet convergence tasks.
- Parsing raw Ansible inventory files (`hosts.ini`, `hosts.yaml`) inside Go. The action plugin extracts already-resolved `groups` and `hostvars` directly from `task_vars`.

## Proposal

The proposed solution introduces an internal package (`src/internal/ansiblemod` and integration in `src/cmd/ansible.go`) that inspects `argv[0]` and `ZARF_ANSIBLE_MODULE`. When invoked under a module alias (such as `zarf_package_deploy`, `zarf_init`, or `zarf_package_remove`), Zarf enters module mode:

1. **Stdout Isolation**: Duplicates file descriptor 1 (`os.Stdout`), redirects `os.Stdout` to `os.Stderr`, and reserves the original descriptor exclusively for emitting the final JSON result.
2. **Parameter Intake**: Reads the JSON arguments file path passed by Ansible (`argv[1]`), deserializing common and module-specific fields.
3. **Inventory Translation**: Reads the projected inventory document (`groups` and `hostvars`), validates mapped node roles and `zarf_*` configuration variables, and writes an ephemeral temporary configuration for the deployment phase.
4. **Command Execution**: Builds the equivalent command line vector and executes it through `cmd.ExecuteArgs` (or internal action runners), capturing phase execution details via a context-backed result sink.
5. **Result Emission**: Produces a standardized JSON response matching Ansible's contract (`changed`, `failed`, `skipped`, `msg`, and nested `zarf` diagnostic details).
6. **Live Heartbeat Status**: Optionally writes atomic heartbeat progress updates to a temporary status file (configured via `ZARF_STATUS_FILE`), enabling concurrent background monitoring and status updates by the action plugin without polluting task streams.

A companion collection supplies symlinks to the `zarf` binary and provides a thin `ActionBase` plugin that extracts `task_vars['groups']` and an allowlisted subset of `task_vars['hostvars']` (specifically `ansible_*` connection variables and `zarf_*` keys).

### User Stories (Optional)

#### Story 1

**As** an infrastructure automation engineer
**I want** to deploy Zarf packages across air-gapped clusters using an existing Ansible playbook
**so that** I do not have to write fragile `shell: zarf package deploy ...` tasks or maintain separate host inventory definitions.

```yaml
- name: Deploy base infrastructure package
  hosts: localhost
  connection: local
  tasks:
    - name: Deploy cluster monitoring
      zarf_dev.zarf.zarf_package_deploy:
        package: /opt/staged/zarf-package-monitoring-amd64.tar.zst
        values:
          - /etc/zarf/custom-values.yaml
        confirm: true
      register: deploy_result

    - name: Print deployment summary
      ansible.builtin.debug:
        var: deploy_result.zarf.phasesRan
```

#### Story 2

**As** an air-gapped platform operator running a compliance check
**I want** to execute playbooks with `--check` against Zarf deployments
**so that** Zarf dry-runs the components and indicates whether changes are pending without altering cluster state or writing live resources.

### Risks and Mitigations

| Path                                   | What it holds                                                                                                                                                                                                                       |
| -------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Stdout Pollution`                     | fd 1 is duplicated immediately upon entry, and Go's `os.Stdout` is pointed to `os.Stderr`. Any stray `pterm`, `fmt.Print`, or third-party log writes flow to stderr, which Ansible records as diagnostics.                          |
| `Secret Leakage`                       | The action plugin only forwards `ansible_*` connection parameters and explicit `zarf_*` keys. Ephemeral config files are created in restricted tempdirs (`0600`) and unlinked upon completion. Task-level `no_log: true` supported. |
| `Name Collision / Misspelled Symlinks` | Dispatch requires matching both the `zarf_` prefix and a registered, valid action name (e.g., `package_deploy`, `init`). Misspellings fail early with an informative error rather than executing as an arbitrary CLI command.       |
| `CLI Flag Drift`                       | Shared argument builders validate parameter keys. End-to-end and unit tests parse each module's full argument vector against the live Cobra command tree in `src/cmd`.                                                              |

## Design Details

### The Zarf Binary as the Module Entrypoint

Instead of shipping separate Go binaries for each action or maintaining an external Python codebase:
- The existing `zarf` binary embeds module handling directly via an entrypoint in `src/cmd/ansible.go`.
- Symlinks named `zarf_<action>` point to `/usr/bin/zarf` (packaged via `.rpm`, `.deb`, and collection link scripts).
- Ansible's `module_common.py` detects binary modules via `_is_binary()` before the AnsiballZ assembly step, so compiled Go module files (unlike new-style Python modules) are staged and executed unmodified, with no `AnsiballZ_`-prefixed wrapper filename involved. Dispatch only needs to match the plain `zarf_<action>` basename.

#### Process Entrypoint Ordering (`main()` vs `init()`)

To ensure standard output is never corrupted by transitive dependencies or CLI initialization logging, the module dispatch and file descriptor redirection must occur at the very beginning of the application lifecycle:

1. **Top-of-`main()` Interception**: `cmd.AnsibleModule(ctx, os.Args)` is invoked as the first statement in `main.go`, prior to setting up global Cobra trees, Viper config parsing, or PTerm terminal styles.
2. **Pre-Cobra Exit**: If an Ansible module invocation is detected, the binary executes the module workflow and terminates the process directly via `os.Exit(0)`. Cobra command routing and CLI help generators are never reached.
3. **Init-Path Audit**: Packages imported by `main.go` must not emit writes to `os.Stdout` inside package `init()` functions. A CI test will execute the binary with invalid module arguments to verify that the entrypoint emits purely valid JSON without pre-dispatch chatter.

#### Why Multiple Separate Binaries Is an Anti-Pattern

An alternative considered was compiling separate dedicated binaries for each module action (e.g., `zarf-module-deploy`, `zarf-module-remove`, `zarf-module-init`). While adding multiple build targets to GoReleaser may seem straightforward, it introduces severe operational, maintenance, and security burdens:

1. **Air-Gap Airlock and Supply Chain Multiplication**: Every release artifact in Zarf carries Cosign signatures, Software Bills of Materials (SBOMs), SLSA provenance attestations, and downstream package builds (RPM, DEB, APK). Multiplying the binary count multiplies the supply-chain verification and airlock transfer burden for air-gapped platform administrators, who must ingest, scan, sign, and verify each binary artifact across security boundaries.
2. **Version Skew and Inconsistent Runtimes**: Distributing independent module binaries makes it possible for the CLI binary (`zarf`) and individual module binaries on the management host to drift in version. A deployment where `zarf-module-deploy` is at `v0.70.0` but the underlying CLI or other module actions are at `v0.69.0` introduces subtle, impossible-to-diagnose failure modes in production. With a single binary, the module answering Ansible is guaranteed to be the exact same binary providing the CLI.
3. **Binary Footprint and Storage Bloat**: Because Zarf vendors and embeds substantial infrastructure libraries (Helm, Kubernetes clients, Kustomize, Cosign, OCI registries), each compiled Go binary is substantial in size (often >100MB). Shipping multiple separate binaries would needlessly bloat installation packages and consume disk image space by gigabytes for identical underlying functionality. Symlinks to a single static binary cost negligible bytes.
4. **Maintenance Overhead and Code Duplication**: Maintaining separate `cmd/` packages per module requires duplicate initialization scaffolding, logger configurations, signal handlers, and Cobra/Viper abstractions. A single multi-call binary allows shared common parameter structures, uniform stdout redirection guards, and synchronized CLI flag translation.

### Module Dispatch and Convention

```go
package ansiblemod

const (
	Prefix    = "zarf_"
	EnvModule = "ZARF_ANSIBLE_MODULE"
)

var modules = map[string]action{
	"package_deploy": runPackageDeploy,
	"package_remove": runPackageRemove,
	"init":           runInit,
}
```

#### Two Dispatch Entry Points, One Registry

Module dispatch supports two distinct entry points into the same `modules` registry, because the two callers exercise the binary differently:

- **`argv[0]` inspection**: used when the binary is invoked under a `zarf_<action>` basename — either directly (e.g. a symlink resolved and executed by a caller other than this proposal's own action plugin, or manual/debugging invocation) or by any Ansible execution path that does stage the module file conventionally via `_execute_module()`. This is the standard, spec-compliant way binary/`WANT_JSON` Ansible modules identify themselves, and keeping it working means the `zarf_<action>` symlinks remain usable as ordinary Ansible modules outside this proposal's own collection (e.g. by a third party invoking them directly without the companion action plugin).
- **`ZARF_ANSIBLE_MODULE` env var**: this is the path the companion action plugin actually uses in practice, since it invokes the already-installed `/usr/bin/zarf` directly rather than going through `_execute_module()`'s per-task staging copy — see [Bypassing Standard Module Staging](#bypassing-standard-module-staging-_execute_module) for why.

Reasons the `argv[0]` path is still load-bearing rather than vestigial:

1. **Adherence to Ansible Module Execution Model**: Ansible executes non-Python binary modules by invoking the module file on the filesystem directly, passing a single positional argument containing the path to a temporary JSON arguments file (`/tmp/.../args`). Ansible does not pass subcommands or CLI flags such as `zarf ansible-module package-deploy`. Therefore, the invocation entrypoint itself — the filename under which the binary was invoked (`argv[0]`) — is the only native contract mechanism Ansible provides to identify which module action was intended, for any caller that does go through standard module staging.
2. **Transparent Symlink Multi-Call Binary**: Distributing symlinks (`zarf_package_deploy -> /usr/bin/zarf`) allows a single compiled binary to masquerade as multiple distinct Ansible modules without producing multiple release artifacts, duplicate SBOMs, or separate signatures. Inspecting `argv[0]` allows Zarf to cleanly dispatch to the targeted action runner before the Cobra command tree or CLI flag parsers are engaged.
3. **No AnsiballZ Wrapping for Binary Modules**: Ansible's `module_common.py` determines module substyle via `_is_binary()` *before* falling through to the AnsiballZ (Python-only) assembly path. A compiled Go module file is therefore staged and executed as-is — the basename Zarf sees at `argv[0]` is exactly the module filename copied into the remote/local execution temp directory (e.g. `zarf_package_deploy`), not an `AnsiballZ_`-prefixed wrapper name. Dispatch only needs to match `zarf_<action>` directly.
4. **Preventing Erroneous Invocations**: Dispatching strictly on `argv[0]` matching both the `zarf_` prefix and an exact registered action key prevents name collisions. For example, local CI binaries named `zarf_linux_amd64` or custom developer builds will not accidentally trigger module mode or emit JSON when invoked directly.
5. **Zero Impact on Standard CLI Invocations**: When `argv[0]` is simply `zarf` (or anything without a recognized `zarf_<action>` suffix), the binary bypasses module initialization entirely and executes normal interactive CLI command routing without performance penalty or altered behavior. An environment variable override (`ZARF_ANSIBLE_MODULE`) is also supported so integration tests and manual debugging can exercise module logic without creating filesystem symlinks.

If `argv[0]` or `ZARF_ANSIBLE_MODULE` matches a known action, `ansiblemod.Run` takes control before Cobra CLI routing. It executes the action, constructs the structured response, prints JSON to the duplicated fd 1, and calls `os.Exit(0)`.

#### Why Not `zarf internal ansible-module`?

Zarf already carries a family of hidden `zarf internal <subcommand>` commands (e.g. supporting the mutating webhook agent, TLS cert generation, and other cross-component plumbing) that exist in the same binary but are deliberately kept out of the documented, stable CLI surface. Given that precedent, a reasonable question is whether module dispatch should just be a normal Cobra subcommand — `zarf internal ansible-module package_deploy <args-file>` — instead of the `argv[0]`/`ZARF_ANSIBLE_MODULE` scheme described above. The answer differs by which of the two dispatch entry points is being replaced:

- **For the `ZARF_ANSIBLE_MODULE` / action-plugin-driven path, it plausibly works.** The action plugin already controls the entire invocation via `_low_level_execute_command()` (see [Bypassing Standard Module Staging](#bypassing-standard-module-staging-_execute_module)) and isn't bound by Ansible's own module-execution contract for this call. It could just as easily run `/usr/bin/zarf internal ansible-module package_deploy` with ordinary CLI args as it could set an environment variable and invoke the plain binary. Doing so would replace this proposal's bespoke `modules` registry and env-var sniffing with Cobra's own command routing, argument parsing, and generated `--help` output, and would let the existing "parse each module's full argument vector against the live Cobra command tree" test strategy (see the `CLI Flag Drift` row in [Risks and Mitigations](#risks-and-mitigations)) become structural rather than a parallel check kept in sync by hand.
- **For the `argv[0]` / genuine-Ansible-module-execution path, it does not work.** When Ansible's own `_execute_module()` staging invokes a `WANT_JSON` binary module, it executes the staged module *file* with exactly one positional argument — the path to the temp JSON args file — and nothing else (see the "Adherence to Ansible Module Execution Model" reasoning in [Two Dispatch Entry Points, One Registry](#two-dispatch-entry-points-one-registry)). There is no room in that invocation for Ansible to also pass an `internal ansible-module <action>` subcommand path; the filename Ansible invokes (`argv[0]`) is the *only* signal available to identify which action was intended for that call. Reworking this path into a Cobra subcommand would require Ansible itself to invoke `zarf` differently than its module-execution model allows, which this proposal has no ability to change.

Even restricted to the action-plugin-driven path, moving dispatch into an ordinary Cobra command has a real cost: it reintroduces the concern raised in [Process Entrypoint Ordering](#process-entrypoint-ordering-main-vs-init): by the time a Cobra `RunE` executes, global command tree setup, persistent flags, and any `PersistentPreRun`/`init()` hooks have already run, so the stdout guard can no longer rely on being the very first thing that happens in `main()`. It would need its own stdout-guard wiring scoped to that specific command, rather than inheriting the blanket top-of-`main()` interception the current design uses for both dispatch paths. There's also a stability-posture question: `zarf internal` commands are conventionally treated as unversioned plumbing between Zarf's own components, not a contract external tooling depends on, so leaning on one here would either need this specific command to be held to a higher bar than the rest of `internal`, or an explicit acceptance that this contract carries the same no-guarantees posture already discussed for `ZARF_ANSIBLE_MODULE` and `_low_level_execute_command()` in [Bypassing Standard Module Staging](#bypassing-standard-module-staging-_execute_module).

Net effect: this is a real, viable alternative for *half* of the dispatch story (the action-plugin path), not a replacement for `argv[0]` dispatch as a whole, and it trades bespoke registry/env-var code for Cobra-routing convenience at the cost of re-deriving the stdout-guard ordering guarantee for that one command.

### Inventory Projection and Thin Action Plugin

A Python action plugin (`ActionModule` inheriting from a shared base) runs in the `ansible-playbook` controller process. It:
1. Gathers `groups` and `hostvars` from Ansible's internal `task_vars`.
2. Filters `hostvars` through an allowlist:
   - `ansible_host`, `ansible_port`, `ansible_user`, `ansible_ssh_private_key_file`
   - All variables prefixed with `zarf_`
3. Merges any nested `parameters` dictionary into the top-level task args.
4. Passes the consolidated payload to the `zarf_<action>` module binary.

#### Bypassing Standard Module Staging (`_execute_module`)

Ansible's default `ActionBase._execute_module()` flow copies the module file itself into a fresh temporary execution directory (`~/.ansible/tmp/ansible-tmp-<timestamp>-<pid>/`) before invoking it, on every task run, regardless of connection type (`local` only changes which host that temp directory lives on, not whether the copy happens). Because Zarf vendors substantial infrastructure libraries and is often >100MB (see [Why Multiple Separate Binaries Is an Anti-Pattern](#why-multiple-separate-binaries-is-an-anti-pattern)), staging the module this way on every task invocation would mean copying a 100MB+ file per task, undermining the "symlinks cost negligible bytes" argument made for the installed footprint.

To avoid this, the action plugin does not call `_execute_module()`. Instead it invokes the binary already installed on the management node (`/usr/bin/zarf`) directly via `_low_level_execute_command()`, setting `ZARF_ANSIBLE_MODULE=<action>` in the command environment rather than relying on `argv[0]` symlink dispatch for this path. This is the same pattern `ansible.builtin.raw` uses to avoid module staging, so it is precedented rather than novel, but it is a deliberate tradeoff with real maintenance cost that should be tracked:

- **Reimplements most of what `_execute_module()` provides for free**: `no_log` redaction of task arguments in verbose/error output, environment variable injection, and check-mode plumbing must be handled explicitly by the action plugin rather than inherited. Getting `no_log` wrong here is a security bug, not just a functionality gap — see the `Secret Leakage` row in [Risks and Mitigations](#risks-and-mitigations).
  - `become`/privilege escalation is explicitly **not** part of this list: Zarf itself does not require host-level root or sudo to operate — it deploys and manages cluster resources through a kubeconfig, not through privileged local system calls. The action plugin therefore has no need to reimplement `become` argument wiring for Zarf's own execution. If an operator sets `become: true` on a task for unrelated host-level reasons (e.g. reading a root-owned kubeconfig), `_low_level_execute_command(cmd, sudoable=True)` still passes escalation through to the connection plugin correctly; it is simply not something this design depends on or needs to special-case.
- **Depends on an unversioned internal contract**: `_low_level_execute_command()` is documented as the expected way to build custom action plugins, but like `_execute_module()` it is not covered by `ansible-core`'s semver/deprecation policy the way public modules are. Its internal implementation has been restructured before (e.g. the Mitogen-related `_execute_module` redirect proposal) while its external calling behavior was preserved. The collection should pin and test against the specific `ansible-core` version range it supports (Ansible officially supports roughly three concurrent `ansible-core` versions at a time) and re-verify against new releases as part of CI, rather than assuming indefinite compatibility.
- **No `ansible-test sanity` coverage at the plugin layer**: this extends the gap already noted in [Drawbacks](#drawbacks) for binary modules (no `ansible-doc`/sanity support) to the action plugin itself, placing more weight on this proposal's own e2e suite for correctness.

Symlinks named `zarf_<action>` are still packaged for module discovery and naming consistency (so `ansible-doc`-adjacent tooling and playbook authors see a conventional module name), but they are not the mechanism actually staged and executed per task — the action plugin's direct invocation of the installed binary is.

#### Argument Delivery When Bypassing `_execute_module()`

Bypassing `_execute_module()` (previous section) has a knock-on consequence: the `WANT_JSON` contract's `argv[1]`-path-to-a-temp-file convention is normally populated by `_execute_module()` itself, as part of the staging it performs. Once that staging is skipped, something else has to deliver task parameters — including registry credentials, tokens, and other `zarf_*` values — from the action plugin to the invoked binary, and this proposal does not yet commit to a mechanism. Two options were considered:

- **Option A — action plugin writes its own args file**: mirrors the existing pattern already used for the inventory document (see [Ephemeral Inventory File and Secret Handling](#ephemeral-inventory-file-and-secret-handling)) — a private tempdir (`0700`) with a `0600` args file, path passed as the invoked command's positional argument. Preserves a single `WANT_JSON` wire format shared by both the action plugin's direct-invocation path and any third-party/symlink invocation path. Cost: parameters (potentially including secrets) touch disk, even briefly and with restrictive permissions, and the action plugin owns that file's full lifecycle (creation, permissioning, and guaranteed cleanup on every exit path, including a killed controller — see the "Process Cancellation Semantics" topic in [Unresolved Questions / Discussion Topics](#unresolved-questions--discussion-topics)).
- **Option B — pass args over stdin**: since the action plugin already controls the exact invocation via `_low_level_execute_command()`, it isn't bound to the file-path convention for this call path — it can pipe the JSON payload on stdin instead, and the binary reads from stdin when dispatched via `ZARF_ANSIBLE_MODULE` (vs. reading a file path from `argv[1]` when invoked under a plain `zarf_<action>` basename for `WANT_JSON`-compliant third-party callers). Secret-bearing parameters then never touch disk on the management node for the primary path — they stay in memory for the lifetime of the pipe. Cost: this makes the argument-delivery contract explicitly bifurcated (stdin for the action-plugin path, file-path for the symlink/third-party path), so the binary's parameter-intake code and its tests need to cover both, and any future non-Ansible caller relying on file-path delivery needs to know that convention does not apply to the primary Ansible-driven path.

Current leaning is toward **Option B**, since these payloads can carry registry credentials and other secrets, and keeping them in-memory-only for the action plugin's own invocation path is a stronger default than "briefly on disk with restrictive permissions," even though it costs wire-format uniformity. This is flagged as an open discussion point rather than a final decision — see the "Argument Delivery Contract for Direct Invocation" topic in [Unresolved Questions / Discussion Topics](#unresolved-questions--discussion-topics) — since Option A's simpler, single-contract shape and reuse of already-reviewed inventory-file handling code is a real, competing consideration.

#### Ephemeral Inventory File and Secret Handling

To prevent sensitive credentials (registry passwords, tokens, private keys) from lingering on multi-tenant management nodes:

- **In-Memory Preference**: When possible, deployment parameters and values are passed in-memory directly to execution runners.
- **Ephemeral Restrictive Filesystem Isolation**: When an inventory or configuration document must be written to disk for external tools, it is created in a uniquely generated private temporary directory (`os.MkdirTemp("", "zarf-ansible-")`) with mode `0700` containing files with mode `0600`.
- **Automatic Deletion and Cleanup**: Ephemeral files are removed in a deferred cleanup hook immediately upon successful run completion. If a run fails, the file path is returned in `zarf.inventoryPath` to allow operator diagnosis.
- **Playbook Redaction**: The companion collection defaults `no_log: true` on credential-bearing tasks to prevent Ansible loggers from printing decrypted arguments to stdout or central log collectors.

### Stdout Protection and WANT_JSON Contract

Ansible modules that communicate via the non-native [WANT_JSON convention](https://docs.ansible.com/ansible/latest/dev_guide/developing_program_flow_modules.html#non-native-want-json-modules) receive task parameters as the path to a temporary JSON file (`argv[1]`) and must produce exactly one valid JSON object on standard output (`stdout`) upon completion. If an unhandled library, dependency, or CLI logger writes raw text, terminal escapes, or formatting to stdout, Ansible fails immediately with a `MODULE FAILURE: unexpected output` parsing error.

```go
func newStdoutGuard() (*stdoutGuard, error) {
	stdoutCopyFd, err := unix.Dup(int(os.Stdout.Fd()))
	if err != nil {
		return nil, err
	}
	// Divert os.Stdout to os.Stderr
	if err := unix.Dup2(int(os.Stderr.Fd()), int(os.Stdout.Fd())); err != nil {
		return nil, err
	}
	return &stdoutGuard{realStdout: os.NewFile(uintptr(stdoutCopyFd), "stdout")}, nil
}
```

All standard output emitted by Zarf sub-packages (Helm, Kustomize, PTerm, etc.) is automatically directed to stderr. The final module response is written directly to `realStdout`.

### Check Mode and Idempotency Signaling

- When `_ansible_check_mode: true` is passed, Zarf sets its internal dry-run flag. Actions lacking dry-run capabilities set `skipped: true` and inform Ansible that check mode is unsupported.
- The module response schema aligns with Ansible expectations:

```json
{
  "changed": true,
  "failed": false,
  "msg": "deployed package successfully",
  "zarf": {
    "module": "package_deploy",
    "checkMode": false,
    "command": ["zarf", "package", "deploy", "/path/to/pkg.tar.zst", "--confirm", "--no-color"],
    "phasesRan": ["ValidatePackage", "ExtractComponents", "DeployHelmCharts"],
    "changedSignal": "complete"
  }
}
```

#### Determining Ground Truth for `changed: true/false`

A module that blindly returns `changed: true` breaks Ansible handler semantics, while returning `changed: false` based on assumptions can cause operators to miss necessary reconciliation. The module implements a phased idempotency reporting model:

1. **Helm Chart Idempotency**: Prior to invoking Helm release updates, Zarf compares the target chart revision, release values, and manifest templates against the existing cluster release state. If no values or chart version changes exist and Kubernetes resources match the desired spec, Helm release actions report `changed: false`.
2. **Component & Manifest Updates**: When applying raw manifests or files, Zarf checks object existence and hashes before applying. If no cluster resources are modified or recreated, the phase reports `changed: false`.
3. **Ad-hoc Actions (`cmd: ...`)**: Arbitrary commands cannot be reliably inspected for side-effects. Phases executing un-instrumented shell commands mark their outcome as unobserved, reporting `changedSignal: "partial"` or `"unknown"`.
4. **Honest Signal Reporting (`changedSignal`)**:
   - `complete`: Every phase and component that executed explicitly verified and reported its change state.
   - `partial`: Some phases reported change state, but one or more unobserved phases (e.g., custom shell actions) ran. The unobserved phases are explicitly listed in `zarf.changedUndeclared`.
   - `unknown`: No phase reported state observations; `changed: true` is reported defensively.
5. **Check Mode (`--check`) Evaluation**: In check mode, `changed` reflects whether changes are *outstanding* (i.e., whether running for real would modify state). If the cluster matches the desired package state, check mode reports `changed: false`.

### Live Phase Heartbeat and Execution Monitoring

Zarf package deployments, cluster initializations, and image pushes often take tens of minutes over constrained network links. Under standard Ansible module execution, the controller blocks silently until the process exits, leading operators to wonder if the play is stalled or deadlocked.

Because standard output is strictly reserved for the single JSON response, real-time feedback cannot be printed directly to stdout. To resolve this:

1. **Heartbeat File Protocol**: Before invoking the module, the action plugin creates an ephemeral status file via `tempfile.mkstemp(prefix="zarf-status-", suffix=".json")` and exports its path through the environment variable `ZARF_STATUS_FILE`.
2. **Atomic Progress Reporting in Go**: During execution, an internal status/heartbeat reporter atomically writes progress updates (`.tmp.<timestamp>` renamed into place) as components and lifecycle phases begin, progress, and finish:

```json
{
  "phase": "DeployHelmCharts:monitoring",
  "index": 3,
  "total": 8,
  "status": "running",
  "timestamp": "2026-09-23T14:32:01Z"
}
```

3. **Background Monitor Thread**: The Python action plugin runs a lightweight background daemon thread (`monitor_status`) polling the heartbeat file at short intervals (e.g., every 500ms). The Go writer guarantees atomic updates by writing to a temporary file in the same directory (`.tmp.<timestamp>`) and renaming it into place. The monitor safely handles empty files and incomplete writes by catching `json.JSONDecodeError` or `ValueError` and waiting for the next tick. When valid state transitions occur, it calls Ansible's controller-side display mechanism:
   - `display.display("[zarf] Phase 3/8: DeployHelmCharts:monitoring [running]", color=C.COLOR_VERBOSE)`
   - `display.display("[zarf] Phase 3/8: DeployHelmCharts:monitoring [done]", color=C.COLOR_OK)`
   - `display.display("[zarf] Phase 3/8: DeployHelmCharts:monitoring [failed]", color=C.COLOR_ERROR)`
4. **Guaranteed Cleanup**: The action plugin safely terminates the background thread in a `finally` block and unlinks the status file, ensuring no orphaned files or threads remain regardless of task outcome.

This provides real-time CLI visibility in the operator's console during interactive playbook runs without interfering with the `WANT_JSON` contract or polluting machine-readable logs.

#### Why Not Ansible's Built-In `async`/`poll`?

Ansible already has a native mechanism for long-running tasks: `async: <seconds>` backgrounds the module on the target and returns immediately with an `ansible_job_id`, and `poll: <seconds>` (or explicit `async_status` calls) checks in periodically until the job completes or the async timeout is hit. It was considered and rejected as the primary mechanism here for a few reasons:

1. **No mid-run detail, only start/poll/done**: `async` polling only tells the controller whether the backgrounded job is still running or has finished (via a results file Ansible itself manages); it has no channel for the job to report *which* phase or component it's currently on. Getting Zarf's phase-by-phase detail (`DeployHelmCharts:monitoring [running]`) into the operator's console still requires an out-of-band signal, so `async`/`poll` alone doesn't solve the observability half of this proposal even if used.
2. **Polling cadence is coarse and playbook-authored**: `poll` intervals are set per-task in the playbook (typically 5-15s), and each poll is a discrete `async_status` call/task in Ansible's own execution loop — not a continuous background watch. The heartbeat design's 500ms in-process monitor thread gives materially tighter feedback without requiring operators to tune `async`/`poll` values per task.
3. **Complicates check-mode and structured result handling**: Async execution changes the task's return shape (the initial call returns a job-poll handle, not the final `changed`/`failed`/`msg` result), which the action plugin would need to unwrap and reconcile with this proposal's `changedSignal` idempotency contract. Keeping this module synchronous (with heartbeats as a side channel) keeps the primary WANT_JSON response the single source of truth for the task's outcome.

`async`/`poll` remains a reasonable escape hatch for operators who don't need live phase detail and are fine with coarser polling — nothing in this design prevents wrapping the module task in `async`/`poll` if desired. But it is not proposed as the default execution mode.

### Test Plan

[X] I/we understand the owners of the involved components may require updates to existing tests to make this code solid enough prior to committing the changes necessary to implement this proposal.

##### Prerequisite testing updates

Ensure test suites can execute the Zarf binary under simulated `argv[0]` aliases and test arguments files without requiring full Kubernetes clusters for contract verification.

##### Unit tests

- Test `ansiblemod.ModuleName` detection across standard basenames and `ZARF_ANSIBLE_MODULE` overrides.
- Test `ansibleinv` translation to ensure `groups` and `hostvars` map correctly to Zarf configurations, verifying that unknown `zarf_*` variables and conflicting group assignments produce descriptive errors.
- Test that `newStdoutGuard` successfully prevents text emitted to `os.Stdout` from leaking into the duplicated descriptor stream.
- Test argument vector parsing against `src/cmd` Cobra command definitions to prevent CLI flag regressions.

##### e2e tests

- Execute `zarf_package_deploy` and `zarf_package_remove` using mock JSON payloads representing `ansible-playbook` invocations.
- Verify that `--check` (check mode) sets `changed: false` or reports outstanding changes without executing deployments.
- Verify that invalid JSON payloads or missing package files return exit code 0 with `failed: true` and a valid JSON error message.

### Graduation Criteria

- **Alpha**:
  - Binary dispatch architecture implemented for `zarf_package_deploy` and `zarf_package_remove`.
  - Stdout guard and `WANT_JSON` protocol implemented and tested.
  - Basic Python action plugin published in an experimental collection.
- **Beta**:
  - Full support for `zarf_init`, check mode, and cluster idempotency signaling.
  - Collection published through whichever distribution channel is chosen (see [Collection Packaging and Distribution](#collection-packaging-and-distribution)).
  - Comprehensive e2e testing covering playbook execution in CI against live test clusters.
- **GA (Stable)**:
  - Proven real-world usage across air-gapped production platforms.
  - Stable parameter schemas and conformance tests verifying compatibility across multiple ansible-core releases.

### Upgrade / Downgrade Strategy

- **Upgrades**: Upgrading the `zarf` binary on the management node automatically updates module behavior because module files are symlinks to `/usr/bin/zarf`.
- **Downgrades**: Symlinks point to the installed binary; downgrading the Zarf package gracefully reverts module behavior without leaving orphaned runner scripts.
- **Compatibility**: Standard CLI invocations remain completely unchanged; non-Ansible users experience no difference in CLI behavior or performance.

### Version Skew Strategy

The action plugin in the Ansible collection only formats and passes task variables to the module binary. It contains no deployment logic or schema validation rules. Version skew between Ansible controller versions and the Zarf binary is therefore decoupled: the binary alone enforces schema validation for the parameters it accepts.

### Collection Packaging and Distribution

This ZEP establishes that a companion Ansible collection is needed (see [Goals](#goals)) and where its *source tree* might live (see the "Collection Governance and Repository Boundaries" topic in [Unresolved Questions / Discussion Topics](#unresolved-questions--discussion-topics)), but it does not yet commit to how the *built* collection actually reaches an operator's management node. That is a distinct question, and it interacts directly with this proposal's air-gap-first framing (see [Motivation](#motivation)): several of the usual answers for distributing an Ansible collection assume outbound network access that air-gapped environments do not have. A few options, at a high level:

- **Bundled inside the `.rpm`/`.deb` alongside the `zarf` binary**: the collection's Python/YAML files are installed to a standard Ansible collection search path (e.g. `/usr/share/ansible/collections/ansible_collections/zarf_dev/zarf`) as part of the same OS package that installs `/usr/bin/zarf`. *Pros*: one install step, and the collection and binary are guaranteed to match versions, the same guarantee the symlink-based module dispatch already relies on (see [Why Multiple Separate Binaries Is an Anti-Pattern](#why-multiple-separate-binaries-is-an-anti-pattern)); rides through the airlock as part of an artifact administrators already ingest, scan, and sign, with no separate transfer step. *Cons*: ties every collection change to a full Zarf binary release — a collection-only bugfix cannot ship independently unless a separate build/release pipeline is stood up for it; the package must install into a path Ansible actually searches by default (`ansible.cfg`'s `collections_path`, `ANSIBLE_COLLECTIONS_PATH`, or the `~/.ansible/collections` / `/usr/share/ansible/collections` defaults), or ship configuration nudging Ansible to look there, without clobbering an operator's existing collection layout.
- **Published to Ansible Galaxy (`galaxy.ansible.com`)**: the conventional distribution channel — `ansible-galaxy collection install zarf_dev.zarf`. *Pros*: familiar to Ansible users, independently versioned via `galaxy.yml`, discoverable, works with standard `requirements.yml` dependency pinning. *Cons*: assumes reachability to `galaxy.ansible.com` (or a private Galaxy-compatible index such as Automation Hub / a pulp-backed mirror), which is exactly the constraint this ZEP's target environments don't have; a disconnected operator would still need to `ansible-galaxy collection download` on a connected machine and carry the resulting tarball through the airlock by hand, which reintroduces a manual transfer step for the collection even though the binary itself no longer needs one. Plausible as a *secondary*, connected-environment-only channel rather than the primary one.
- **Standalone signed collection tarball as its own release artifact**: build `ansible-galaxy collection build` output (`zarf_dev-zarf-X.Y.Z.tar.gz`) and publish it through the same signed release pipeline as the `zarf` binary (Cosign, SBOM, GitHub Releases), letting operators run `ansible-galaxy collection install <tarball>` fully offline once the tarball has been carried through the airlock like any other artifact. *Pros*: decouples the collection's release cadence from the binary's, fits naturally alongside existing airlock ingestion workflows (it's simply another signed artifact), and remains compatible with an operator's own private/offline Galaxy mirror if they run one. *Cons*: yet another artifact for administrators to track, scan, and verify; introduces an explicit collection-version-to-binary-version compatibility matrix that must be documented and, ideally, checked at runtime (e.g. the action plugin or module refusing to run against a binary older than the collection expects).
- **Self-extracting via the `zarf` binary itself** (e.g. `zarf tools ansible-collection install`, with the collection's Python/YAML embedded in the binary via `go:embed`): a CLI subcommand lays the collection down into the standard search path on demand. *Pros*: the strongest version-match guarantee of any option (identical binary, identical collection, by construction), zero additional release artifacts, zero additional airlock transfers beyond the binary itself. *Cons*: unconventional relative to how Ansible users expect to obtain collections — not discoverable via `ansible-galaxy search`, awkward to express as a `requirements.yml` dependency for playbook portability, and would complicate ever also publishing to Galaxy for connected users without maintaining two parallel packaging paths.

These options are not necessarily mutually exclusive: for example, RPM/DEB bundling (or self-extraction) could serve as the primary air-gapped path while a Galaxy publish serves connected users as a convenience channel, at the cost of maintaining both. Whichever channel is chosen, it needs to be reconciled with the decoupled-versioning claim made in [Version Skew Strategy](#version-skew-strategy) above — that claim holds for the *runtime* protocol between the action plugin and the binary, but packaging choices that tie the collection's build/release cadence to the binary's (or don't) have their own, separate version-skew implications that should be spelled out once a channel is picked.

## Implementation History

- 2026-09-23: Initial proposal drafted.

## Drawbacks

- Requires maintaining a companion Ansible collection containing action plugins and packaging symlinks.
- Binary modules do not generate documentation via `ansible-doc` or pass standard `ansible-test sanity` without additional stubbing.
- Parameter validation must be coordinated between Go parameter structs and the collection's action plugins.

## Unresolved Questions / Discussion Topics

<<[UNRESOLVED topic="Collection Governance and Repository Boundaries" maintainers="@zarf-maintainers"]>>
**Where should the Ansible collection source tree and CI live?**
- **Option A (Monorepo in `zarf-dev/zarf`)**: Place collection files in `ansible/zarf_dev/zarf` alongside the Go codebase.
  - *Pros*: Joint PRs ensure Go parameter structures and Python action plugin projection never drift out of sync; single release tag.
  - *Cons*: Introduces Python and Ansible linting/testing requirements (`ansible-lint`, `ansible-test`) into Zarf's core Go CI pipeline.
- **Option B (Separate Repository `zarf-dev/ansible-collection-zarf`)**:
  - *Pros*: Keeps Zarf core Go repository cleanly isolated from Python toolchains; independent collection release cadence.
  - *Cons*: Risk of cross-repository version drift; requires coordinated integration tests and multi-repo release tagging.
<<[/UNRESOLVED]>>

<<[UNRESOLVED topic="Process Cancellation Semantics" maintainers="@zarf-maintainers"]>>
**What should happen to an in-flight `zarf_package_deploy`/`zarf_init` run when the Ansible controller cancels the task (Ctrl-C, `ansible-playbook` process killed, an `async` timeout is hit)?**

The current proposal does not specify this. The spawned `zarf` process is a child of the connection Ansible used to invoke it; if the controller disconnects or is killed mid-run, the child's fate depends on signal propagation through that connection (local subprocess, SSH session, etc.), which is not addressed anywhere in this design.

- **Leaning**: on cancellation, the in-progress deployment should be terminated rather than left running unattended — an orphaned `zarf` process continuing to mutate a cluster after the controller has given up on it is a worse outcome than a cleanly aborted, partially-applied deployment that a subsequent run can reconcile. This likely means the module needs to install a signal handler that attempts a best-effort abort/rollback of the current phase before exiting, and the action plugin needs to ensure signals delivered to the connection actually propagate to the child process (not always guaranteed, e.g. across some SSH configurations).
- **Open for discussion**: whether "terminate" should attempt any in-flight cleanup (e.g. rolling back a partially-applied Helm release) versus a bare process kill leaving state for the next run's idempotency logic to reconcile, and how this interacts with `changedSignal: "partial"`/`"unknown"` reporting if a run is killed mid-phase.
<<[/UNRESOLVED]>>

<<[UNRESOLVED topic="Argument Delivery Contract for Direct Invocation" maintainers="@zarf-maintainers"]>>
**How should task parameters (including secrets) reach the `zarf` binary once the action plugin bypasses `_execute_module()`'s automatic args-file staging?** See [Argument Delivery When Bypassing `_execute_module()`](#argument-delivery-when-bypassing-_execute_module) for full detail.

- **Option A**: action plugin writes its own restrictive-permission temp args file, path passed positionally — same `WANT_JSON` wire format as third-party/symlink invocations, one contract to maintain and test, at the cost of secret-bearing parameters briefly touching disk.
- **Option B**: action plugin pipes JSON parameters over stdin for its own direct-invocation path, keeping secrets in-memory only; the binary would then support two intake conventions (stdin when dispatched via `ZARF_ANSIBLE_MODULE`, file-path `argv[1]` when dispatched via `argv[0]`/`WANT_JSON`).
- **Leaning**: Option B, on the grounds that keeping registry credentials and other secrets off disk entirely for the primary (action-plugin-driven) path is a stronger default than briefly-on-disk-with-restrictive-permissions, even at the cost of a bifurcated intake contract. Not a final decision — Option A's single-contract simplicity and reuse of the already-described inventory-file handling pattern is a real competing tradeoff maintainers should weigh in on.
<<[/UNRESOLVED]>>

<<[UNRESOLVED topic="Collection Packaging and Distribution Channel" maintainers="@zarf-maintainers"]>>
**How should the built Ansible collection actually reach an operator's management node?** See [Collection Packaging and Distribution](#collection-packaging-and-distribution) for full detail on the options.

- **Option A**: bundle the collection inside the `.rpm`/`.deb` alongside the `zarf` binary — strongest version-match guarantee and no extra airlock transfer, but ties collection changes to full Zarf releases.
- **Option B**: publish to Ansible Galaxy (`galaxy.ansible.com`) — the conventional channel, but assumes network reachability this ZEP's target environments generally don't have; would need to be a secondary, connected-only channel at best.
- **Option C**: ship a standalone signed collection tarball as its own release artifact — decouples release cadence from the binary, fits existing airlock workflows, but adds another artifact and an explicit version-compatibility matrix to maintain.
- **Option D**: embed the collection in the binary itself and extract it via a `zarf tools` subcommand — matches Option A's version guarantee with zero extra artifacts, but is unconventional for Ansible users and complicates any future Galaxy publish.
- **No leaning stated**: unlike the other open topics in this document, this proposal does not currently favor one option. The choice depends heavily on how much packaging/release engineering effort maintainers are willing to take on versus how strict the air-gap-only assumption should be treated — a call best made by the maintainers rather than presumed here. Options are also not mutually exclusive (e.g. A or D as primary, B as a connected-environment convenience), which itself has a maintenance cost worth weighing.
<<[/UNRESOLVED]>>

## Alternatives

- **Multiple separate binary release artifacts**: Compiling individual Go binaries for each Ansible module action (`zarf-module-deploy`, etc.). Rejected due to the multiplied airlock/supply-chain burden (SBOMs, Cosign signatures, package builds), the risk of runtime version skew across binaries, storage and package bloat (>100MB per binary), and redundant initialization maintenance.
- **Subprocessing via Python wrapper module (`AnsibleModule`)**: Wraps `zarf` CLI calls using Python subprocesses. Rejected due to Python interpreter matrix incompatibilities in air-gapped target environments and duplicate logic between CLI output parsing and Python wrappers.
- **Copying the binary to every managed node**: Distributing `zarf` across all nodes and executing tasks per host. Rejected because Zarf is an orchestrator that manages cluster nodes centrally from a single management node; distributing per-node tasks breaks lock scoping and phase lifecycle ordering.
- **Pure shell tasks (`ansible.builtin.shell`)**: Encouraging operators to invoke `shell: zarf ...`. Rejected due to poor idempotency signaling, risk of log and credential exposure, and failure to integrate with Ansible's structured error handling.
- **Relying solely on Ansible's built-in `async`/`poll`**: Using `async: <seconds>` with `poll: <seconds>` (or explicit `async_status` checks) for long-running deploys instead of the custom heartbeat protocol. Rejected as the primary mechanism because it only exposes coarse-grained running/finished status with no channel for phase-level detail, its polling cadence is set per-task rather than continuous, and unwrapping the async job-poll result shape would complicate reconciling this proposal's `changedSignal` contract. See [Why Not Ansible's Built-In `async`/`poll`?](#why-not-ansibles-built-in-asyncpoll) for detail; it remains a valid option for operators layered on top of this design.
- **Dispatching entirely through a `zarf internal ansible-module <action>` Cobra subcommand**: Considered as a full replacement for the `argv[0]`/`ZARF_ANSIBLE_MODULE` dispatch scheme. Rejected as a full replacement because Ansible's own `_execute_module()` staging invokes binary modules with a single positional argument and no room for a subcommand path, so genuine `WANT_JSON`-contract callers still require `argv[0]`-style dispatch regardless. Remains a plausible alternative for the action-plugin-driven half of dispatch specifically (replacing just the `ZARF_ANSIBLE_MODULE` env-var path), at the cost of re-deriving the stdout-guard entrypoint ordering inside that command. See [Why Not `zarf internal ansible-module`?](#why-not-zarf-internal-ansible-module) for detail.

## Future Work / Follow-Up ZEPs

### Multi-Call Dispatch for Embedded Tooling (`kubectl`, `helm`, `k9s`, `syft`)

The `argv[0]` multi-call dispatch mechanism introduced by this ZEP establishes a generic, reliable entrypoint pattern for the Zarf binary. A natural follow-up enhancement is to extend this dispatch architecture to Zarf's vendored CLI tools (`zarf tools kubectl`, `zarf tools helm`, `zarf tools k9s`, `zarf tools syft`).

Today, air-gapped management nodes often require symlinking or scripting shims so standard CLI tools or automation can run commands like `kubectl get pods` or `helm list` without needing separate binary downloads through an airlock. By formalizing multi-call binary behavior:

- Symlinking `/usr/local/bin/kubectl -> /usr/bin/zarf` allows `argv[0] == "kubectl"` to transparently execute `zarf tools kubectl "$@"` without Cobra sub-command routing overhead or log formatting interference.
- Symlinking `/usr/local/bin/helm -> /usr/bin/zarf` allows standard Helm scripts and CI steps to seamlessly invoke Zarf's embedded, pinned Helm distribution.
- Packaging tools (`rpm`, `deb`) can optionally create these symlinks during installation on management nodes, delivering a complete "busybox for air-gapped Kubernetes" experience with zero extra release artifacts or supply chain dependencies.

A future ZEP will propose the formal routing rules, signal delegation, and packaging conventions for this embedded tool multi-call functionality.

## Infrastructure Needed (Optional)

- Depending on the outcome of the [Collection Packaging and Distribution](#collection-packaging-and-distribution) discussion: a repository or namespace in Ansible Galaxy (`zarf_dev` or `zarf`), a signed standalone collection artifact in the release pipeline, and/or packaging updates in Zarf's release pipelines (goreleaser/nfpm) to include symlinks and/or an embedded/bundled collection.
