# Native Performance Reference Guide

## 1. Core Concepts

### 1.1 Top-Down Microarchitecture Analysis
Modern CPUs report where pipeline slots go: retiring (useful work), bad speculation (mispredicted branches), frontend-bound (instruction starvation), backend-bound (data starvation). `perf stat` with `topdown` events classifies every cycle. Rule of thumb: backend-bound means fix data layout and cache behavior; bad speculation means fix branch predictability; frontend-bound means shrink code footprint (LTO, `-Os` on cold code).

### 1.2 Cache Hierarchy Math
A cache miss costs roughly 4 cycles (L1), 12 cycles (L2), 40 cycles (L3), 200+ cycles (DRAM). Effective access time:
```
T_eff = H_L1*T_L1 + (1-H_L1)*(H_L2*T_L2 + (1-H_L2)*(H_L3*T_L3 + (1-H_L3)*T_DRAM))
```
At 6.4% L1 miss with 30% L2 miss, DRAM stalls dominate frame time. Cutting L1 misses from 6.4% to 2.8% through SoA layout routinely buys back 1-2 ms per frame in entity-heavy scenes.

### 1.3 False Sharing
Two threads writing different fields of the same 64-byte cache line bounce the line between cores at ~100 ns per transfer. Pad shared atomics and per-thread job state to 64 bytes (`alignas(64)` in C++, `#[repr(align(64))]` in Rust). `perf c2c` confirms line contention before and after.

### 1.4 Allocation Churn vs Leaks
Churn (allocate/free every frame) fragments heaps and burns allocator locks; leaks grow RSS monotonically. heaptrack distinguishes them: churn shows a sawtooth allocation timeline with flat RSS, leaks show rising RSS. Fix churn with arenas reset per frame; fix leaks with ownership audits under ASan/LSan.

### 1.5 SIMD and Compiler Reports
Compilers vectorize simple stride-1 loops automatically. Verify with `-Rpass=loop-vectorize` (Clang) or `-fopt-info-vec` (GCC) rather than assuming. Pointer aliasing blocks vectorization: `__restrict` / `restrict` qualifiers and SoA layouts unlock it. Hand intrinsics (AVX2/NEON) are a last resort, justified only when the vectorizer report names an unfixable blocker.

## 2. Patterns and Anti-Patterns

| Pattern (do) | Anti-pattern (do not) |
|---|---|
| Profile on slowest target, fix largest cycle bucket first | Optimize the function that "looks slow" in review |
| SoA hot data, flat arrays, 64-byte-aware structs | `std::map` of heap-allocated polymorphic entities in the inner loop |
| Arena per frame with one reset point, zero steady-state malloc | `new`/`delete` per entity per frame "because RAII handles it" |
| Cache-line padding on cross-thread state, verified with `perf c2c` | Adjacent atomics hammered by 8 threads on one line |
| CI benchmark with p99 alert at 105% of budget | One heroic optimization pass, never measured again |

## 3. Minimal Example: SoA Layout With Padding (C++)

```cpp
// Before: 96-byte AoS stride, two cache lines per entity, false-shared counter.
struct Entity { float x, y, z, vx, vy, vz; int hp; std::atomic<int> touched; };

// After: hot fields stream linearly; cross-thread counter isolated per line.
struct alignas(64) Positions { float* x; float* y; float* z; };   // 12 B per entity streamed
struct alignas(64) Velocities { float* vx; float* vy; float* vz; };
struct alignas(64) ShardCounter { std::atomic<int> value; };      // one line per thread shard

void integrate(Positions& p, const Velocities& v, int n, float dt) {
    // Stride-1, aliasing-free, auto-vectorized (check -Rpass=loop-vectorize).
    for (int i = 0; i < n; ++i) {
        p.x[i] += v.vx[i] * dt;
        p.y[i] += v.vy[i] * dt;
        p.z[i] += v.vz[i] * dt;
    }
}
```

Workflow: `perf record -g -F 997 ./game --bench dock_scene`, `perf report` to find `integrate` at 11% cycles, convert layout, re-capture, confirm L1-D miss drops and p99 frame falls from 21.8 ms to 15.9 ms. Archive both captures with the change.
