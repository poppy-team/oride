# Native Memory Management Verification Checklist

## 1. Allocation Strategy & Bounds Safety
- [ ] Hot-path allocation strategy is appropriate (Arena for batch/frame; Pool for uniform objects; general heap avoided).
- [ ] Custom allocators enforce explicit power-of-two alignment (`align_forward`).
- [ ] SIMD buffers (AVX2/AVX-512) are aligned to 32/64 bytes respectively.
- [ ] Capacity overflows return explicit `NULL`/`Err` or trigger controlled assertions, never memory overruns.
- [ ] Zero dynamic memory allocations (`malloc`, `new`, `std::vector::push_back` beyond capacity) in critical inner loops.

## 2. Temporal & Spatial Memory Safety
- [ ] Pointers into reset arenas are never retained across lifecycle boundaries (Zero Use-After-Free).
- [ ] Deallocated memory is poisoned using AddressSanitizer macros (`__asan_poison_memory_region`).
- [ ] Custom pool allocators use double-free detection on their freelist.
- [ ] Virtual memory mappings over 1MB include `PROT_NONE` guard pages.

## 3. Cache Locality & Concurrency Hygiene
- [ ] Shared data updated by concurrent threads is aligned to 64 bytes (`alignas(64)`) to eliminate false sharing.
- [ ] Hot iteration loops operate on contiguous memory (preferring Structure of Arrays for vectorization).
- [ ] Struct fields are ordered from largest to smallest to eliminate compiler alignment padding bloat.

## 4. Sanitizer & Profiling Verification
- [ ] Test suite executes cleanly under AddressSanitizer (`-fsanitize=address,undefined`).
- [ ] LeakSanitizer reports 0 leaked bytes upon graceful process termination.
- [ ] Memory footprint is bounded and matches declared subsystem memory budget.
