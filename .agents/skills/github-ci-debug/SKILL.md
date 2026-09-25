# GitHub CI Debugging

## Purpose
Diagnose GitHub Actions failures systematically — failed steps, matrix mismatches, flaky tests, cache poisoning, runner drift — and land minimal, verified fixes with the failing job linked as evidence.

## Use when
- A workflow run fails (red checks) and the cause is unknown or disputed.
- Matrix jobs fail on a subset of OS/version combinations while others pass.
- A pipeline is flaky: same commit passes and fails intermittently.
- Operating in mode(s): `implementation`, `testing`.

## Do not use when
- Authoring new workflows from scratch without a failure to diagnose (use CI authoring flows).
- Reviewing PR code quality (use `github-pr-review`).
- Managing repository settings or branch protection rules (use `github-repository`).

## Required context
- Failing workflow run URL and the exact failed step log excerpt.
- Workflow YAML (`.github/workflows/*.yml`), matrix definition, and runner labels.
- Recent changes: commits since last green run, dependency or runner-image bumps.

## Procedure
1. **Capture the failure precisely**: record run URL, job name, step name, exit code, and the first error lines — not just the tail output. Link the run in the diagnosis.
2. **Bisect the cause**: compare against the last green run; classify as code regression, dependency drift (lockfile vs fresh resolve), runner-image change, cache poisoning, secret/permission gap, or concurrency/cancellation artifact.
3. **Reproduce locally first**: replicate with `act` or the same container image and commands before editing the workflow; a fix without local reproduction is a guess.
4. **Fix minimally**: pin the drifted dependency, correct the matrix include/exclude, add the missing permission (`contents: read`, `id-token: write`), scope the cache key with lockfile hash, or quarantine the flaky test with a tracked issue — one change per hypothesis.
5. **Verify on the branch**: re-run the exact failing job via `gh run rerun` or a draft PR; require the previously-failing job green plus no new failures before closing.
6. **Harden against recurrence**: add timeout-minutes, fail-fast policy, cache-key versioning, and a flaky-test quarantine note where applicable.

## Decision rules
- **Log evidence over intuition**: every diagnosis cites step logs with timestamps; "probably flaky" without a rerun history is not a diagnosis.
- **One hypothesis per change**: bundled workflow edits that mix fixes are prohibited; bisect instead.
- **Never weaken required checks**: disabling a failing required check to go green is forbidden; fix or formally waive with owner approval.
- **Secrets stay masked**: logs pasted into issues must be scrubbed of tokens and OIDC material.

## Evidence required
- Failing run URL + step log excerpt and the passing re-run URL after the fix.
- Local reproduction notes (commands used, container image).
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Structured GitHub artifact (diagnosis comment or fix PR with failure classification).
- Validation confirmation (green re-run of the exact failing job).

## Stop conditions
- Previously-failing job is green on the branch with cause classified and linked.
- Flaky source quarantined with a tracked follow-up issue if not yet fixed.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to repository owner if the fix requires organization-level runner, secret, or policy changes.
- Escalate to the test owner when a flaky test needs quarantine or deletion approval.
