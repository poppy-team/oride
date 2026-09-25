# CI/CD Pipeline Design, Hardening & Reproducibility

## Purpose
Design, implement, and audit continuous integration and delivery pipelines (GitHub Actions, GitLab CI) so every validation is reproducible from a clean checkout, with hermetic dependencies, safe dependency caching, fail-fast contract gates, and secret-free logs.

## Use when
- Creating or modifying CI workflows: build, lint, test, security scan, artifact publish.
- Debugging flaky or slow pipelines: non-deterministic failures, cache poisoning, runner drift.
- Enforcing merge gates: required checks, branch protection, evidence-backed promotion from staging to production.
- Auditing pipeline security: secret handling, OIDC authentication, third-party action pinning.

## Do not use when
- Writing application code or fixing a product bug (use the language or domain skill; CI only verifies it).
- Building container images beyond what the pipeline needs (use `containers` for image design).
- Cutting a versioned release with changelogs, signing, and rollback plans (use `release-engineering`).
- Managing branches, commits, and pull request review flow (use `git-workflow`).

## Required context
- CI platform and runner fleet: GitHub Actions `ubuntu-24.04` runners or GitLab SaaS `saas-linux-medium-amd64`, plus self-hosted runner labels if any.
- Workflow definitions: `.github/workflows/*.yaml` or `.gitlab-ci.yml` at the pinned commit under change.
- Merge-gate policy: required check names, branch protection rules, environment approvers for staging and production.
- Secret inventory: which secrets exist (names only), their owning environment, and rotation dates.

## Procedure
1. **Pin the execution environment**:
   - Pin runner images by digest or dated label (`ubuntu-24.04`, never `latest`); pin third-party actions by full commit SHA (`actions/checkout@11bd719`).
   - Pin language toolchains with lockfiles checked into the repo (`go.mod`/`go.sum`, `package-lock.json`, `Cargo.lock`) and verify with `go mod verify`, `npm ci`, `cargo fetch --locked` in the job.
2. **Order gates fail-fast by cost**:
   - Stage 1 (under 90 seconds): formatting, lint, secret scan (`gitleaks detect --no-git`), lockfile consistency.
   - Stage 2 (under 10 minutes): unit and integration tests with `go test ./...`, `pytest -q`, `npm test`.
   - Stage 3: SAST/dependency audit (`govulncheck ./...`, `npm audit --audit-level=high`, `trivy fs --severity HIGH,CRITICAL .`).
   - A Stage 1 failure must cancel downstream stages via `if: failure()` guards or `needs:` dependencies so broken commits cost seconds, not minutes.
3. **Cache dependencies safely**:
   - Key caches on lockfile hashes (`hashFiles('go.sum')`, `hashFiles('package-lock.json')`) with OS prefix; never cache across lockfile changes without a key miss.
   - Restore keys must be prefix-scoped so a partial hit cannot poison a build with stale objects older than 7 days.
   - Record cache hit rate and p50 restore time; a cache slower than a clean fetch is deleted, not kept.
4. **Harden secrets and provenance**:
   - Inject secrets only as masked environment variables scoped to the single job that needs them; never echo them, never pass them as action inputs that land in logs.
   - Prefer OIDC federation (`permissions: id-token: write`) over long-lived tokens for cloud deploys and artifact registries.
   - Generate build provenance (SLSA level 2 minimum) and attach SBOMs to release artifacts.
5. **Verify reproducibility from clean checkout**:
   - Re-run the full workflow on a fresh runner with cleared caches (`gh workflow run ci.yaml --ref <sha>` twice) and confirm identical pass/fail outcomes.
   - Run the verification script `scripts/verify.sh` from the repo root; it must exit 0.

## Decision rules
- **Clean-Checkout Reproducibility Is Mandatory**: Any pipeline that passes only with warm caches or local state is broken and must be fixed before merge.
- **Fail Fast on Contract Violations**: Formatting, lint, and secret-scan failures stop the pipeline in Stage 1; expensive tests never run on a contract-broken commit.
- **Pin Everything Executable**: Unpinned actions, floating runner images, and unlocked toolchain installs are prohibited in protected-branch workflows.
- **Secrets Never Touch Logs**: Any job that prints, exports to artifacts, or forwards a secret value fails review regardless of green tests.
- **Required Checks Are Code**: Branch protection and required-check lists live in versioned config (or documented API calls), never as untracked UI clicks.

## Evidence required
- Pipeline specification adhering to `templates/ci-cd-spec.md` with stage timings and gate thresholds.
- Two consecutive clean-run logs from a fresh runner with cleared caches showing identical outcomes.
- Cache audit: keys, hit rates, restore timings, and the 7-day staleness policy.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Versioned workflow files with pinned actions, staged gates, and cache keys.
- Merge-gate configuration: required checks, branch protection, environment approvers.
- Reproducibility report: clean-run logs, timing breakdown per stage, flake rate over the last 50 runs.
- Secret-handling statement: inventory names, scoping, OIDC usage, rotation dates.

## Stop conditions
- Full pipeline green on two consecutive clean-cache runs from a fresh checkout.
- Flake rate below 2 percent across the last 50 runs on the protected branch.
- All required checks wired to branch protection with approvers assigned.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to Platform Lead if runner capacity, self-hosted runner trust, or cloud OIDC federation blocks hardening.
- Escalate to Security if a secret is found in logs, artifacts, or cache contents; rotate immediately and purge affected caches.
- Escalate to the owning team if a third-party action compromise (tag move, repo takeover) forces emergency repinning.
