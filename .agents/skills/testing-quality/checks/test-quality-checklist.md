# Testing Quality — Verification Checklist

## 1. Pyramid Shape & Placement
- [ ] Suite follows the pyramid: dominant fast unit base, thinner integration middle, minimal E2E peak.
- [ ] Tests pushed down the pyramid where possible; E2E reserved for critical user journeys.
- [ ] New code paths carry tests at the lowest feasible layer.

## 2. Determinism & Isolation
- [ ] Suite passes 10 consecutive runs with zero flakes; no network, wall-clock, RNG, or shared-state nondeterminism.
- [ ] Execution order does not alter outcomes; fixtures reset (DB, caches, objects) per test via setup/teardown.
- [ ] Time controlled with fakes/freezing; async behavior uses deterministic synchronization, not sleeps.

## 3. Behavior Assertions & Fixtures
- [ ] Assertions target observable behavior through public APIs, not private internals.
- [ ] Fixtures representative and minimal; golden fixtures versioned for regression comparison.
- [ ] Failure output carries diagnostic context (inputs, expected vs actual, seed) for sub-minute triage.

## 4. Regression Discipline
- [ ] Every fixed bug lands with a regression test that fails before and passes after the fix.
- [ ] Coverage gates hold (business-logic target ≥85%); drops block merge.
- [ ] Flaky tests quarantined with tracked issues instead of retried into greenness.

## 5. Evidence & Sign-Off
- [ ] Coverage report and full pass log recorded with command lines.
- [ ] `scripts/verify.sh` exits 0 from the repository root.
