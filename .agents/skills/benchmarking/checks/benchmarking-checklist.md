# Performance Benchmarking — Verification Checklist

## 1. Environment Control
- [ ] Benchmarks pinned to one runner class with recorded CPU model, core count, and governor (`performance` where applicable).
- [ ] No parallel CI jobs share the bench runner during measurement; load average recorded alongside results.
- [ ] Harness versions pinned: Go 1.24.3, pytest-benchmark 4.1.1, k6 0.57.2, or hyperfine 1.18.0 as applicable.
- [ ] Dataset matches production shape (50,000 invoices, median 2.1 KB, p99 48 KB), not toy inputs.

## 2. Warmup & Repetition Discipline
- [ ] Explicit warmup configured and discarded: Go `-count=6` dropping run 1, pytest `--warmup-iterations=3`, k6 60-second ramp before the steady window.
- [ ] At least 5 measured repetitions per variant; min, median, p90, p99, and max all recorded.
- [ ] Compiler-elimination guards in place: results consumed via `black_box`, checksums, or `ReportMetric`; suspicious speedups inspected.
- [ ] A/A control (same code twice) shows a delta interval including zero before any A/B verdict is trusted.

## 3. Statistical Verdicts
- [ ] Before/after comparison reports relative delta with a 95 percent confidence interval (`benchstat old.txt new.txt`).
- [ ] Improvement declared only when the full interval clears the noise floor in the favorable direction.
- [ ] Percentile budgets (p50, p90, p99) reported, never means alone; tail behavior explicitly addressed.
- [ ] k6 threshold gates (`p(99)<120`) encoded in the script so breaches fail CI automatically.

## 4. Budgets & Regression Gates
- [ ] Performance budgets (p99 ceiling, throughput floor, startup cap) documented with the traffic shape they assume.
- [ ] Budget checks run on every protected-branch CI execution, not as manual side quests.
- [ ] Regressions bisected to a commit range with metric drift quantified, or formally accepted with owner and revisit date.
- [ ] No single-run `time`-based claim accepted anywhere in the report.

## 5. Evidence & Sign-Off
- [ ] Benchmark specification follows `templates/benchmarking-spec.md` with environment, flags, and raw series.
- [ ] Raw repetition logs attached, including discarded warmup runs and the A/A control.
- [ ] CI gate configuration for the budget committed and green on the protected branch.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
