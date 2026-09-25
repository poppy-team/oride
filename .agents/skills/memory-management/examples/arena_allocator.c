// Production-grade Arena (Bump) Memory Allocator in C99
// Demonstrates strict power-of-two alignment calculation, bounds protection,
// bulk deallocation, and AddressSanitizer manual poisoning hooks.

#include <stdio.h>
#include <stdlib.h>
#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>
#include <assert.h>
#include <string.h>

#if defined(__SANITIZE_ADDRESS__) || (defined(__has_feature) && __has_feature(address_sanitizer))
#include <sanitizer/asan_interface.h>
#define ASAN_POISON(addr, size)   __asan_poison_memory_region((addr), (size))
#define ASAN_UNPOISON(addr, size) __asan_unpoison_memory_region((addr), (size))
#else
#define ASAN_POISON(addr, size)   ((void)0)
#define ASAN_UNPOISON(addr, size) ((void)0)
#endif

typedef struct {
    uint8_t* buffer;
    size_t capacity;
    size_t offset;
    size_t high_water_mark;
} MemoryArena;

/// Forward aligns a raw address to a required power-of-two alignment boundary.
static inline uintptr_t align_forward(uintptr_t ptr, size_t align) {
    assert((align & (align - 1)) == 0 && "Alignment must be a power of two");
    uintptr_t a = (uintptr_t)align;
    uintptr_t modulo = ptr & (a - 1);
    if (modulo != 0) {
        ptr += (a - modulo);
    }
    return ptr;
}

/// Initializes an arena with a pre-allocated memory buffer.
MemoryArena arena_create(size_t capacity) {
    MemoryArena arena;
    arena.buffer = (uint8_t*)malloc(capacity);
    assert(arena.buffer != NULL && "Arena backing store allocation failed");
    arena.capacity = capacity;
    arena.offset = 0;
    arena.high_water_mark = 0;

    // Poison entire arena initially so unallocated memory cannot be read/written
    ASAN_POISON(arena.buffer, capacity);
    return arena;
}

/// Allocates an aligned chunk of memory from the arena.
void* arena_alloc_align(MemoryArena* arena, size_t size, size_t alignment) {
    if (size == 0) return NULL;

    uintptr_t current_addr = (uintptr_t)arena->buffer + (uintptr_t)arena->offset;
    uintptr_t aligned_addr = align_forward(current_addr, alignment);
    size_t padding = aligned_addr - current_addr;
    size_t total_needed = padding + size;

    if (arena->offset + total_needed > arena->capacity) {
        // Out of memory in this arena
        return NULL;
    }

    arena->offset += total_needed;
    if (arena->offset > arena->high_water_mark) {
        arena->high_water_mark = arena->offset;
    }

    void* allocated_ptr = (void*)aligned_addr;
    // Unpoison exactly the chunk handed to user code
    ASAN_UNPOISON(allocated_ptr, size);
    return allocated_ptr;
}

/// Allocates memory with default max_align_t alignment (typically 8 or 16 bytes).
void* arena_alloc(MemoryArena* arena, size_t size) {
    return arena_alloc_align(arena, size, sizeof(max_align_t));
}

/// Resets the arena, freeing all contained allocations in O(1) time.
void arena_reset(MemoryArena* arena) {
    // Re-poison all previously allocated regions
    ASAN_POISON(arena->buffer, arena->offset);
    arena->offset = 0;
}

/// Completely destroys the arena backing store.
void arena_destroy(MemoryArena* arena) {
    if (arena->buffer) {
        ASAN_UNPOISON(arena->buffer, arena->capacity);
        free(arena->buffer);
        arena->buffer = NULL;
        arena->capacity = 0;
        arena->offset = 0;
    }
}

// Verification test
int main(void) {
    printf("=== Initializing 1MB Memory Arena ===\n");
    MemoryArena arena = arena_create(1024 * 1024);

    // 1. Allocate 64-byte aligned SIMD vector
    float* simd_vec = (float*)arena_alloc_align(&arena, 64 * sizeof(float), 64);
    assert(simd_vec != NULL);
    assert(((uintptr_t)simd_vec % 64) == 0 && "Vector must be 64-byte aligned");

    for (int i = 0; i < 64; i++) {
        simd_vec[i] = (float)i * 1.5f;
    }

    // 2. Allocate an array of strings
    char* msg = (char*)arena_alloc(&arena, 128);
    assert(msg != NULL);
    snprintf(msg, 128, "Arena allocation successful at offset %zu", arena.offset);

    printf("Report: %s\n", msg);
    printf("Peak Memory Consumption: %zu bytes (Capacity: %zu bytes)\n",
           arena.high_water_mark, arena.capacity);

    // 3. O(1) bulk reset
    arena_reset(&arena);
    printf("Arena reset complete. Current offset: %zu\n", arena.offset);

    arena_destroy(&arena);
    printf("Arena test finished cleanly.\n");
    return 0;
}
