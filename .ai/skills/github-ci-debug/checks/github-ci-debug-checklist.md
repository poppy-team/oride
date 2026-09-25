# GitHub CI Debugging — Verification Checklist

## 1. Failure Capture & Classification
- [ ] Failing run URL, job name, step name, exit code, and first error lines recorded and linked.
- [ ] Cause classified: code regression, dependency drift, runner-image change, cache poisoning, permission gap, or concurrency artifact.
- [ ] Last green run identified; suspect delta (commits, bumps, image changes) listed.

## 2. Reproduction & Minimal Fix
- [ ] Failure reproduced locally (`act` or matching container image) before editing workflow YAML.
- [ ] One hypothesis per change; matrix include/exclude, pins, permissions, or cache keys edited surgically.
- [ ] No required check disabled; pasted logs scrubbed of secrets and OIDC material.

## 3. Verification & Hardening
- [ ] Previously-failing job re-run green on the branch with no new failures.
- [ ] Flaky tests quarantined with tracked follow-up issues where applicable.
- [ ] Hardening applied: `timeout-minutes`, fail-fast policy, lockfile-hashed cache keys.

## 4. Evidence & Sign-Off
- [ ] Failing and passing run URLs linked in the diagnosis artifact.
- [ ] `scripts/verify.sh` exits 0 from the repository root.
