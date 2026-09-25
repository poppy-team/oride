# Native Performance Optimization

## Purpose
Profile and optimize native CPU, GPU, and memory performance: cycle-level hotspots, cache-miss behavior, allocation churn, data-layout efficiency, and GPU pipeline stalls, with every claim backed by profiler captures before and after.

## Use when
- Diagnosing frame-time spikes, hitch budgets, or p99 latency regressions in native code (C, C++, Rust, Zig).
- Reducing L1/L2 cache misses, TLB pressure, branch mispredictions, or allocator contention.
- Optimizing memory layout (AoS to SoA), SIMD utilization, or multithreaded job-system throughput.
- Cutting GPU cost: draw calls, overdraw, pipeline state changes, or VRAM bandwidth.

## Do not use when
- Tuning browser asset caching or web vitals without native profiling (use `performance-web`).
- Reviewing cache-invalidation topologies or distributed caching layers (use `caching`).
- Fixing logical correctness bugs that merely look slow; profile first, and if the profile is flat, fix the algorithm choice elsewhere.

## Required context
- Performance budget: frame-time target (e.g., 16.6 ms at 60 fps with 2 ms hitch ceiling), p50/p99 latency SLAs, memory ceiling (e.g., 512 MB RSS on console profile).
- Reproducible workload: benchmark scene, trace capture, or load script that triggers the regression deterministically.
- Build profile under test: compiler, optimization flags (e.g., clang 18 `-O2 -fno-omit-frame-pointer`), LTO/PGO status, target microarchitecture (e.g., x86-64-v3, Apple M2).
- Profiler access: `perf`, Tracy, Superluminal, VTune, RenderDoc, or NSight availability on the target machine.

## Procedure
1. **Reproduce under the profiler**: Capture a 30-second `perf record -g -F 997` (or Tracy frame trace) of the exact regressing workload; record baseline: p50 frame 14.2 ms, p99 frame 21.8 ms, IPC 1.1, L1-D miss 6.4%.
2. **Apply top-down analysis**: Classify cycles into retiring, bad speculation, frontend-bound, backend-bound (`perf stat -e cycles,instructions,cache-misses,branch-misses`). Attack the largest category first; ignore any function below 2% of total cycles.
3. **Fix data layout before algorithms**: Convert hot AoS arrays to SoA, shrink hot structs below 64-byte cache lines, replace pointer-chasing (`std::map` node hops) with flat arrays; re-measure L1-D miss rate, targeting below 3%.
4. **Cut allocation churn**: Replace per-frame `malloc`/`new` in hot loops with arenas, pools, or bump allocators; verify with heaptrack that steady-state frame allocations drop to zero and RSS stays flat across 10,000 frames.
5. **Parallelize and vectorize deliberately**: Move independent work to the job system with cache-line-padded job structs (no false sharing), then auto-vectorize inner loops (`-Rpass=loop-vectorize` must name the loop); confirm scaling efficiency above 70% from 4 to 8 threads.
6. **Verify GPU and persist evidence**: Capture RenderDoc frame (draw calls, state changes, overdraw heatmap); assert the fix holds on the slowest declared target, store before/after captures, and run `scripts/verify.sh` to confirm artifact completeness.

## Decision rules
- **No optimization without a profile**: Changes justified by intuition alone are rejected; every hotspot fix cites a capture with symbol, cycles share, and miss counters.
- **Measure on the slowest target**: Desktop-only measurements never close a console or mobile performance gate.
- **Correctness gates stay green**: `-fsanitize=address,undefined` plus the full test suite must pass; a 5% speedup that introduces UB is a regression.
- **Hitch ceiling is absolute**: Any single frame above the 2 ms over-budget hitch limit fails, even if averages look healthy.
- **Memory growth must be flat**: Steady-state RSS growth above 1 MB per 10,000 frames blocks the change until the leak or churn source is fixed.

## Evidence required
- Before/after profiler captures (perf.data, Tracy trace, or RenderDoc capture) with p50/p99 frame times and IPC.
- `perf stat` counter deltas: cycles, instructions, cache-misses, branch-misses, page faults.
- heaptrack or sanitizer logs proving zero steady-state allocation churn and no leaks.
- Completed `templates/performance-native-spec.md` with budget table and verdicts.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Performance specification document with budgets, captures, and per-fix deltas.
- Optimized implementation with layout, allocation, and threading changes isolated per commit.
- Benchmark harness addition so CI tracks p99 frame time on every merge.

## Stop conditions
- p99 frame time within budget on the slowest declared target with hitch ceiling held.
- L1-D miss rate below 3% on hot loops and steady-state allocation churn at zero.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the lead architect if the budget requires algorithmic redesign (e.g., O(n²) broadphase must become spatial hash) rather than tuning.
- Escalate to platform owners if the bottleneck is driver, OS scheduler, or hardware errata outside application control.
- Escalate immediately if profiling reveals memory-safety violations, data races, or thermal throttling that invalidates measurements.
