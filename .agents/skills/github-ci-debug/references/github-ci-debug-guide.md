# GitHub CI Debugging — Technical Reference Guide

## 1. Core Concepts

**Failure classification** decides the fix: code regression (last-green diff implicates), dependency drift (lockfile vs fresh resolve differs), runner-image change (`ubuntu-22.04` → `ubuntu-24.04` behavior shift), cache poisoning (stale key restores bad artifacts), permission gap (missing `contents:`/`id-token:` scopes), or concurrency artifact (cancelled superseded runs).

**Matrix triage** compares failing vs passing cells: one OS fails → platform path/toolchain issue; one version fails → language-version incompatibility; all fail → shared step (checkout, setup, cache restore).

**Flakiness proof** is statistical: a test failing 3/20 identical reruns with no code change is quarantined with a tracked issue, not retried into greenness.

## 2. Patterns

- **First-error capture**: `gh run view <id> --log-failed` and read the first error block; tail logs mislead.
- **Last-green bisect**: `git log` between last green and first red SHA; revert-suspect or pin-suspect one at a time.
- **Cache-key scoping**: `key: ${{ runner.os }}-deps-${{ hashFiles('**/package-lock.json') }}` plus `restore-keys` fallback; bump a `v2` prefix when poisoning is suspected.
- **Permission minimalism**: declare per-job `permissions:` (e.g. `contents: read`, `pull-requests: write`) instead of workflow-wide write-all.
- **Local reproduction**: `act -j <job>` or the same container image (`catthehacker/ubuntu:full-22.04`) running the exact step commands before editing YAML.

## 3. Anti-Patterns

- Editing five workflow fields at once ("fix" by shotgun).
- Disabling the failing required check to go green.
- Retrying flaky jobs until pass and calling it fixed.
- Pasting full logs with secrets into public issues (scrub `token`, `secrets.*`, OIDC material).

## 4. Worked Example

`ci.yml` job `test (ubuntu-22.04, node 20)` fails with `ERR_OSSL_EVP_UNSUPPORTED` at the webpack step; macOS/Windows cells pass. Classification: runner-image drift — Ubuntu image bumped OpenSSL 3 defaults. Local reproduction with the matching container confirms. Minimal fix: set `NODE_OPTIONS=--openssl-legacy-provider` for that cell via matrix `include`, one-line change. Re-run of the exact job goes green; follow-up ticket migrates webpack off MD4 hashing permanently.

## 5. Verification Pointers

- Diagnosis cites run URL + step name + exit code + first error lines.
- The previously-failing job (not just the workflow) is green after the fix.
- No required check was disabled; flaky tests quarantined with issue links.
