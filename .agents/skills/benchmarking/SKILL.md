# Statistical Performance Benchmarking & Regression Detection

## Purpose
Design and execute deterministic, statistically sound performance benchmarks (Go benchmarks, pytest-benchmark, k6, hyperfine) with warmup discipline, percentile reporting, and confidence-gated regression detection, so performance claims are evidence, not anecdotes.

## Use when
- Proving a change is faster: before/after benchmarks with confidence intervals for an optimization Goal.
- Setting or enforcing performance budgets: p99 latency ceilings, throughput floors, startup time caps.
- Detecting regressions in CI: benchmark tracking over commits with statistical gating, not single-run comparisons.
- Choosing between implementations: A/B measurement of algorithms, serializers, or query shapes under identical load.

## Do not use when
- Auditing code complexity or duplication without timing anything (use `code-quality`).
- Debugging a functional bug where performance is irrelevant (fix it; do not benchmark it).
- Designing cache topologies to avoid load rather than measure it (use `caching`).
- Running one-off timing with `time` and eyeballing the result (that is not a benchmark; this skill forbids it).

## Required context
- Benchmark harness and versions: Go 1.24.3 `testing` benchmarks, pytest-benchmark 4.1.1, k6 0.57.2, hyperfine 1.18.0.
- Target metric and budget: e.g. invoice rendering p99 at most 120 ms at 500 rps, or JSON marshal at least 180 MB/s.
- Load profile: concurrency, dataset size, warmup iterations, and production traffic shape being simulated.
- Execution environment: pinned machine or runner class, CPU governor, isolation from noisy neighbors.

## Procedure
1. **Fix the environment and harness**:
   - Pin benchmarks to one runner class (`c6i.xlarge` dedicated, or the `bench` GitHub runner label); record CPU model, core count, and governor (`performance`).
   - Disable frequency scaling variance: run `cpupower frequency-set -g performance` on bare metal, or document that cloud runs carry ±5 percent noise.
   - Choose the harness per layer: `go test -bench` for functions, `pytest-benchmark` for Python paths, `k6` for HTTP endpoints, `hyperfine` for CLI startup.
2. **Design the measurement with warmup and repetitions**:
   - Discard warmup explicitly: `-benchtime=5s -count=6` in Go (drop run 1), `--warmup=on --warmup-iterations=3` in pytest-benchmark, k6 `stages` with a 60-second ramp before the 5-minute steady window.
   - Size datasets to production shape: 50,000 invoices with the real size distribution (median 2.1 KB, p99 48 KB), not 10 synthetic rows.
   - Run at least 5 measured repetitions; record min, median, p90, p99, and max for every series.
3. **Analyze with statistics, not vibes**:
   - Compare before/after with relative change plus a 95 percent confidence interval; use `benchstat old.txt new.txt` for Go series.
   - Declare improvement only when the full confidence interval clears the noise floor: e.g. `delta -8.4% ±1.9%` is a win, `delta -1.2% ±2.6%` is noise.
   - For k6 runs, gate on `p(99)<120` thresholds in the script itself (`thresholds: { http_req_duration: ['p(99)<120'] }`) so CI fails on breach.
4. **Guard against benchmark lies**:
   - Verify the compiler did not eliminate the work: consume results via `benchmark.ReportMetric`, `black_box`, or writing checksums; check assembly or instruction counts when a result looks impossibly fast.
   - Run the A/A control: benchmark the same code twice and confirm the delta interval includes zero before trusting any A/B delta.
   - Isolate from neighbors: no parallel CI jobs on the bench runner during measurement; record load average alongside results.
5. **Record and gate**:
   - Store results in `templates/benchmarking-spec.md` with environment, harness flags, raw numbers, and the benchstat verdict.
   - Wire the budget into CI (`k6` thresholds, `benchstat` delta gates) so the next regression fails the build, not a human review.
   - Run the verification script `scripts/verify.sh` from the repo root; it must exit 0.

## Decision rules
- **No Single-Run Claims**: Any latency or throughput claim from fewer than 5 measured repetitions is rejected as evidence.
- **Warmup Is Mandatory**: Benchmarks without explicit warmup and discarded first runs are invalid; cold-start numbers measure page faults, not code.
- **Confidence Interval Decides**: A change ships as faster only when the 95 percent interval excludes zero in the favorable direction.
- **A/A Control Before A/B Verdict**: If the same-code control shows nonzero drift beyond the noise floor, fix the environment before judging the change.
- **Budgets Are Gates, Not Aspirations**: A breached p99 budget fails CI exactly like a failed test; waivers need owner and expiry.

## Evidence required
- Benchmark specification adhering to `templates/benchmarking-spec.md` with environment, flags, and raw series.
- Raw repetition logs: at least 5 measured runs per variant plus the discarded warmup runs.
- `benchstat` (or equivalent) comparison output with deltas and confidence intervals.
- A/A control run proving the environment is stable before the A/B verdict.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Measured verdict per comparison: delta percent, 95 percent confidence interval, and ship/no-ship recommendation.
- CI gate configuration enforcing the budget on every protected-branch run.
- Environment record: runner class, CPU, governor, dataset shape, harness flags.
- Regression report when applicable: offending commit range, metric drift, and bisect notes.

## Stop conditions
- All claimed improvements backed by confidence intervals excluding zero, with A/A control green.
- Performance budgets wired into CI thresholds and passing on the protected branch.
- Regression root-caused to a commit range or formally accepted with owner and revisit date.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to the owning team if the A/A control never stabilizes (noisy shared runners, thermal throttling) and dedicated bench hardware is needed.
- Escalate to Lead Architect if meeting the budget requires an architecture change (new index, protocol change, caching layer) beyond the Goal scope.
- Escalate immediately if benchmarking exposes a correctness bug under load (data races, dropped writes); stop measuring and file the defect.
