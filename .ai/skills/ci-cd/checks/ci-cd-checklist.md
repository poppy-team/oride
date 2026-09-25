# CI/CD Pipeline — Verification Checklist

## 1. Reproducibility From Clean Checkout
- [ ] Full pipeline passes on a fresh runner with all caches cleared, twice consecutively with identical outcomes.
- [ ] Runner images are pinned (`ubuntu-24.04`, never `latest`); third-party actions pinned by full commit SHA.
- [ ] Language toolchains install from lockfiles (`go mod verify`, `npm ci`, `cargo fetch --locked`); no floating installs.
- [ ] No job depends on local state, prior workflow artifacts, or warm caches to pass.

## 2. Fail-Fast Gate Ordering
- [ ] Stage 1 (format, lint, secret scan, lockfile check) completes in under 90 seconds and cancels downstream stages on failure.
- [ ] Stage 2 (unit/integration tests) completes in under 10 minutes with deterministic, non-flaky results.
- [ ] Stage 3 (SAST, dependency audit, container scan) runs `govulncheck`, `npm audit --audit-level=high`, or `trivy fs --severity HIGH,CRITICAL`.
- [ ] Flake rate is below 2 percent across the last 50 runs on the protected branch.

## 3. Dependency Caching Safety
- [ ] Cache keys include lockfile hashes (`hashFiles('go.sum')`) with an OS prefix; lockfile changes force a key miss.
- [ ] Restore keys are prefix-scoped; no cache object older than 7 days is reused.
- [ ] Cache hit rate and p50 restore time are recorded; caches slower than clean fetch are removed.
- [ ] Cache poisoning drill documented: purge procedure (`gh cache delete`) completes in under 5 minutes.

## 4. Secrets & Supply-Chain Hardening
- [ ] Secrets are masked, job-scoped environment variables; none appear in logs, artifacts, or action inputs.
- [ ] Cloud deploys and registry pushes use OIDC federation (`permissions: id-token: write`), not long-lived tokens.
- [ ] Release artifacts carry SLSA level 2 provenance and an attached SBOM.
- [ ] `gitleaks detect --no-git` (or equivalent secret scan) runs in Stage 1 and blocks on any finding.

## 5. Merge Gates & Evidence
- [ ] Branch protection requires the named checks; required-check lists are versioned, not untracked UI clicks.
- [ ] Staging and production environments have named approvers; deploys without approval are impossible.
- [ ] Pipeline specification follows `templates/ci-cd-spec.md` with per-stage timings and thresholds.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
