# CI/CD Pipeline Specification Template

## 1. Platform & Scope
- **Platform**: GitHub Actions, runners `ubuntu-24.04` (SaaS, 4 vCPU)
- **Workflows**: `.github/workflows/ci.yaml`, `.github/workflows/security.yaml`, `.github/workflows/release.yaml`
- **Protected branches**: `main` (required checks: contract, test, scan), `release/2.x` (plus release-dry-run)
- **Baseline date**: 2026-09-08; measured over the last 50 runs on `main`

## 2. Stage Gates & Timings

| Stage | Jobs | p50 duration | Budget | Cancel downstream on failure |
|---|---|---|---|---|
| 1 contract | gofmt, go mod verify, golangci-lint, gitleaks | 68 s | 90 s | yes |
| 2 test | go test ./... -count=1, pytest -q | 6 min 40 s | 10 min | yes |
| 3 scan | govulncheck, trivy fs HIGH/CRITICAL | 3 min 10 s | 5 min | n/a (last stage) |

- **Flake rate**: 1 of last 50 runs (2.0 percent) — quarantined `TestBillingRace` fixed 2026-09-05
- **Clean-cache verification**: runs 4821 and 4822 on fresh runners, caches purged, both green with identical 10 min 12 s total

## 3. Pinning & Lockfiles
- **Actions**: `actions/checkout@11bd719` (v4.2.2), `actions/setup-go@d35c59a` (v5.5.0), `gitleaks/gitleaks-action@ff98106` (v2.3.9)
- **Toolchains**: Go 1.24.3 via `setup-go`, Python 3.12.4 via `setup-python`, Node 22.6.0 via `setup-node`
- **Lockfiles enforced**: `go.sum` (go mod verify), `package-lock.json` (npm ci), `Cargo.lock` (cargo fetch --locked)

## 4. Cache Audit
- **Go modules**: key `linux-go-9f3a2c1e` from `hashFiles('go.sum')`; hit rate 94 percent; p50 restore 22 s vs 74 s clean fetch
- **npm**: key `linux-node-77bd01aa` from `hashFiles('package-lock.json')`; hit rate 89 percent; p50 restore 18 s
- **Staleness policy**: objects older than 7 days purged weekly by scheduled `gh cache delete` job
- **Poison drill**: full purge to green rebuild completed in 4 min 30 s on 2026-08-28

## 5. Secrets & Provenance
- **Secrets in scope**: `REGISTRY_PASSWORD` (release job only), `CODECOV_TOKEN` (test job only)
- **OIDC**: registry pushes and cloud staging deploys use `permissions: id-token: write`; zero long-lived deploy tokens
- **Provenance**: SLSA level 2 attestations on all release artifacts; SBOM attached as `sbom.spdx.json`
- **Rotation**: registry credential rotated 2026-08-01; next rotation due 2026-11-01

## 6. Verification Evidence
- [ ] Clean-run logs for runs 4821 and 4822 attached with identical outcomes
- [ ] Branch protection screenshot or API dump showing required checks on `main`
- [ ] `gitleaks detect --no-git` Stage 1 log with zero findings
- [ ] `scripts/verify.sh` exits 0 (log attached)
