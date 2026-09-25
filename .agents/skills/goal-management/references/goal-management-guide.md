# Prumo Goal Authoring & Lock Discipline Reference Guide

## 1. Core Concepts & Models

### 1.1 The Criterion Predicate Model
A criterion is a boolean function over evidence: `satisfied(c) = exists(e in
evidence(c)) and passes(e)`. Anything that cannot be evaluated this way —
`code is clean`, `performance is good` — is not a criterion but a wish. Rewrite
wishes into predicates: `gocyclo reports max M=6 on internal/billing`,
`k6 p99 104 ms against a 120 ms budget`.

### 1.2 Lock Semantics: Append-Only Criteria
Locking divides the Goal file into two regions: immutable criteria lines and
mutable status/evidence lines. Formally, for lock commit $L$:

$$\text{allowed diff}(L, HEAD) \subseteq \{\text{status}, \text{evidence}, \text{notes}\}$$

Any diff touching criteria lines without a linked re-scope entry is a lock
violation, detected mechanically with `git diff L..HEAD -- <goal-file>`. The
re-scope entry is the only legal mutation path: reason, authorizer, impact.

### 1.3 Phase Gating Arithmetic
A Goal with $n$ phases and per-phase success probability $p$ completes with
probability $p^n$ absent gates. Gates do not raise $p$; they convert late
catastrophic failure into early cheap failure by checking the exit predicate
before the next phase spends budget. Size phases so each gate costs under 10
percent of the phase: a 2-day phase earns a 2-hour gate, not a 2-day review.

### 1.4 Staleness Detection
$$\text{stale}(c) = \text{now} - \text{last\_evidence\_update}(c) > 7\text{ days and status}(c) \ne \text{satisfied}$$

The weekly health check evaluates this predicate per open criterion. Stale items
get exactly one of three verdicts: reaffirmed (still valid, next action dated),
re-scoped (criterion or owner changes), archived (no longer wanted, with reason).
A fourth state — ignored — is the failure mode this skill exists to prevent.

## 2. Patterns and Anti-Patterns

**Do: name the evidence in the criterion.**
`AC2 passes when TestRoundTieredTax 12/12 green including 0.125→0.13` tells the
implementer exactly what done looks like and the reviewer exactly what to check.

**Do: record non-goals at authoring time.**
`Multi-currency excluded; PDF layout untouched` settles week-3 scope arguments in
week 0, when they cost a sentence instead of a re-plan.

**Do not: lock and keep editing.** Locking on Monday and rewriting criteria on
Wednesday without a re-scope entry voids the lock's meaning. The entry is cheap;
write it.

**Do not: phase by layer without gates.**
`Backend week, frontend week, testing week` with no exit predicates is a schedule,
not a plan. Each phase needs its falsifiable exit or progress is unmeasurable.

**Do not: mark satisfied on intention.**
`Will verify after merge` is not evidence. The criterion stays open until the
named log exists and passes; optimism is not a test result.

## 3. Goal File Example

```markdown
# Goal: billing-tiered-pricing — Tiered pricing computes within 0.01 of fixtures

## Non-goals
- Multi-currency support excluded; PDF invoice layout untouched.

## Acceptance criteria (LOCKED 2026-09-09 by Marina Duarte at commit 1a2b3c4d)
- AC1: TestCalculateTieredPrice 18/18 green against golden fixtures (tolerance 0.01).
- AC2: TestRoundTieredTax 12/12 green including 0.125→0.13 and 2.675→2.68.
- AC3: k6 p99 at most 120 ms at 500 rps on load/invoice-endpoint.js.
- AC4: go-apidiff v2.13.0 reports zero breaking changes.

## Phases
1. Data model (Marina, exits: TieredPrice struct validated, 6 unit tests green).
2. Computation (Marina, exits: AC1 + AC2 green).
3. API exposure (Rafael, exits: AC4 green, needs phase 1 struct).
4. Rollout (Marina, exits: AC3 green in staging, needs release plan).

## Evidence ledger
- AC1: PR billing-api#482, CI run 4821 — SATISFIED 2026-09-15.
- AC2: PR billing-api#482, CI run 4821 — SATISFIED 2026-09-15.
- AC3: k6 run 2026-09-15, p99 104 ms — SATISFIED 2026-09-15.
- AC4: apidiff report 2026-09-15 — SATISFIED 2026-09-15.
```
