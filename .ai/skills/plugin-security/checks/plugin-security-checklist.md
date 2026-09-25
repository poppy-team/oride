# Plugin & Extension Security — Verification Checklist

## 1. Manifest & Provenance
- [ ] Every plugin declares capabilities in the versioned manifest schema; nothing implicit.
- [ ] No wildcard capabilities (`fs:*`, `net:*`); update channels pinned.
- [ ] Manifests signed and content hashes verified; unsigned bundles quarantined by default.

## 2. Runtime Isolation & Broker Enforcement
- [ ] Plugin code runs sandboxed (WASM memory isolation or separate OS process with seccomp/AppArmor).
- [ ] Every host call crosses the capability-checked broker; deny-by-default for undeclared capabilities.
- [ ] No ambient authority: plugins receive explicit handles only, never host env, tokens, or filesystem roots.

## 3. Adversarial Abuse Cases
- [ ] Path traversal (`../../etc/passwd`, symlink escapes) blocked at every plugin file API.
- [ ] SSRF via plugin fetch blocked (private/loopback ranges, metadata IP, redirects off).
- [ ] Host-handle leakage across the bridge attempted and blocked; crash of plugin process leaves host running.

## 4. Lifecycle & Update Hygiene
- [ ] Permission-expanding updates re-enter review and re-prompt; no silent auto-install.
- [ ] Disable/revoke kills plugin processes and revokes tokens within seconds.
- [ ] Broker allow/deny log retained per plugin for the capability matrix.

## 5. Evidence & Sign-Off
- [ ] Capability matrix (declared vs exercised) with broker log excerpts recorded.
- [ ] Adversarial test results attached with severity-rated findings.
- [ ] `scripts/verify.sh` exits 0 from the repository root.
