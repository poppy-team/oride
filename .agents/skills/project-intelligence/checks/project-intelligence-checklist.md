# Project Intelligence — Verification Checklist

## 1. Estimate Integrity
- [ ] Estimate (hours, tokens, complexity class) is timestamped before work starts.
- [ ] No estimate is backfilled or edited after completion; corrections append a new line.
- [ ] Complexity class S/M/L follows the documented rubric, not ad-hoc labels.

## 2. Measurement Honesty
- [ ] Actual hours come from wall-clock logs, actual tokens from the token ledger.
- [ ] Test counts, review rounds, and reopen flags are recorded as observed.
- [ ] Variance is reported with sign for every metric; no metric is silently dropped.

## 3. Calibration Discipline
- [ ] Calibration factors derive from class means over at least 5 historical tasks.
- [ ] Raw and calibrated estimates are both recorded for the next task.
- [ ] Single-outlier adjustments are rejected; class statistics govern.

## 4. Ledger Hygiene
- [ ] Every entry validates against the intelligence JSONL schema (required fields present, types correct).
- [ ] Entries carry task ids and classes only; no individual attribution fields exist.
- [ ] Weekly aggregation runs and publishes class means for hours and tokens.

## 5. Verification Gates
- [ ] Variance formula applied uniformly: (actual - estimate) / estimate * 100.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
