# Rust Production Verification Checklist

## 1. Safety & Unsafe Quarantine
- [ ] Root crate declares `#![forbid(unsafe_code)]` unless an explicit escape hatch is required.
- [ ] Every `unsafe` block or function has an immediate preceding `// SAFETY:` explanation.
- [ ] All raw pointer dereferences are guarded by null, alignment, and bounds checks.
- [ ] Uninitialized memory uses `std::mem::MaybeUninit<T>`, never deprecated `mem::uninitialized()`.
- [ ] Unsafe code is verified with Miri under Stacked Borrows or Tree Borrows (`cargo miri test`).

## 2. Error Handling & Invariants
- [ ] Zero unhandled `.unwrap()` or `.expect()` calls in non-test source code.
- [ ] Public library errors use domain-specific enums derived with `thiserror`.
- [ ] Binary application commands use `anyhow::Result` with `.context()` annotations.
- [ ] Invariants are enforced via newtypes and fallible constructors (`TryFrom`).

## 3. Concurrency & Thread Safety
- [ ] All shared state across threads is wrapped in safe concurrency types (`Arc<Mutex<T>>`, atomics, or channels).
- [ ] Lock acquisition order is consistent across all call paths to prevent deadlocks.
- [ ] Async futures in `tokio::select!` branches are verified to be cancellation-safe.
- [ ] Manual `unsafe impl Send` or `Sync` implementations are accompanied by mathematical safety proofs.

## 4. Linting & Cargo Tooling
- [ ] Passes `cargo clippy --all-targets --all-features -- -D warnings`.
- [ ] `clippy::unwrap_used` and `clippy::expect_used` are denied or treated as errors in production crates.
- [ ] Code is formatted strictly with `cargo fmt --all -- --check`.
- [ ] Dependencies have zero critical security vulnerabilities (`cargo audit`).
