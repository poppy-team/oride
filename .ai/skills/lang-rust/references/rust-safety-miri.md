# Unsafe Rust, Pointer Provenance & Miri Verification Guide

## 1. The Role of Miri in Systems Engineering
Miri is an interpreter for Rust's Mid-level Intermediate Representation (MIR). It tracks pointer provenance, memory allocations, and borrow permissions at runtime to detect Undefined Behavior (UB) that the static compiler cannot catch, including:
- Out-of-bounds memory accesses and buffer overflows.
- Use-after-free (accessing memory after `dealloc` or after an owned value was dropped).
- Violations of the aliasing rules (e.g., mutating memory while an immutable reference `&T` exists).
- Reading uninitialized memory.
- Creating invalid values (e.g., a `bool` that is neither 0 nor 1, or a null reference `&T`).
- Unaligned pointer dereferences.
- Data races in multithreaded execution.

## 2. Aliasing Models: Stacked Borrows vs Tree Borrows

### Stacked Borrows (SB)
Stacked Borrows assigns a unique tag to every pointer and reference derived from an allocation, maintaining a stack of permissions (`SharedReadOnly`, `Unique`, `Disabled`) per memory location.
- **Rule**: When a pointer is used, its tag must be found in the stack. Any items above it on the stack that conflict with the operation are popped and permanently disabled.
- **Limitation**: Can be overly strict when using two raw pointers derived from the same base pointer to write into non-overlapping sub-slices.

### Tree Borrows (TB)
Tree Borrows replaces the linear stack with a tree of permissions:
- Retains pointer parent-child relationships in a tree structure.
- Permissive for complex data structures (e.g. self-referential collections, raw-pointer intrusive linked lists, ring buffers).
- Invoke under Miri using: `cargo miri test -- -Zmiri-tree-borrows`.

## 3. The `MaybeUninit<T>` Paradigm
Never initialize memory with `mem::uninitialized()`, which is unsound even for primitive integers if padding or uninitialized bytes are read.
```rust
use std::mem::MaybeUninit;

// CORRECT:
let mut buffer: [MaybeUninit<u8>; 1024] = unsafe {
    MaybeUninit::uninit().assume_init()
};

// Fill the buffer safely:
for (i, elem) in buffer.iter_mut().enumerate() {
    elem.write(i as u8);
}

// Convert slice to initialized:
let initialized: &[u8] = unsafe {
    // SAFETY: Entire 1024 elements were populated in the loop above.
    std::slice::from_raw_parts(buffer.as_ptr() as *const u8, 1024)
};
```

## 4. CI Integration for Unsafe Crates
Include the following step in GitHub Actions or local verification pipelines:
```yaml
- name: Run Miri
  run: |
    rustup component add miri
    cargo miri test --all-features
    cargo miri test --all-features -- -Zmiri-tree-borrows
```
