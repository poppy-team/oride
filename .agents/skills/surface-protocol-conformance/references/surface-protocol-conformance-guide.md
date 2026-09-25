# Surface & Protocol Conformance — Technical Reference Guide

## 1. Core Concepts

**Single domain source**: business rules live once in application/domain services. CLI, TUI, GUI, and daemon are thin surfaces over the public Prumo Protocol — they render, collect input, and forward; they never duplicate validation, budgeting, or permission logic locally.

**Non-bypassable governance**: every surface routes through the Permission Engine, Budget Manager, and Gauntlet Quality Gates. A surface that can mutate state without a gate verdict is a conformance failure, not a feature.

**Protocol version negotiation**: clients handshake a versioned protocol, tolerate skew within the supported window, and replay missed events after reconnect without state corruption (idempotent handlers + monotonic sequence numbers).

## 2. Patterns

- **Conformance matrix**: rows = surfaces (CLI/TUI/GUI/daemon), columns = operations; each cell cites the shared domain entry point and the gate verdict path. Empty cells are drift suspects.
- **Golden-trace replay**: record a canonical protocol session once; replay it against every surface and diff resulting domain states byte-for-byte.
- **Error-taxonomy parity**: the same domain error (e.g. `BUDGET_EXCEEDED`, `PERMISSION_DENIED`) surfaces with identical codes and semantics everywhere; assert with a shared fixture table.
- **Skew-window test**: run client at protocol vN against daemon at vN+1 (and reverse) inside the supported window; assert negotiation succeeds and degraded-but-safe behavior outside it.

## 3. Anti-Patterns

- Client-side "pre-validation" that diverges from domain rules (two sources of truth).
- Daemon-only shortcuts that skip the Permission Engine for "internal" calls.
- Silent protocol upgrades with no negotiation or replay support.
- Surface-specific error strings that hide the canonical domain code.

## 4. Worked Example

Operation `budget.commit`: CLI, TUI, and daemon all invoke `domain/budget.commit` and surface `BUDGET_EXCEEDED` identically. Golden-trace replay across three surfaces yields identical ledger state hashes (`b7e2…`). GUI initially bypassed the Budget Manager with a local estimate — flagged as drift, fixed by routing through the protocol call; matrix cell updated with the shared entry point and gate verdict log.

## 5. Verification Pointers

- Every matrix cell cites a shared domain entry point (no "local logic" cells).
- Golden-trace replay diffs are empty across all surfaces.
- Version-skew test passes inside the window and fails safe outside it.
