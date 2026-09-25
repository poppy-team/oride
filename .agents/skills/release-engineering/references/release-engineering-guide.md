# Release Engineering Reference Guide

## 1. Core Concepts & Formulas

### 1.1 SemVer Decision Procedure
Given the previous tag, classify the largest change: any breaking API or schema
change forces MAJOR; additive backward-compatible features force MINOR; anything
else is PATCH. Pre-release ordering follows `1.0.0-alpha < 1.0.0-alpha.1 <
1.0.0-beta < 1.0.0`. When in doubt between minor and major, choose major; an
undersized major breaks consumers silently, an oversized one merely annoys.

### 1.2 Recovery Time Budget
$$\text{RTO}_{\text{release}} = T_{\text{detect}} + T_{\text{decide}} + T_{\text{execute}} + T_{\text{verify}}$$

Budget 15 minutes total: detection under 5 minutes via halt-condition paging,
decision under 2 minutes (pre-authorized rollback for on-call), execution under
5 minutes (pre-staged artifacts), verification under 3 minutes (production smoke).
Rehearse quarterly; the measured drill time (8 min 40 s) is the only honest value
for $T_{\text{execute}} + T_{\text{verify}}$.

### 1.3 Canary Statistics
A 5 percent canary over 2 hours at 4,000 rps baseline observes ~1.44M requests,
enough to detect a 0.5 percent error-rate regression with 99 percent confidence.
Shorter or smaller canaries cannot distinguish regressions from noise; do not
shorten the canary to "save time" — that trades measurement for hope.

### 1.4 Migration Safety Ordering
Expand-then-contract: deploy code that tolerates both schemas (expand), migrate
data (migrate), then deploy code that requires the new schema (contract). Rollback
at any point before contract is a code revert; after contract, it requires the
tested downgrade path. Never migrate and contract in the same release.

## 2. Patterns and Anti-Patterns

**Do: pre-stage the previous version.** Keep 2.13.4 artifacts, configs, and the
48-to-47 downgrade tested before canary starts. Rollback then means executing a
rehearsed plan, not improvising under pager duty.

**Do: sign and attest everything.** `cosign sign --yes` plus SLSA provenance turns
"did we ship what we built?" from a trust question into a verification command.

**Do not: Friday 17:00 full rollouts.** The rollout calendar respects on-call
coverage: full production promotion happens Tuesday–Thursday mornings, never when
the responders are asleep or the author is on leave.

**Do not: migrate without a snapshot restore test.** A backup that has never been
restored is a rumor. The rehearsal restores the 12 GB anonymized snapshot first,
then migrates — proving both halves of the safety story.

**Do not: bundle migrations with refactors.** A release containing schema change
48 plus a pricing rewrite cannot be bisected when the canary burns. One risky axis
per release; sequence the rest.

## 3. Release Command Example

```bash
# Cut, sign, and verify release 2.14.0 from a clean checkout
git tag -s v2.14.0 -m "Release 2.14.0: tiered pricing API fields"
git verify-tag v2.14.0
goreleaser release --clean --timeout 30m
sha256sum dist/* > SHA256SUMS && sha256sum -c SHA256SUMS
cosign sign --yes registry.example.com/billing-api:2.14.0
syft registry.example.com/billing-api:2.14.0 -o spdx-json > sbom.spdx.json

# Migration rehearsal both directions on the staging snapshot
./scripts/db-restore --snapshot anonymized-2026-09-06.dump --target staging
./scripts/migrate --to 48 --target staging   # expect ~4 min, reconcile row counts
./scripts/migrate --to 47 --target staging   # downgrade must round-trip canary tables
```
