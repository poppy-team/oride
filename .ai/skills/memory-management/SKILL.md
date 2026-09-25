---
name: memory-management
description: Native systems memory management, arena & pool allocators, cache locality, false sharing prevention, ASan/Valgrind verification, and leak-free resource lifecycles.
---

# Native Memory Management & Allocation Architecture

## 1. Title and Description
**Native Memory Management & Allocation Architecture (`memory-management`)**
Governs the design, implementation, and verification of low-level memory architectures across C, C++, Rust, Zig, and systems runtimes. Covers custom allocation strategies (Arenas, Fixed-size Pools, Slabs), cache-conscious layout optimization (SoA vs AoS, false sharing prevention), spatial/temporal safety, and AddressSanitizer (ASan) validation.

## 2. Purpose
Eliminate memory fragmentation, latency spikes from general-purpose `malloc`/`free` contention, use-after-free (UAF) vulnerabilities, and cache-thrashing bottlenecks in performance-critical systems.

## 3. Prerequisites
- Target language toolchain supporting low-level memory addressing (C, C++, Rust, Zig, Odin).
- Sanitizer tooling: AddressSanitizer (`-fsanitize=address`) or Valgrind Memcheck.

## 4. Inputs
- Subsystem architecture specifications and memory footprint targets.
- Allocation frequency requirements: per-frame, per-request, or persistent.
- Hardware constraints: CPU cache hierarchy (L1/L2/L3), cache line size (typically 64 bytes), and page size (typically 4KB).

## 5. Outputs
- Custom allocator implementations (Arena, Pool, Ring Buffer) with alignment proofs.
- Cache-optimized data layouts (Structure of Arrays, cache-aligned hot structs).
- Memory audit reports conforming to `templates/memory-management-template.md`.
- Sanitizer-clean verification logs with zero detected leaks or invalid accesses.

## 6. Execution Steps

### Step 1: Select the Optimal Allocation Strategy
Match the problem's lifecycle to the mathematically optimal memory allocator:

| Strategy | Allocation Complexity | Deallocation Complexity | Ideal Workload |
|---|---|---|---|
| **Arena / Bump Allocator** | O(1) (pointer increment) | O(1) (bulk reset) | Per-frame game ticks, per-request HTTP parsing, compiler AST passes |
| **Fixed-Size Pool Allocator** | O(1) (freelist pop) | O(1) (freelist push) | Homogeneous objects (graph nodes, ECS components, network packets) |
| **Stack / Scratch Allocator** | O(1) | O(1) (LIFO rewind) | Scoped temporary buffers, intermediate string transformations |
| **Slab / Buddy Allocator** | O(log N) | O(log N) | Dynamic objects of variable power-of-two sizes with coalescing |

### Step 2: Enforce Strict Memory Alignment
Hardware requires values to reside at memory addresses that are multiples of their size:
1. **Alignment Calculation**:
   ```c
   uintptr_t align_forward(uintptr_t ptr, size_t align) {
       uintptr_t p = ptr;
       uintptr_t a = (uintptr_t)align;
       uintptr_t modulo = p & (a - 1); // Only valid when align is power-of-two
       if (modulo != 0) {
           p += (a - modulo);
       }
       return p;
   }
   ```
2. **SIMD & Page Alignment**:
   - Ensure AVX-256 buffers are aligned to 32 bytes; AVX-512 to 64 bytes.
   - Ensure direct I/O and shared memory buffers align with OS virtual memory page boundaries (4096 bytes).

### Step 3: Cache Locality & False Sharing Prevention
1. **False Sharing**:
   - Occurs when two independent threads modify variables residing on the same 64-byte cache line, causing cache-coherence invalidate traffic across CPU cores.
   - **Remediation**: Align per-thread structures to 64-byte boundaries using `alignas(64)` (C++), `#[repr(align(64))]` (Rust), or explicit padding fields.
2. **Structure of Arrays (SoA) vs Array of Structures (AoS)**:
   - Prefer AoS (`struct Particle { float x, y, z, mass, color; }`) for random individual access.
   - Prefer SoA (`struct ParticleCloud { float* x; float* y; float* z; }`) for tight vectorized update loops that scan single fields sequentially.

