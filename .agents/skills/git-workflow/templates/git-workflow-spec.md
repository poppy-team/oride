# Pull Request & Branch Record Specification Template

## 1. Change Identity
- **Branch**: `feat/billing-tiered-pricing-481` from `origin/main` at `1a2b3c4d` (fetched 2026-09-15 08:55 UTC)
- **PR**: billing-api#482, squash-merge policy, title `feat(billing): add tiered pricing API fields`
- **Size**: 8 commits, 312 changed lines, all conventional format
- **Reviewers**: Marina Duarte (CODEOWNERS for `internal/billing/`, approved 2026-09-15 16:40 UTC), Rafael Costa (approved 2026-09-15 15:02 UTC)

## 2. Commit List (Each Builds Green)
- `feat(billing): add TieredPrice struct with validation` — 96 lines, `go test ./internal/billing/` 18/18
- `feat(billing): compute volume discounts per tier` — 88 lines, 24/24
- `fix(billing): round tiered tax half-up at line-item level` — 41 lines, 31/31
- `refactor(billing): extract discount helpers without behavior change` — 52 lines, identical 31/31 before/after
- `test(billing): cover promo-code stacking edge cases` — 35 lines, 38/38

## 3. Acceptance Mapping
- **AC1** tiered prices compute within 0.01 of golden fixtures → `TestCalculateTieredPrice` 18/18 plus fixture diff clean
- **AC2** tax rounds half-up per line item → `TestRoundTieredTax` 12/12, cases `0.125→0.13`, `2.675→2.68`
- **AC3** p99 render latency at most 120 ms at 500 rps → k6 run 2026-09-15, p99 104 ms (summary attached)
- **AC4** no public API breakage → `go-apidiff v2.13.0` reports zero breaking changes

## 4. Quality & Risk Notes
- **Coverage delta**: +1.8 percent on `internal/billing` (87.4 to 89.2 percent)
- **`git diff --check`**: clean; **`gitleaks protect --staged`**: clean at every commit
- **Risk**: medium — pricing math; mitigated by golden fixtures plus canary rollout in release 2.14.0
- **Rollback pointer**: revert squash commit `9f3a2c1e` on `main`; no migration in this change

## 5. Merge & Cleanup
- **CI**: run 4821 green (contract 68 s, test 6 min 40 s, scan 3 min 10 s)
- **Merge**: squash as `feat(billing): add tiered pricing API fields (#482)` on 2026-09-15 17:05 UTC
- **Cleanup**: branch deleted, no worktrees left (`git worktree list` shows `/home/marina/prumo` only), zero stashes

## 6. Verification Evidence
- [ ] PR link and approval timestamps recorded above
- [ ] Per-commit test logs attached (5/5 green)
- [ ] k6 p99 summary and apidiff report attached
- [ ] `scripts/verify.sh` exits 0 (log attached)
