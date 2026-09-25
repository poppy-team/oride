# Statistical Benchmarking Reference Guide

## 1. Core Concepts & Formulas

### 1.1 Percentiles Over Means
Latency distributions are skewed: a mean of 40 ms can hide a p99 of 900 ms that
violates the SLA for 1 percent of users. Always report the percentile ladder:

$$p50 \le p90 \le p99 \le p99.9 \le \max$$

Budget on the percentile the SLA names (usually p99); report the full ladder so tail
regressions cannot hide behind a stable mean.

### 1.2 Confidence Intervals for Deltas
A benchmark delta is an estimate with error. For two series with means
$\bar{x}_{old}$, $\bar{x}_{new}$ and standard errors $SE_{old}$, $SE_{new}$:

$$\Delta = \frac{\bar{x}_{new} - \bar{x}_{old}}{\bar{x}_{old}}, \quad
CI_{95\%} \approx \Delta \pm 1.96 \sqrt{SE_{old}^2 + SE_{new}^2}$$

In practice, run `benchstat old.txt new.txt` and read the verdict: `delta -8.4%
±1.9%` is a real win; `delta -1.2% ±2.6%` overlaps zero and is noise. Never ship on
a point estimate without its interval.

### 1.3 Warmup and Steady State
Cold runs measure page faults, lazy class loading, and empty JIT caches. The
measurement window must start after steady state:

$$\text{reported} = \text{median}(\text{runs}_{k+1..n}), \quad \text{runs}_{1..k} = \text{warmup (discarded)}$$

Typical $k$: 1 of 6 Go benchmark counts, 3 pytest warmup iterations, 60-second k6
ramp before a 5-minute steady window. If run 1 differs from runs 2–6 by more than
10 percent, warmup was load-bearing and the design is correct to discard it.

### 1.4 A/A Controls and Noise Floors
Run the identical binary twice (A/A). The resulting interval is the environment's
noise floor (typically ±1–3 percent on dedicated hardware, ±5 percent on shared
cloud runners). An A/B delta smaller than the A/A floor is unmeasurable, not a tie.

## 2. Patterns and Anti-Patterns

**Do: gate k6 thresholds in code.**
`thresholds: { http_req_duration: ['p(99)<120'], http_req_failed: ['rate<0.01'] }`
turns the budget into a CI failure instead of a dashboard nobody reads.

**Do: consume every benchmarked value.**
Go: assign to package-level `var sink`; Rust: `criterion::black_box`; Python: assert
on a checksum of the output. An eliminated loop benchmarks nothing at 0.3 ns/op.

**Do not: benchmark on a laptop with Slack open.** Thermal throttling plus 40
background processes produces ±20 percent swings that swallow real 5 percent wins.
Bench on pinned, isolated runners or do not claim numbers.

**Do not: compare across machines or days.** A Tuesday laptop number versus a Friday
server number measures hardware and weather, not code. A/B runs share one machine,
one hour, interleaved order.

**Do not: report only the best run.** Best-of-5 rewards lucky scheduling and punishes
honest measurement. Report median with p90/p99, or the full series.

## 3. Harness Configuration Example

```bash
# Go micro-benchmark with warmup discipline and benchstat verdict
go test ./internal/pricing/ -bench=BenchmarkCalculateTieredPrice \
  -benchtime=5s -count=6 -run=^$ | tee new.txt
git stash -q
go test ./internal/pricing/ -bench=BenchmarkCalculateTieredPrice \
  -benchtime=5s -count=6 -run=^$ | tee old.txt
git stash pop -q
benchstat old.txt new.txt
# expected verdict shape: delta -8.42% (-9.9%, -6.3%), p=0.008, n=5+5
```

```javascript
// k6 endpoint benchmark with in-script budget gates
import http from 'k6/http';
export const options = {
  stages: [
    { duration: '60s', target: 500 },  // ramp + warmup, discarded
    { duration: '5m', target: 500 },   // steady measurement window
  ],
  thresholds: {
    http_req_duration: ['p(99)<120'],
    http_req_failed: ['rate<0.01'],
  },
};
export default function () {
  http.get('https://staging.example.com/api/v1/invoices/INV-2026-04812');
}
```
