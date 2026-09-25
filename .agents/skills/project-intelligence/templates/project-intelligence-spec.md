# Intelligence Report Specification

## 1. Reporting Window
- **Week**: 2026-09-15 through 2026-09-21
- **Ledger**: evidence/intelligence.jsonl, schema v3, 14 tasks closed
- **Classes covered**: S (6 tasks), M (9 tasks after this week), L (2 tasks, below calibration minimum)

## 2. Closed TaskEntries

| Task | Class | Est hrs / Act hrs | Est tokens / Act tokens | Reviews | Reopens |
|---|---|---|---|---|---|
| T-338 Cache TTL jitter | S | 3.0 / 2.5 | 12,000 / 10,500 | 1 | 0 |
| T-342 Token-refresh race | M | 6.0 / 8.5 | 25,000 / 31,000 | 1 | 0 |
| T-345 llms.txt rollout | S | 2.0 / 2.2 | 9,000 / 9,800 | 2 | 0 |

## 3. Variance Summary

| Class | N | Mean hrs variance | Mean tokens variance | Reopen rate |
|---|---|---|---|---|
| S | 6 | -4.0% | -6.5% | 0% |
| M | 9 | +38.0% | +21.0% | 11% |
| L | 2 | n/a (below minimum of 5) | n/a | 0% |

## 4. Calibration Factors Applied Next Week
- **Class S**: factor 1.00 (mean within noise band of +/-10%, no adjustment).
- **Class M**: factor 1.38 on hours, 1.21 on tokens; e.g. raw 10-hour guess calibrates to 13.8 hours.
- **Class L**: uncalibrated with warning; sample too small.

## 5. Verification Evidence
- [ ] All 14 ledger lines validate against schema v3.
- [ ] Variance recomputed independently: T-342 hours (8.5-6.0)/6.0 = +41.7%.
- [ ] No backfilled estimates: all estimate timestamps precede first commit timestamps.
- [ ] `scripts/verify.sh` exits 0.
