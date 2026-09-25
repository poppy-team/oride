# Performance Benchmark Report Specification Template

## 1. Objective & Budget
- **Question**: Is the tiered-pricing rewrite (PR 481) faster than the 2.13 implementation under production-shaped load?
- **Budget**: invoice rendering p99 at most 120 ms at 500 rps sustained; JSON marshal at least 180 MB/s
- **Decision rule**: ship only if the 95 percent interval for the delta excludes zero favorably

## 2. Environment & Harness
- **Runner**: dedicated `c6i.xlarge`, Intel Xeon 8375C, 4 vCPU, governor `performance`, no co-tenants
- **Software**: Go 1.24.3, k6 0.57.2, benchstat from golang.org/x/perf 2026-06
- **Dataset**: 50,000 invoices, median 2.1 KB, p99 48 KB, production size distribution sampled 2026-09-01
- **Harness flags**: `go test -bench . -benchtime=5s -count=6 -run=^$`; k6 stages 60 s ramp plus 5 min steady at 500 rps

## 3. A/A Control (Environment Stability)
- **Series**: same binary `billing-api@9f3a2c1e` run twice, 30 min apart, interleaved order
- **Result**: `BenchmarkCalculateTieredPrice` A/A delta +0.6 percent ±1.8 percent, interval includes zero
- **Verdict**: environment stable; noise floor ±1.8 percent for this benchmark

## 4. A/B Results

| Benchmark | Old median | New median | Delta | 95 percent CI | Verdict |
|---|---|---|---|---|---|
| BenchmarkCalculateTieredPrice | 41.2 µs/op | 37.7 µs/op | -8.5 percent | -10.1 percent to -6.4 percent | real win, ship |
| BenchmarkRenderInvoice | 88.4 ms/op | 87.9 ms/op | -0.6 percent | -2.4 percent to +1.3 percent | noise, no claim |
| k6 invoice endpoint p99 | 118 ms | 104 ms | -11.9 percent | sustained 5 min window | budget met |

- **Compiler-elimination guard**: outputs folded into a CRC32 checksum asserted per iteration; checksum matched golden `9d04f1aa`
- **Warmup**: run 1 discarded in all Go series (run 1 was 14 percent slower, confirming warmup was load-bearing)

## 5. CI Gates Wired
- **k6 thresholds**: `http_req_duration: ['p(99)<120']`, `http_req_failed: ['rate<0.01']` in `load/invoice-endpoint.js`
- **benchstat gate**: `.github/workflows/perf.yaml` fails when the delta interval overlaps zero unfavorably or p99 exceeds 120 ms
- **Protected branch**: `perf-budget` check required on `main` since 2026-09-14

## 6. Verification Evidence
- [ ] Raw series `old.txt` / `new.txt` (6 runs each, run 1 marked discarded) attached
- [ ] `benchstat old.txt new.txt` output: `delta -8.42% (-9.9%, -6.3%), p=0.008, n=5+5`
- [ ] k6 summary JSON with p50/p90/p99 ladder attached
- [ ] A/A control log attached showing interval including zero
- [ ] `scripts/verify.sh` exits 0 (log attached)
