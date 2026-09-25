# Goal Specification Template

## 1. Goal Identity & Outcome
- **ID**: billing-tiered-pricing; **file**: `goals/billing-tiered-pricing.md`
- **Title**: Tiered pricing computes within 0.01 of golden fixtures at production latency
- **Problem**: flat pricing overcharges high-volume customers; finance approved tiered model 2026-08-30
- **Non-goals**: multi-currency support excluded; PDF invoice layout untouched; admin UI out of scope
- **Sponsor**: Finance lead Paula Nogueira; **activated**: 2026-09-09

## 2. Acceptance Criteria (Locked 2026-09-09 by Marina Duarte at commit 1a2b3c4d)
- **AC1**: `TestCalculateTieredPrice` 18/18 green against golden fixtures, tolerance 0.01
- **AC2**: `TestRoundTieredTax` 12/12 green including cases `0.125→0.13` and `2.675→2.68`
- **AC3**: k6 p99 at most 120 ms at 500 rps on `load/invoice-endpoint.js` in staging
- **AC4**: `go-apidiff v2.13.0` reports zero breaking changes on the public API

## 3. Phases & Exit Gates
- **Phase 1 data model** (Marina Duarte, 2026-09-09–2026-09-10): exits when `TieredPrice` struct validated with 6 unit tests green
- **Phase 2 computation** (Marina Duarte, 2026-09-10–2026-09-12): exits when AC1 and AC2 green
- **Phase 3 API exposure** (Rafael Costa, 2026-09-12–2026-09-14, needs Phase 1 struct): exits when AC4 green
- **Phase 4 rollout** (Marina Duarte, 2026-09-15–2026-09-17, needs release plan): exits when AC3 green in staging

## 4. Evidence Ledger
- **AC1**: PR billing-api#482, CI run 4821 — SATISFIED 2026-09-15
- **AC2**: PR billing-api#482, CI run 4821 — SATISFIED 2026-09-15
- **AC3**: k6 run 2026-09-15, p99 104 ms — SATISFIED 2026-09-15
- **AC4**: apidiff report 2026-09-15, zero breaking — SATISFIED 2026-09-15

## 5. Lock & Re-Scope Record
- **Lock**: `prumo goal lock billing-tiered-pricing --criteria AC1,AC2,AC3,AC4` at commit `1a2b3c4d`, 2026-09-09 10:00 UTC
- **Lock check**: `git diff 1a2b3c4d..9f3a2c1e -- goals/billing-tiered-pricing.md` shows status and evidence lines only
- **Re-scope entries**: none; scope held from activation to closure

## 6. Closure
- **Outcome**: all 4 criteria satisfied; closed at commit `9f3a2c1e` on 2026-09-15
- **Outcome note**: tiered pricing shipped in release 2.14.0 after staged rollout; finance signed off 2026-09-17
- **Health history**: zero stale flags; weekly checks 2026-09-09 and 2026-09-15 both clean

## 7. Verification Evidence
- [ ] Lock diff output attached showing criteria lines untouched
- [ ] Per-criterion evidence logs linked above and retrievable
- [ ] Closure commit recorded with outcome note
- [ ] `scripts/verify.sh` exits 0 (log attached)
