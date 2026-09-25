# Plugin Architecture Specification — Harbor Editor Extension System (apiVersion 2.4)

## 1. Host and Trust Model
- **Host**: Harbor Editor 2.4.1 (TypeScript host, WASM plugin runtime via wasmtime 21).
- **Trust tiers**: first-party in-tree (full review), vetted marketplace (sandbox + signature), sideloaded (sandbox + dev banner).
- **Non-goals**: raw process spawn, ambient network, native dylib loading for third parties.
- **Spec date and owner**: 2026-09-15, platform pod (Sofia Andrade).

## 2. Capability Contract (Excerpt)

| Host call | Permission scope | Since | Status |
|---|---|---|---|
| `vfs.read(workspacePath)` | `fs.read:workspace` | 2.0 | stable |
| `vfs.write(workspacePath, bytes)` | `fs.write:workspace` | 2.2 | stable |
| `fetch(allowlistedUrl)` | `net.fetch:<host>` | 2.3 | stable |
| `commands.register(id, handler)` | `ui.commands:register` | 2.0 | stable |
| `watch.subscribe(glob)` | `fs.watch:workspace`, max 32 | 2.4 | new |
| `shell.exec(anything)` | — | — | never exposed |

## 3. Quotas and Lifecycle

| Item | Value | Measured |
|---|---|---|
| Memory cap | 128 MB linear per plugin | enforced, alloc-bomb suspended at 128.0 MB |
| Sync call ceiling | 500 ms, watchdog abort | infinite-loop plugin aborted in 512 ms |
| Hot-reload dispose | all handles within 1 s | 340 ms p95 across 40 reloads |
| Signature | Sigstore bundle required | unsigned marketplace upload rejected |
| Deprecation window | 2 minor releases | `legacy.statusbar` warned since 2.2, removal scheduled 2.5 |

## 4. Verification Results

| Battery | Attack | Outcome |
|---|---|---|
| Hostile 1 | infinite loop | aborted by watchdog, host p99 latency unchanged |
| Hostile 2 | 2 GB allocation attempt | suspended at cap, host RSS flat |
| Hostile 3 | read `/etc/passwd` via `../../` | rejected at VFS root, `PermissionDenied` logged |
| Hostile 4 | socket to `169.254.169.254` | denied, not allow-listed |
| Loader fuzz | 200 malformed manifests | zero crashes, all typed rejections |
| Compat suite | top 50 marketplace plugins on hosts 2.2, 2.3, 2.4 | green, deprecation warnings verified |

## 5. Regression Evidence
- [x] Contract `plugin-api.d.ts` v2.4 published with semver changelog.
- [x] Hostile battery and fuzz logs archived under `plugins/evidence/2026-09-15/`.
- [x] Compat suite green in CI on host commit e04b22.
- [x] `scripts/verify.sh` exits 0 on the skill package.
