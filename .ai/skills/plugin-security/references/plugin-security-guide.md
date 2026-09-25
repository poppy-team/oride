# Plugin & Extension Security — Technical Reference Guide

## 1. Core Concepts

**Capability-based isolation** means a plugin holds only explicit, unforgeable handles to host resources. The **broker** is the single mediation point: every file, network, clipboard, or shell request crosses it with an allow/deny check against the plugin's declared manifest. The **sandbox** is the enforcement boundary — a WASM module with linear-memory isolation or a separate OS process under seccomp/AppArmor.

**Manifest** declares capabilities in a versioned schema (e.g. `fs.read: ["workspace/**"]`, `net.fetch: ["https://api.example.com"]`). Anything undeclared is denied. Permission expansion on update re-enters review and re-prompts the user.

**Provenance** chains signature → pinned hash → update channel: a signed manifest whose content hash matches the loaded bundle, fetched over a pinned channel. Breaks in the chain quarantine the plugin.

## 2. Patterns

- **Broker with audit log**: `broker.request(plugin_id, capability, args)` returns allow/deny and appends `{plugin_id, capability, args_hash, verdict, timestamp}` to an append-only log for the capability matrix.
- **Path-jail file access**: resolve plugin paths with `filepath.Clean` + prefix check against the granted subtree; reject symlinks escaping the jail.
- **SSRF guard on plugin fetch**: resolve DNS, reject private/loopback/link-local ranges and cloud metadata IPs (`169.254.169.254`), disable redirects by default.
- **Crash isolation**: out-of-process plugins supervised with restart budgets (max 3 restarts / 5 min); a crashing plugin never takes down the host.

## 3. Anti-Patterns

- Wildcard capabilities (`fs:*`, `net:*`) or ambient authority (inheriting host env vars, tokens, cwd).
- Loading unsigned bundles from mutable URLs with no hash pin.
- Direct FFI / raw sockets / unsandboxed `eval` reachable from plugin code.
- Auto-installing updates that add capabilities without re-consent.

## 4. Worked Example

Plugin `csv-importer@1.4.0` declares `fs.read: ["workspace/uploads/**"]` and no network capability. Broker log shows 212 `fs.read` allows and 3 `net.fetch` denies when the plugin attempts `https://telemetry.example.net/beacon` — the exfiltration path is blocked at the broker and the finding ships as HIGH with log excerpts. Manifest tightened to remove an unused `clipboard.read` declaration; re-test shows zero denies under normal use.

## 5. Verification Pointers

- Attempt `../../etc/passwd` traversal through every plugin file API; expect deny.
- Attempt metadata-IP fetch through plugin network APIs; expect deny.
- Kill the plugin process mid-request; expect host continuity and supervisor restart within budget.
