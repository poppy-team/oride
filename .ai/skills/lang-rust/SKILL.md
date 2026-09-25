---
name: lang-rust
description: Modern Rust (2021/2024 editions) engineering, safe-by-default with #![forbid(unsafe_code)], zero-unwrap production paths, Miri-verified unsafe quarantine with Stacked/Tree Borrows, Send/Sync thread invariants, and cargo clippy -D warnings.
---

# Rust Safe-by-Default & Systems Reliability Contract

## 1. Title and Description
**Rust Safe-by-Default & Systems Reliability Contract (`lang-rust`)**
Defines production-grade Rust engineering standards across application, library, embedded/`no_std`, and FFI profiles. Guarantees memory safety, data-race freedom, disciplined unsafe quarantine with Miri verification, idiomatic error handling, and zero-defect linting under Clippy.

## 2. Purpose
Enforce non-negotiable memory safety and type-level correctness in Rust projects within Prumo, ensuring code never relies on undefined behavior (UB), unchecked unwraps, or leaky concurrency abstractions.

## 3. Prerequisites
- Rust stable toolchain installed (`rustc`, `cargo`, `clippy`, `rustfmt`).
- Recommended: `cargo-audit`, `cargo-deny`, and `miri` (for unsafe validation: `rustup component add miri`).
- Standard `Cargo.toml` workspace or package manifest.

## 4. Inputs
- Rust source code (`src/**/*.rs`), tests, and build configurations (`Cargo.toml`, `Cargo.lock`).
- Project profile: `application`, `library`, `systems-engine`, or `embedded-no_std`.

## 5. Outputs
- Fully formatted (`cargo fmt`) and clippy-clean (`cargo clippy -- -D warnings`) Rust code.
- Exhaustive error propagation using type-driven error hierarchies (`thiserror` / `anyhow`).
- Formally justified unsafe blocks with `// SAFETY:` invariants and Miri verification evidence.
- Verification report conforming to `templates/rust-verification-report.md`.

## 6. Execution Steps

### Step 1: Establish Project Profile & Safety Baseline
1. **Default to `#![forbid(unsafe_code)]`**:
   - Every binary application and safe library crate must declare `#![forbid(unsafe_code)]` at the root of `lib.rs` or `main.rs`.
   - If low-level operations (e.g. FFI, SIMD, lock-free memory, custom allocators) strictly require `unsafe`, change root to `#![deny(unsafe_code)]` and isolate unsafe blocks into quarantined leaf modules.
2. **Edition & MSRV**: Specify the edition (2021 or 2024) and `rust-version` (MSRV) explicitly in `Cargo.toml`.

### Step 2: Idiomatic Error Handling
1. **Zero `.unwrap()` or `.expect()` in non-test code**:
   - Never call `.unwrap()` or `.expect()` on paths that could execute in production.
   - Use the question mark operator `?` to bubble errors up the call stack.
2. **Library vs Application Error Patterns**:
   - **Libraries**: Define domain error enums using `thiserror`. Never expose erased error types (e.g. `Box<dyn Error>`) in public APIs.
   - **Applications**: Use `anyhow::Result<T>` or `color_eyre` at the binary boundary to attach user-facing operational context (`.context("reading config file")`).
3. **Parse, Don't Validate**: Use the type system to enforce invariants at construction time (e.g., `NonZeroUsize`, validated newtypes with private fields and fallible constructors).

### Step 3: Unsafe Code Guidelines (UCG) & Quarantine
When an escape hatch into `unsafe` is unavoidable:
1. **Document `// SAFETY:` Invariant**: Every `unsafe` block or `unsafe fn` must be preceded immediately by a `// SAFETY:` comment proving:
   - Why pointers are valid, aligned, non-null, and point to initialized memory.
   - Why lifetime bounds are preserved without aliasing violations.
   - How no data races or unaligned reads can occur.
2. **Enforce Pointer Provenance**:
   - Never cast integer to pointer directly; use `std::ptr::NonNull` for non-null pointers.
   - Use `MaybeUninit<T>` for uninitialized buffers. Never use `mem::uninitialized()`.
   - Respect strict aliasing: never create `&mut T` and `&T` simultaneously to the same memory location.
3. **Miri Verification**: Run `cargo miri test` on all crates containing unsafe code under both Stacked Borrows and Tree Borrows (`-Zmiri-tree-borrows`).

### Step 4: Concurrency & Thread-Safety Invariants
1. **Send and Sync Markers**:
   - Types automatically implement `Send` and `Sync` if all constituent fields do.
   - If manually implementing `unsafe impl Send` or `Sync`, prove that access through `&T` across thread boundaries cannot introduce data races.
2. **Locking Hierarchy & Contention**:
   - Minimize lock hold times: never perform I/O, channel sends, or heavy computation while holding a `MutexGuard`.
   - Prefer lock-free primitives from `std::sync::atomic` with explicit `Ordering` (`Relaxed`, `Acquire`, `Release`, `SeqCst`) rather than defaulting lazily to `SeqCst`.
   - For channel communications, prefer `crossbeam-channel` or `tokio::sync::mpsc`.

