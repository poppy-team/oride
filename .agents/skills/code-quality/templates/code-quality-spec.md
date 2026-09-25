# Code Quality Audit Report Specification Template

## 1. Audit Scope & Baseline
- **Repository**: billing-service, commit 9f3a2c1e, branch release/2.14
- **Toolchains**: Go 1.24.3, Python 3.12.4, Node 22.6.0
- **Analyzer versions**: golangci-lint 2.1.6, ruff 0.11.9, eslint 9.26.0, jscpd 4.0.5
- **Baseline date**: 2026-09-10, measured from a clean checkout
- **Scope**: `internal/billing/`, `internal/invoices/`, `src/billing/` (47,300 non-blank lines)
- **Excluded paths**: `internal/mocks/`, `src/generated/`, `migrations/` with config references

## 2. Thresholds & Policy
- **Cyclomatic ceiling**: 10 per function; cognitive ceiling 15 per function
- **Duplication ceiling**: 3 percent of non-blank lines per module
- **Lint gate**: zero findings above info severity on `golangci-lint`, `ruff`, `eslint`, `clippy`
- **Slice budget**: at most 200 changed lines per remediation slice

## 3. Hotspot Backlog (Ranked by Risk)

| Rank | Location | Metric | Churn 90d | Defects 90d | Risk score |
|---|---|---|---|---|---|
| 1 | `internal/billing/pricing.go:calculateTieredPrice` | M=22, cognitive 31 | 34 commits | 5 defects | 22 × 3.56 × 6 = 470 |
| 2 | `src/billing/taxHelpers.ts:applyRegionalTax` | duplicated 4×, 61 lines each | 21 commits | 3 defects | clone mass 244 lines |
| 3 | `internal/invoices/render.go:RenderInvoice` | M=14, funlen 132 lines | 12 commits | 1 defect | 14 × 2.56 × 2 = 72 |

## 4. Remediation Slices

### Slice 1 — pricing.go calculateTieredPrice
- **Technique**: Extract Function into `resolveBaseTier` (M=4), `applyVolumeDiscount` (M=6), `applyPromoCode` (M=5)
- **Before**: M=22, cognitive 31, 118 lines
- **After**: M=6 max per piece, cognitive 8 max, 124 lines total
- **Tests**: `go test ./internal/billing/ -run TestCalculateTieredPrice -v` — 18/18 pass
- **Changed lines**: 142

### Slice 2 — applyRegionalTax clones
- **Technique**: Extract shared `roundTax(amount, region)` helper in `src/billing/taxCore.ts`; 4 call sites updated
- **Before**: duplication 4.1 percent in `src/billing/`
- **After**: duplication 1.2 percent in `src/billing/`
- **Tests**: `npx vitest run src/billing` — 64/64 pass
- **Changed lines**: 96

## 5. Verification Evidence
- [ ] Baseline logs: `golangci-lint run ./...` reported 9 findings (4 high, 5 medium); `jscpd` reported 4.1 percent duplication in `src/billing/`
- [ ] Post-remediation logs: zero findings above info; `jscpd` reports 1.2 percent duplication
- [ ] `go test ./internal/billing/ ./internal/invoices/` — all packages pass
- [ ] `npx vitest run src/billing` — 64/64 pass
- [ ] `scripts/verify.sh` exits 0 (log attached)

## 6. Residual Tech Debt
- `internal/invoices/render.go:RenderInvoice` (M=14) accepted until template engine replacement lands; owner Marina Duarte, revisit 2026-12-01.
- Legacy `src/billing/legacyAdapter.ts` excluded from the duplication gate; scheduled for deletion in release 2.16.
