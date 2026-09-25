# Release Engineering — Verification Checklist

## 1. Versioning & Scope Freeze
- [ ] Version assigned by SemVer rules (2.13.4 to 2.14.0 justified by additive API fields, no breaking change).
- [ ] Changelog generated from conventional commits; every commit classified, none unclassified.
- [ ] Release tag signed (`git tag -s v2.14.0`) and signature verified with `git verify-tag`.
- [ ] Scope freeze recorded: no feature commits land after the tag without a new version.

## 2. Artifact Integrity
- [ ] Artifacts built from the tag in a clean environment (`goreleaser release --clean` or documented `make dist`).
- [ ] `SHA256SUMS` generated with `sha256sum` and verified on a pristine machine with `sha256sum -c`.
- [ ] Cosign signatures present (keyless OIDC) and verified; SBOM and SLSA provenance attached per artifact.
- [ ] Registry install smoke passes from the published location before the release is announced.

## 3. Migration Safety Both Directions
- [ ] Upgrade rehearsal 47 to 48 on the production-shaped snapshot (12 GB) completed in 4 min 12 s with row counts reconciled.
- [ ] Downgrade rehearsal 48 to 47 confirms lossless round-trip on canary tables, or a dated waiver with owner is filed.
- [ ] Migration exercised under 1.5× peak write load with no lock-escalation deadlocks.
- [ ] Migration failure wiring pages on-call; partial migrations are resumable or fully rolled back, never half-applied silently.

## 4. Staged Rollout & Halt Conditions
- [ ] Stages executed in order: 24-hour staging soak, 5 percent canary 2 hours, 50 percent 6 hours, then full rollout.
- [ ] Halt conditions encoded and paging: error rate above 1 percent for 5 minutes, p99 above 200 ms for 10 minutes.
- [ ] Each stage promotion recorded with timestamp and operator; no silent auto-promotion past canary.
- [ ] Production smoke suite green within 30 minutes of full rollout.

## 5. Rollback Readiness & Evidence
- [ ] Previous-version artifacts (2.13.4) pre-staged before canary; rollback completes within the 15-minute budget.
- [ ] Full rollback rehearsed quarterly on staging; last drill measured 8 min 40 s on 2026-08-19.
- [ ] Release specification follows `templates/release-engineering-spec.md` with hashes and decisions.
- [ ] Automated verification script `scripts/verify.sh` executes with exit code 0.
