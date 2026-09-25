# Documentation Change Specification Template

## 1. Change Scope & Audience
- **Pages touched**: `docs/runbooks/billing-recovery.md` (canonical), `docs/api/billing-flags.md` (canonical), `docs/index.md` (projection link update)
- **Audience**: billing operators for the runbook; integrators for the flags reference
- **Reading goals**: restore invoice rendering in under 15 minutes; configure tiered pricing flags without source diving
- **Owners**: runbook Marina Duarte (last review 2026-09-10, next 2026-12-10); flags reference Rafael Costa (last review 2026-09-12, next release 2.15)

## 2. Authority Roles
- **Canonical**: `billing-recovery.md`, `billing-flags.md` — normative; code conforms to these
- **Projection**: `docs/index.md` — index links regenerated; no hand-written content added
- **Historical touched**: none (frozen pages left untouched)

## 3. Content Summary
- **Runbook**: rewrote restore procedure example-first: 6 copy-paste steps with expected outputs (healthy probe `200 OK in 41 ms`, migration `47→48 in 4 min 12 s`)
- **Flags reference**: documented `--dry-run` (added in 2.14.0) and corrected `max-tier` default 6 → 8 with code citation `internal/billing/config.go:42`
- **Size**: runbook 118 lines, flags reference 96 lines, both under the 150-line ceiling

## 4. Drift Check Results
- **Commands verified**: `./scripts/migrate --to 48 --target staging`, `curl -s localhost:8080/ready`, `billing-api --dry-run` — all reproduce verbatim 2026-09-15
- **Flags verified**: `max-tier` default 8 confirmed in code; `--dry-run` exit codes 0/2 confirmed
- **Paths verified**: `test -f` passes for all 14 documented paths; 2 stale paths fixed (`scripts/old-seed.sh` → `scripts/db-seed.sh`)
- **Drift defects filed**: DOC-118 (webhook retry table outdated, owner Rafael, fix-by 2.15); otherwise clean bill

## 5. Validation Logs
- **markdownlint-cli2**: zero findings on `docs/**/*.md` under repo config
- **lychee**: 214 links checked, zero broken (198 internal, 16 external)
- **Site build**: `npm run build` in `docs-site/` (Astro 4.16), zero errors, 38 pages emitted
- **Preview**: changed pages render under `latest` and pinned `v2.14` without layout breakage

## 6. Review Schedule
- **Runbook**: next review 2026-12-10 (quarterly cadence)
- **Flags reference**: next review at release 2.15 (per-release cadence)
- **Stale sweep**: 3 pages past review date reaffirmed (vision, glossary) with updated dates

## 7. Verification Evidence
- [ ] Drift grep logs and `test -f` path results attached
- [ ] lychee report and site-build log attached
- [ ] DOC-118 filed with owner and fix-by version
- [ ] `scripts/verify.sh` exits 0 (log attached)
