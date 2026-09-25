# Native Performance Optimization — Verification Checklist

## 1. Workload Reproduction & Baselines
- [ ] Regressing workload reproduces deterministically (seeded scene, fixed camera path, scripted inputs).
- [ ] Baseline captured on the slowest declared target: p50/p99 frame time, IPC, L1-D miss rate recorded.
- [ ] Build profile pinned: compiler version, flags, LTO/PGO status, and target microarchitecture documented.
- [ ] Thermal and power state controlled (plugged in, fans nominal); throttled runs discarded and re-captured.

## 2. CPU Hotspots & Cache Behavior
- [ ] Top-down analysis performed; the largest cycle category (retiring, bad speculation, frontend, backend) attacked first.
- [ ] No effort spent on functions below 2% of total cycles without a written justification.
- [ ] Hot structs fit 64-byte cache lines; pointer-chasing containers replaced with flat arrays on hot paths.
- [ ] L1-D miss rate on hot loops below 3%; branch mispredict rate below 1% on hot branches.
- [ ] False sharing eliminated: shared atomics and job structs padded to cache-line boundaries.

## 3. Allocation & Memory Discipline
- [ ] Zero per-frame heap allocations in steady state (heaptrack allocation timeline flat across 10,000 frames).
- [ ] Arenas/pools own transient frame data with explicit reset points; no `malloc`/`new` in hot loops.
- [ ] RSS growth below 1 MB per 10,000 steady-state frames; leak check clean under AddressSanitizer.
- [ ] VRAM budget respected: textures within streaming pool, no redundant uploads per frame.

## 4. Threading, SIMD & GPU
- [ ] Job-system scaling efficiency above 70% from 4 to 8 threads on the target CPU.
- [ ] Vectorized loops confirmed via `-Rpass=loop-vectorize`; scalar fallbacks documented where the compiler refuses.
- [ ] RenderDoc capture shows draw-call, state-change, and overdraw reductions; no redundant pipeline flushes.
- [ ] Frame pacing holds: zero frames above the 2 ms hitch ceiling in the verification run.

## 5. Regression Evidence & Safety
- [ ] Before/after captures archived with counter deltas (cycles, cache-misses, branch-misses, page faults).
- [ ] Full test suite plus ASan/UBSan green on the optimized build.
- [ ] CI benchmark tracks p99 frame time per merge with alert threshold at 105% of budget.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
