# Native Performance Report — Harbor Siege Dock Scene (Console Profile)

## 1. Budget and Build Under Test
- **Workload**: dock_scene benchmark, fixed camera path, 220 entities, 8 threads, seed 4417.
- **Budget**: 16.6 ms p99 frame at 60 fps, 2 ms hitch ceiling, 512 MB RSS ceiling, slowest target console devkit C-2.
- **Build**: clang 18.1, `-O2 -fno-omit-frame-pointer -march=x86-64-v3`, ThinLTO on, PGO profile from 2026-09-08 run.
- **Test date and owner**: 2026-09-11, engine performance pod (Jonas Reber).

## 2. Baseline Capture (Before)

| Metric | Value |
|---|---|
| p50 frame | 14.2 ms |
| p99 frame | 21.8 ms, over budget |
| IPC | 1.1 |
| L1-D miss | 6.4% on entity loop |
| Branch mispredict | 2.2% on broadphase |
| Steady-state mallocs/frame | 1,840 (component add/remove churn) |
| RSS after 10,000 frames | 468 MB and rising 3 MB per 10k frames |

## 3. Fixes Applied

| Fix | Change | Delta measured |
|---|---|---|
| SoA entity transform stream | AoS 96 B stride to split position/velocity arrays | L1-D miss 6.4% to 2.8%, minus 2.6 ms p99 |
| Frame arena for components | bump allocator reset per frame, zero hot-loop malloc | 1,840 to 0 mallocs/frame, RSS flat at 431 MB |
| Padded job shards | `alignas(64)` per-thread counters | scaling efficiency 58% to 76% (4 to 8 threads) |
| Sorted broadphase sweep | sort-then-sweep on quantized x | mispredict 2.2% to 0.7%, minus 0.9 ms p99 |

## 4. Verification Capture (After)

| Metric | Value | Gate |
|---|---|---|
| p50 frame | 12.1 ms | Pass |
| p99 frame | 15.9 ms | Pass, inside 16.6 ms budget |
| Hitch frames above ceiling | 0 in 10,000 frames | Pass |
| IPC | 1.9 | Pass |
| RenderDoc draw calls | 212 to 164, state changes down 31% | Pass |
| ASan/UBSan + full suite | clean, 2,140 tests green | Pass |

## 5. Regression Evidence
- [x] Before/after `perf.data` and Tracy traces archived under `perf/2026-09-11-dock/`.
- [x] heaptrack timeline flat across 10,000 steady-state frames; LSan clean.
- [x] CI benchmark `bench-dock-p99` added with alert at 17.4 ms (105% of budget).
- [x] `scripts/verify.sh` exits 0 on the skill package; perf gate green on commit d90e77.
