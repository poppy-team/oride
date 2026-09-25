# Plugin Architecture — Verification Checklist

## 1. Capability Contract & Permissions
- [ ] Every host function callable by plugins is listed in the versioned API contract with a permission tag.
- [ ] No ambient authority: raw `fs`, `net`, `child_process`, or equivalent reachable from plugin code.
- [ ] `plugin.json` declares all requested permissions with justification; undeclared calls throw typed `PermissionDenied`.
- [ ] Declared-but-unused permissions warn at review; over-scoped manifests are rejected with guidance.

## 2. Sandbox & Resource Quotas
- [ ] Plugins execute in WASM sandboxes or isolated realms, never in the host process with full privileges.
- [ ] Memory capped at 128 MB and synchronous host calls at 500 ms per plugin; quota breaches suspend the plugin, not the host.
- [ ] Filesystem access scoped through a capability VFS rooted at the workspace; path traversal (`../`, absolute, symlink escape) blocked and logged.
- [ ] Network egress allow-listed per permission; metadata endpoints (`169.254.169.254`) and private ranges denied by default.

## 3. Versioning & Compatibility
- [ ] API surface gated by `apiVersion` (currently `2.4`) with semver changelog: major breaks, minor adds, patch fixes.
- [ ] Removals ship only after 2 minor releases of `@deprecated` warnings, verified by the compat suite.
- [ ] Compat suite runs green across all declared LTS host versions (2.2, 2.3, 2.4) in CI.
- [ ] Host changes breaking any compat plugin block release until migration ships.

## 4. Lifecycle, Distribution & Evidence
- [ ] Install verifies Sigstore signatures; unsigned loads work only in dev mode with a visible banner.
- [ ] Enable/disable and hot-reload complete without host restart; old handles disposed within 1 s (timers, listeners, watches).
- [ ] Loader fuzzed with 200 malformed manifests: zero host crashes, all rejected with typed errors.
- [ ] Hostile battery (loop, 2 GB alloc, `/etc/passwd` read, metadata socket) fully contained with host p99 latency intact.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