### Step 5: Async Rust & Cancellation Safety
1. **Cancellation Invariant in `tokio::select!`**:
   - Ensure branches inside `tokio::select!` are cancellation-safe.
   - If an async future holds partial state across `.await` points and is dropped when another branch completes, it must not leave corrupted shared state or dropped resource locks.
2. **Pin Projections**: Use `pin-project-lite` when implementing custom async streams or futures that require self-referential pinned pointers.

### Step 6: Clippy, Deny & Security Linting
1. Run `cargo clippy --all-targets --all-features -- -D warnings`.
2. Configure `clippy::pedantic` and selectively deny:
   - `clippy::unwrap_used`
   - `clippy::expect_used`
   - `clippy::panic`
   - `clippy::wildcard_imports`
3. Verify dependencies against known vulnerability advisories using `cargo audit`.

## 7. Verification
Run the verification sequence:
```bash
# 1. Format verification
cargo fmt --all -- --check

# 2. Strict linter gate
cargo clippy --all-targets --all-features -- -D warnings -D clippy::unwrap_used

# 3. Unit and integration test suite
cargo test --all-features

# 4. Unsafe memory interpretation (if unsafe blocks present)
cargo miri test

# 5. Dependency security audit
cargo audit
```

## 8. Fallbacks & Failure Recovery
| Failure Class | Root Cause | Deterministic Remediation |
|---|---|---|
| Borrow Checker E0502/E0499 | Simultaneous mutable and immutable borrow in scope | Reorder statements to shorten borrow lifetime. Use scoped blocks `{ ... }`, or clone primitive small fields before borrowing mutable parent. |
| Lifetime Invariant E0106/E0495 | Lifetime of returned reference unbound from input struct | Tie lifetimes explicitly (`fn get<'a>(&'a self) -> &'a Data`), or return owned type (`String`, `Arc<T>`) if caller requires detached lifecycle. |
| Miri Aliasing Violation | Multiple raw pointers dereferenced into mutable reference alias | Switch raw pointer manipulation to `std::cell::UnsafeCell` or inspect under Tree Borrows (`-Zmiri-tree-borrows`). |
| Async Deadlock | Mutex held across `.await` point blocking async runtime executor | Replace `std::sync::Mutex` with `tokio::sync::Mutex` only if held across `.await`, or restructure code to drop guard before awaiting. |

## 9. Constraints
- **Forbid unhandled panics**: No `.unwrap()` in production code. Use `?` or explicit match/if-let.
- **Mandatory `// SAFETY:` proofs**: Unsafe code lacking an invariant proof will be rejected by CI.
- **Zero Clippy warnings**: Code with compiler or clippy warnings fails the conformance gate immediately.
- **No speculative dependencies**: Only add dependencies approved in the project manifest.

## 10. Examples

### Anti-Pattern: Unchecked Unwrap and Flawed Unsafe
```rust
// BAD: Panics on bad input, unaligned raw pointer dereference without safety invariants
pub fn parse_and_read(raw: &str, ptr: *const u32) -> u32 {
    let _id: u32 = raw.parse().unwrap(); // PANIC RISK!
    unsafe {
        *ptr // UB RISK: ptr might be null, unaligned, or dangling!
    }
}
```

### Idiomatic Pattern: Type-Driven Safety & Sound Unsafe Boundary
```rust
use std::num::ParseIntError;
use std::ptr::NonNull;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum ProcessingError {
    #[error("invalid numeric payload: {0}")]
    Parse(#[from] ParseIntError),
    #[error("null buffer pointer encountered")]
    NullBuffer,
    #[error("index {index} exceeds buffer bounds {capacity}")]
    OutOfBounds { index: usize, capacity: usize },
}

/// A sound, memory-safe wrapper around a contiguous allocated buffer.
pub struct SafeBuffer<T> {
    ptr: NonNull<T>,
    capacity: usize,
}

impl<T> SafeBuffer<T> {
    /// Constructs a safe buffer from a verified non-null pointer and capacity.
    ///
    /// # Safety
    /// Caller must guarantee `ptr` points to `capacity` continuously allocated elements of `T`.
    pub unsafe fn from_raw_parts(ptr: *mut T, capacity: usize) -> Result<Self, ProcessingError> {
        let non_null = NonNull::new(ptr).ok_or(ProcessingError::NullBuffer)?;
        Ok(Self { ptr: non_null, capacity })
    }

    pub fn get(&self, index: usize) -> Result<&T, ProcessingError> {
        if index >= self.capacity {
            return Err(ProcessingError::OutOfBounds { index, capacity: self.capacity });
        }
        // SAFETY: index is bounds-checked above, ptr is verified non-null,
        // and caller lifetime is bound to &'self.
        unsafe {
            Ok(&*self.ptr.as_ptr().add(index))
        }
    }
}

// Send + Sync markers proven safe because T is Send + Sync and access is immutable
unsafe impl<T: Send> Send for SafeBuffer<T> {}
unsafe impl<T: Sync> Sync for SafeBuffer<T> {}
```
