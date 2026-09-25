# Plugin & Extension Security

## Purpose
Enforce capability-based plugin isolation, manifest permission verification, and memory safety boundaries so third-party extensions execute with least privilege and cannot exfiltrate data, escalate privileges, or destabilize the host.

## Use when
- Designing or auditing a plugin loading pipeline (discovery, manifest parsing, signature verification, sandbox spawn).
- Reviewing a plugin manifest for over-broad permissions, undeclared network/filesystem access, or unsafe native bridges.
- Investigating a suspected plugin escape, data exfiltration, or host crash caused by extension code.
- Operating in mode(s): `implementation`, `review`.

## Do not use when
- Vetting an entire third-party repository before onboarding (use `untrusted-project-security`).
- Scanning dependency packages for known CVEs or typosquatting (use `supply-chain-security`).
- Reviewing first-party application code with no extension boundary (use `security-review`).

## Required context
- Plugin manifest schema and the capability catalog (filesystem, network, process, device, UI surface permissions).
- Host sandbox architecture (in-process WASM, out-of-process IPC, container boundary) and enforcement points.
- Threat model for the extension ecosystem (untrusted authors, malicious updates, confused-deputy flows).

## Procedure
1. **Inventory the trust boundary**: list every host API exposed to plugins (file, network, clipboard, shell, DOM) and classify each by risk (read-only, mutating, exfiltrating, executing).
2. **Verify the manifest**: require explicit capability declarations with versioned schema; deny by default anything not declared. Reject wildcard permissions (`fs:*`, `net:*`) and unpinned update channels.
3. **Verify provenance**: require signed manifests and pinned content hashes; untrusted or unsigned plugins load only in a deny-by-default quarantine with no host capabilities.
4. **Enforce runtime isolation**: run plugin code in the sandbox (WASM memory-isolated module or separate OS process with seccomp/AppArmor profile); mediate every host call through a capability-checked broker that logs allow/deny decisions.
5. **Check update and lifecycle hygiene**: updates re-verify signatures and re-prompt on permission expansion; disable or revoke kills plugin processes and revokes tokens within seconds.
6. **Verify with adversarial tests**: attempt manifest-declared vs actual behavior mismatch, path traversal via plugin file APIs (`../../etc/passwd`), SSRF via plugin fetch, and host-handle leakage across the bridge.

## Decision rules
- **Deny by default**: any capability not explicitly declared and approved is denied at runtime, never warned-and-allowed.
- **No ambient authority**: plugins receive only explicit handles passed through the broker; inheriting host environment, tokens, or filesystem roots is prohibited.
- **Permission expansion requires re-consent**: an update requesting new capabilities must not auto-install; it re-enters review.
- **Broker mediates everything**: direct FFI, raw sockets, or unsandboxed `eval` from plugin context are forbidden.

## Evidence required
- Capability matrix: declared vs exercised permissions per plugin with broker allow/deny log excerpt.
- Adversarial test results (traversal, SSRF, escape attempts blocked).
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Security audit report (findings by severity with manifest excerpts and broker logs).
- Vulnerability remediation patches (manifest tightening, broker policy fixes, sandbox profile updates).
- Security gate evidence (signed manifest verification + capability test results).

## Stop conditions
- All in-scope plugins run under declared least-privilege capabilities with broker enforcement verified by adversarial tests.
- Unsigned or over-permissioned plugins quarantined or remediated.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to lead architect if a business-critical plugin cannot function without a high-risk capability (e.g. raw socket access).
- Escalate immediately upon discovering an active sandbox escape or data exfiltration path.