### Step 4: AddressSanitizer (ASan) Poisoning Integration
Custom memory arenas bypass libc `malloc`/`free`, which blinds default ASan checks. Manually instrument custom allocators using the ASan C API:
```c
#if defined(__SANITIZE_ADDRESS__) || (defined(__has_feature) && __has_feature(address_sanitizer))
#include <sanitizer/asan_interface.h>
#define ASAN_POISON_MEMORY_REGION(addr, size)   __asan_poison_memory_region((addr), (size))
#define ASAN_UNPOISON_MEMORY_REGION(addr, size) __asan_unpoison_memory_region((addr), (size))
#else
#define ASAN_POISON_MEMORY_REGION(addr, size)   ((void)0)
#define ASAN_UNPOISON_MEMORY_REGION(addr, size) ((void)0)
#endif
```
- Poison unallocated arena capacity: `ASAN_POISON_MEMORY_REGION(arena->curr, arena->capacity - arena->offset)`.
- Unpoison allocated blocks on bump: `ASAN_UNPOISON_MEMORY_REGION(allocated_ptr, requested_size)`.
- Re-poison everything upon `arena_reset()`.

### Step 5: Virtual Memory & Operating System Guard Pages
For large allocations (> 1MB):
1. Use `mmap(MAP_ANONYMOUS | MAP_PRIVATE)` (POSIX) or `VirtualAlloc` (Windows) to allocate virtual address space lazily.
2. Place a `PROT_NONE` guard page at the end of the allocation block to immediately trap stack/buffer overflows with a hardware segment fault (`SIGSEGV`) rather than silent memory corruption.

### Step 6: Verification & Stress Testing
1. Compile with `-fsanitize=address,undefined -g`.
2. Run differential tests with Valgrind: `valgrind --tool=memcheck --leak-check=full --track-origins=yes ./binary`.
3. Check high-water memory mark and fragmentation metrics.

## 7. Verification
Run the memory audit routine:
```bash
# 1. AddressSanitizer build and execution
clang -fsanitize=address,undefined -g -O1 src/allocator_test.c -o /tmp/asan_test
/tmp/asan_test

# 2. Check for naked malloc/free in hot code
grep -rn "malloc(" src/ | grep -v "allocator"
```

## 8. Fallbacks & Failure Recovery
| Failure Class | Root Cause | Deterministic Remediation |
|---|---|---|
| Use-After-Free (UAF) | Pointer accessed after arena reset or block free | Poison freed regions using ASan macros. Clear pointers to `NULL` immediately upon release. |
| Memory Fragmentation | Interleaved allocation/deallocation of heterogeneous sizes in general heap | Segregate objects by size class into dedicated pool allocators or arenas. |
| Cache Thrashing / Stalls | Heavy cache misses due to fragmented pointer chasing | Flatten linked data structures into contiguous arrays indexed by 32-bit handles instead of raw 64-bit pointers. |
| False Sharing CPU Bottleneck | Cores modifying adjacent memory in parallel loop | Pad data to 64 bytes (`alignas(64)`) or allocate thread-local scratch buffers. |

## 9. Constraints
- **Zero dynamic heap allocations in hot loops**: No `malloc`, `new`, or heap reallocation in per-frame or high-frequency real-time loops.
- **Explicit alignment proofs**: All custom allocations must be mathematically guaranteed to meet type alignment requirements.
- **Clean ASan passes**: Zero memory leaks, out-of-bounds, or misaligned accesses tolerated under `-fsanitize=address`.

## 10. Examples

### Anti-Pattern: Unaligned Allocator & Blind Use-After-Free
```c
// BAD: Ignores alignment requirements, no ASan poisoning, leaks on individual free
typedef struct {
    char* buffer;
    size_t offset;
    size_t capacity;
} NaiveArena;

void* naive_alloc(NaiveArena* a, size_t size) {
    void* ptr = a->buffer + a->offset; // ERROR: Does not align ptr! UB on ARM / SIMD!
    a->offset += size;                 // ERROR: No capacity check; buffer overflow!
    return ptr;
}
```

### Idiomatic Pattern: Robust Aligned Arena with ASan Instrumentation
```c
// GOOD: Safe alignment math, bounds enforcement, and sanitizer poison guards
#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

typedef struct {
    uint8_t* buffer;
    size_t offset;
    size_t capacity;
} MemoryArena;

static inline uintptr_t align_forward_ptr(uintptr_t ptr, size_t align) {
    uintptr_t p = ptr;
    uintptr_t a = (uintptr_t)align;
    uintptr_t modulo = p & (a - 1);
    if (modulo != 0) {
        p += (a - modulo);
    }
    return p;
}

void* arena_alloc_align(MemoryArena* arena, size_t size, size_t alignment) {
    uintptr_t current_ptr = (uintptr_t)arena->buffer + (uintptr_t)arena->offset;
    uintptr_t aligned_ptr = align_forward_ptr(current_ptr, alignment);
    size_t new_offset = (aligned_ptr - (uintptr_t)arena->buffer) + size;

    if (new_offset > arena->capacity) {
        return NULL; // Out of memory
    }

    arena->offset = new_offset;
    return (void*)aligned_ptr;
}

void arena_reset(MemoryArena* arena) {
    arena->offset = 0;
}
```
