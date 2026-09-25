# Testing Quality Delivery Record

## Metadata
- Skill: Testing Quality
- Date: 2026-09-23
- Author: quality-agent
- Goal: TEST-31 — protect checkout idempotency

## Pyramid
- Unit: 126 tests for amount parsing, policy decisions, and error mapping.
- Integration: 18 tests against PostgreSQL and the payment adapter.
- E2E: 3 critical checkout journeys on Chromium and WebKit.

## Regression Case
- Test: `test_duplicate_checkout_is_idempotent`
- Seed: `checkout-2026-09-23`
- Baseline: failed before the idempotency-key fix
- Current: passes consistently across 10 isolated runs

## Quality Evidence
- No test reaches the public network; the adapter is replaced by a deterministic fake.
- Business-logic line coverage is 91%; the threshold is 85%.
- Time and random sources are injected, so failures are reproducible.

## Follow-Up
- Add a mutation check for the retry branch in the next quality slice.
