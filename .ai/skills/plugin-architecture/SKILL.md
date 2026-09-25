# Plugin Architecture

## Purpose
Design and implement sandboxed extension points with versioned contracts and explicit permissions: capability-scoped plugin APIs, semver-guaranteed ABI stability, isolated execution with resource quotas, and safe hot-reload lifecycles for third-party and first-party extensions.

## Use when
- Adding extension points (commands, themes, file handlers, language backends) to an editor, CLI, game engine, or server.
- Defining a plugin API contract with versioning, deprecation windows, and compatibility tests.
- Sandboxing untrusted plugin code with permission manifests, filesystem/network scoping, and CPU/memory quotas.
- Supporting install, update, disable, and hot-reload lifecycles without host restarts.

## Do not use when
- Exposing tools over Model Context Protocol to AI clients (use `mcp-integration` for the server surface, `mcp-tooling` for SDK work).
- Reviewing dependency supply-chain risk without extension execution (use `supply-chain-security`).
- Scripting one-off internal automation with no third-party code boundary (a plain module suffices; no plugin system needed).

## Required context
- Host capabilities to expose (command registry, file events, editor buffers) and explicit non-goals (no raw process spawn, no ambient network).
- Trust tiers: first-party (in-tree), vetted marketplace, and untrusted sideloaded plugins with per-tier sandbox depth.
- ABI and distribution constraints: language (WASM, JS isolated-vm, native dylib), packaging format, and signature requirements.
- Compatibility promise: semver policy, deprecation window (e.g., 2 minor releases), and LTS host versions under test.

## Procedure
1. **Declare the capability contract first**: Write `plugin-api.d.ts` (or IDL) listing every host function a plugin may call, each tagged with its permission (`fs.read`, `net.fetch`, `ui.commands`); anything unlisted is unreachable by construction.
2. **Enforce permission manifests**: Every plugin ships `plugin.json` declaring requested permissions with justification; the host prompts once at install and denies undeclared calls at runtime with a logged `PermissionDenied` error naming the missing scope.
3. **Sandbox execution with quotas**: Run plugins in WASM sandboxes (or isolated JS realms) with 128 MB memory and 500 ms synchronous-call ceilings; filesystem access goes through a capability-scoped VFS rooted at the workspace, never the real cwd.
4. **Version and deprecate explicitly**: Gate API surface by `apiVersion` (currently `2.4`); breaking changes bump major, additions bump minor, and removals ship only after 2 minor releases of `@deprecated` warnings verified by the compat suite.
5. **Harden the lifecycle**: Implement install (signature verify via Sigstore), enable/disable without restart, and hot-reload that disposes old handles (timers, listeners, file watches) within 1 s; fuzz the loader with 200 malformed manifests and assert zero host crashes.
6. **Prove isolation and verify**: Run the hostile-plugin battery (infinite loop, 2 GB allocation attempt, path-traversal read of `/etc/passwd`, socket to `169.254.169.254`) and assert all four are contained with host uptime intact; run `scripts/verify.sh` to confirm artifact completeness.

## Decision rules
- **Capabilities, not ambient authority**: Plugins receive only declared handles; ambient filesystem, network, and process APIs are never exposed.
- **Manifest lies fail closed**: A permission used but not declared throws at call time; a permission declared but unused warns at review.
- **No silent breaking changes**: Any host change breaking a compat-suite plugin blocks the release until migrated or reverted.
- **Hot-reload must dispose**: Leaked timers, listeners, or file watches across reload fail the lifecycle gate.
- **Signatures required for distribution**: Marketplace and auto-update channels accept only Sigstore-signed bundles; unsigned loads are dev-mode only with a banner.

## Evidence required
- API contract (IDL/types) with permission tags and semver changelog for the release.
- Hostile-plugin battery log: all four attacks contained, host uptime and p99 command latency unaffected.
- Compat-suite report across declared host versions with deprecation warnings verified.
- Loader fuzz log: 200 malformed manifests, zero host crashes, all rejected with typed errors.
- Completed `templates/plugin-architecture-spec.md` with contract and verdicts.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Plugin architecture specification with capability contract, permission model, and lifecycle guarantees.
- Sandboxed loader, manifest validator, and permission enforcement with hostile-battery tests.
- Compat suite wired into CI across declared host versions.

## Stop conditions
- Hostile battery fully contained and loader fuzz clean with typed rejections.
- Compat suite green across declared host versions with no silent breakage.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the security owner if a sandbox escape, signature bypass, or permission-enforcement gap is found; freeze distribution until fixed.
- Escalate to product when a requested plugin capability (e.g., raw sockets) conflicts with the sandbox policy; never widen the sandbox silently.
- Escalate to the lead architect if ABI stability blocks a needed host refactor; negotiate a major bump with migration tooling.
