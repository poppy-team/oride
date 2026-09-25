# Native Memory Architecture & Allocator Engineering Guide

## 1. The Cost of General-Purpose Heap Allocators
Standard heap allocators (`ptmalloc` in glibc, Windows Heap) are general-purpose: they handle arbitrary sizes, threads, and lifetimes. To achieve this, they must:
1. Maintain global or per-thread lock structures to prevent race conditions during freelist traversal.
2. Incur bookkeeping metadata headers (8 to 16 bytes per allocation).
3. Search for best-fit or first-fit memory blocks, causing pointer chasing and memory fragmentation.
4. Call into the operating system kernel via `brk`/`mmap` when arena bins are depleted.

In low-latency or high-throughput systems, replacing general heap calls with specialized allocators yields 10x to 100x performance improvements.

## 2. Arena Allocator Architecture (Linear / Bump Allocator)
An Arena pre-allocates a contiguous slab of memory and satisfies allocation requests simply by bumping an offset pointer:
```
[ Allocated Block A | Allocated Block B | Allocated Block C | ... Unallocated ... ]
^ buffer                                                    ^ offset              ^ capacity
```
- **Allocation Cost**: 2 additions and 1 bitwise mask (for alignment) -> ~1 nanosecond.
- **Deallocation Cost**: Reset `offset = 0` -> O(1) instantaneous cleanup of all allocations.
- **Lifetime Discipline**: Objects in an arena must share the same lifetime (e.g. per-frame, per-request, or per-AST compilation pass).

## 3. CPU Cache Hierarchy & False Sharing

Modern x86-64 and ARM64 CPUs fetch data into cache lines of **64 bytes**.

### False Sharing Mechanism:
```
Core 0 writes to Variable A (bytes 0..7)  \  Both sit on the same 64-byte Cache Line!
Core 1 writes to Variable B (bytes 8..15) /
---------------------------------------------------------------------------------
Result: Core 0's write marks the line INVALID in Core 1's L1 cache via MESI protocol.
Core 1 must stall while fetching the line from L3 or main memory!
```

### Remediation in C/C++ and Rust:
```cpp
// C++
struct alignas(64) ThreadWorkerStats {
    uint64_t processed_tasks;
    uint64_t failed_tasks;
    // Remaining bytes are padded to 64 bytes automatically
};
```
```rust
// Rust
#[repr(align(64))]
struct ThreadWorkerStats {
    processed_tasks: u64,
    failed_tasks: u64,
}
```

## 4. AddressSanitizer Manual Poisoning
When developing custom allocators, include the ASan interface:
```c
#include <sanitizer/asan_interface.h>

// When memory is handed out:
__asan_unpoison_memory_region(allocated_ptr, size);

// When memory is reset or freed:
__asan_poison_memory_region(freed_ptr, size);
```
This enables ASan to catch out-of-bounds reads and use-after-free bugs directly inside custom arena slabs.
