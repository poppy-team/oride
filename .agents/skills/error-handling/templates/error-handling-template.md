# Error Handling Report — Payments Charge Path

## 1. Metadata
- **Skill**: error-handling (Error Handling)
- **Date**: 2026-09-09
- **Author / Agent**: payments-team
- **Target Goal / Phase**: GOAL-077 charge-error-taxonomy

## 2. Executive Summary
Applied the error taxonomy to the payments charge path: 14 domain errors typed (`InsufficientFunds`, `CardDeclined`, `IdempotencyConflict`, …), PSP timeouts retried with exponential backoff + jitter (max 4 attempts, idempotency-keyed), and user-facing messages sanitized to 6 safe codes. Fatal states (ledger divergence) trip the fail-fast guard. Error budget: charge-path 5xx ≤ 0.1% over 30 d; current 0.03%.

## 3. Inputs & Scope
- **Inputs Evaluated**: `payments/charge.go` error sites (23), PSP timeout profile (p99 2.1 s), logging pipeline (JSON schema v2)
- **Artifacts Modified**: `payments/errors.go` (new taxonomy), `payments/charge.go` (wrapping), `api/middleware/problem_details.go` (RFC 7807 mapping)

## 4. Key Findings & Implementation Details
- **Taxonomy**: `ChargeError{Code, Kind(domain|transient|fatal), retryable bool, safe_message}`. Transient (`PSPTimeout`, `RateLimited`) → retry; domain (`InsufficientFunds`) → return immediately; fatal (`LedgerDivergence`) → panic-guard + page.
- **Retries**: Backoff 200 ms × 2^n + uniform jitter, max 4 attempts, only with idempotency key present; verified no double-charge in 500 chaos-injected timeout runs.
- **Context**: Low-level PSP errors wrapped with `charge_id`, `merchant_id`, `attempt` at each layer; user messages mapped to 6 localized safe strings (no PSP internals leak; verified by response-scan test).
- **Boundaries**: HTTP middleware converts panics to RFC 7807 500s with correlation IDs; circuit breaker on PSP (opens after 20 failures/30 s, half-open probe 1 req/10 s).
- **Monitoring**: `charge_error_rate` SLI + alert at 0.08%; dashboard panels per error code.

## 5. Verification & Evidence
- **Evidence Type**: test
- **Test Results**: Passed — 500/500 chaos timeout runs single-charge; response scan 0 leaks; breaker trip/recover 5/5 cycles
- **Static Analysis Status**: Pass — exhaustive `switch` over taxonomy enforced by linter; no bare error returns in charge path

## 6. Next Steps & Handoff
- Extend taxonomy to refund path (GOAL-078); owner: payments-team.
- Localize 6 safe messages to pt-BR/es; owner: content-team.
