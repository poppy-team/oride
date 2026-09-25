# Release Engineering, Packaging & Rollback Readiness

## Purpose
Cut safe, traceable software releases: SemVer versioning, reproducible packaging, checksum and signature verification, migration testing with downgrade paths, staged rollouts, and rehearsed rollback plans with measured recovery time.

## Use when
- Preparing a versioned release: version bump, changelog, artifact build, publish to a registry.
- Verifying release integrity: checksums, cosign signatures, SBOM attachment, provenance attestation.
- Testing data migrations: upgrade and downgrade rehearsals against production-shaped snapshots.
- Planning rollouts and rollbacks: staged traffic shifting, automatic halt conditions, recovery time budgets.

## Do not use when
- Designing CI stage ordering or merge gates without shipping a version (use `ci-cd`).
- Building or hardening the container image itself (use `containers`).
- Writing user documentation or changelogs as prose artifacts (use `documentation`; this skill owns the release mechanics, not the docs site).
- Managing day-to-day branches and pull requests (use `git-workflow`).

## Required context
- Release scope: version number under SemVer (e.g. 2.14.0), included change list, target environments (staging, production).
- Artifact inventory: binaries, container images, migration bundles, and their destination registries.
- Migration surface: schema versions involved (e.g. 47 to 48), downgrade support matrix, production-shaped snapshot for rehearsal.
- Rollback budget: maximum acceptable recovery time (e.g. 15 minutes) and the on-call rotation covering the release window.

## Procedure
1. **Freeze scope and version**:
   - Assign the version by SemVer rules: breaking API change increments major, additive feature increments minor, fixes increment patch (2.13.4 to 2.14.0 for the new tiered-pricing API fields).
   - Generate the changelog from conventional commits (`git log v2.13.0..HEAD --oneline`) and classify every entry; unclassified commits block the release.
   - Tag the exact commit (`git tag -s v2.14.0 -m "Release 2.14.0"`) and verify the tag signature with `git verify-tag v2.14.0`.
2. **Build reproducible, signed artifacts**:
   - Build from the tag in a clean environment: `goreleaser release --clean --timeout 30m` or the documented `make dist` path; record artifact hashes with `sha256sum dist/* > SHA256SUMS`.
   - Sign artifacts with cosign keyless OIDC and attach the syft SBOM plus SLSA provenance to each publishable unit.
   - Verify installability from the registry on a pristine machine: download, `sha256sum -c SHA256SUMS`, and run the smoke suite before announcing.
3. **Rehearse migrations both directions**:
   - Restore the production-shaped snapshot (12 GB, anonymized 2026-09-06) into staging; run upgrade migration 47 to 48 and record duration (4 min 12 s) plus row counts before/after.
   - Run the downgrade path 48 to 47 and confirm data round-trips losslessly on the canary tables; a migration without a tested downgrade ships only with a dated waiver.
   - Exercise the migration under 1.5× peak write load to catch lock-escalation deadlocks before production.
4. **Roll out in stages with halt conditions**:
   - Stage 1: internal staging soak 24 hours. Stage 2: 5 percent canary for 2 hours. Stage 3: 50 percent for 6 hours. Stage 4: full rollout.
   - Automatic halt and page on: error rate above 1 percent for 5 minutes, p99 latency above 200 ms for 10 minutes, or any migration failure.
   - Each stage requires an explicit promote decision recorded with timestamp and operator; silent auto-promotion past canary is prohibited.
5. **Keep rollback hot and rehearsed**:
   - Pre-stage the previous version artifacts (2.13.4) and the tested downgrade path before starting Stage 2; rollback must complete within the 15-minute budget.
   - Rehearse the full rollback quarterly on staging and record the measured recovery time (last drill: 8 min 40 s on 2026-08-19).
   - Run the verification script `scripts/verify.sh` from the repo root; it must exit 0.

## Decision rules
- **No Untested Downgrade Ships Quietly**: Migrations without a rehearsed downgrade require a dated waiver with owner; otherwise they block the release.
- **Stages Promote Explicitly**: Every stage transition is a recorded human or policy decision; no silent auto-promotion past canary.
- **Halt Conditions Are Code**: Error-rate, latency, and migration-failure thresholds live in the rollout config and page automatically; dashboard-watching is not a control.
- **Checksums and Signatures on Everything Publishable**: Unsigned or unchecksummed artifacts never reach a registry or download page.
- **Rollback Budget Is a Release Gate**: If the rehearsed recovery time exceeds 15 minutes, fix the rollback path before shipping forward.

## Evidence required
- Release specification adhering to `templates/release-engineering-spec.md` with version, artifacts, and hashes.
- Signed tag verification log (`git verify-tag`) and artifact signature verification logs.
- Migration rehearsal logs both directions with durations, row counts, and load-test notes.
- Staged rollout record with per-stage promote decisions and halt-condition wiring.
- Passing execution log from `scripts/verify.sh`.

## Output contract
- Published release: version tag, changelog, signed artifacts with checksums, SBOM, and provenance.
- Migration report: upgrade/downgrade timings, data-integrity confirmation, waiver if one direction is unsupported.
- Rollout and rollback plan with stage gates, halt thresholds, and measured recovery time.
- Post-release verification: smoke results from production within 30 minutes of full rollout.

## Stop conditions
- Release published with verified signatures, rehearsed migrations both directions, and staged rollout complete with production smoke green.
- Rollback rehearsed within the 15-minute budget with previous-version artifacts pre-staged.
- Release formally aborted with artifacts yanked, versions marked, and a recorded abort reason.
- Token budget exhausted.
- Blocked on external dependency.

## Escalation rules
- Escalate to Release Captain if halt conditions trigger mid-rollout; the default action is halt and assess, never push through.
- Escalate to DBA if a migration rehearsal shows lock escalation, data loss, or downgrade impossibility on production-shaped data.
- Escalate to Security if artifact signing fails, a checksum mismatch appears, or provenance cannot be generated.
