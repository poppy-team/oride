# Release Specification Template

## 1. Release Identity & Scope
- **Version**: 2.14.0 (minor: additive tiered-pricing API fields, backward compatible with 2.13.x)
- **Tag commit**: `9f3a2c1e`, signed tag `v2.14.0`, `git verify-tag` passes
- **Scope**: 34 commits since v2.13.0, all conventional-commit classified; 6 features, 21 fixes, 7 chores
- **Target environments**: staging (soak 24 h), production canary 5 percent, 50 percent, full
- **Release window**: 2026-09-16 09:00–13:00 UTC; on-call Marina Duarte plus DBA Rafael Costa for migration

## 2. Artifacts & Integrity
- **Container**: `registry.example.com/billing-api:2.14.0`, digest `sha256:9c4e1a2b...7e8f90a1`
- **Binaries**: `billing-api-linux-amd64`, `billing-api-darwin-arm64` via `goreleaser release --clean`
- **Checksums**: `SHA256SUMS` verified with `sha256sum -c` on pristine VM `release-check-03`
- **Signatures**: cosign keyless OIDC, identity `.../workflows/release.yaml@refs/tags/v2.14.0`
- **SBOM/provenance**: `sbom.spdx.json` (63 packages), SLSA level 2 attestation attached

## 3. Migration Report (47 to 48)
- **Upgrade rehearsal**: 4 min 12 s on 12 GB anonymized snapshot; invoices 50,000 rows, line_items 212,400 rows reconciled
- **Downgrade rehearsal**: 48 to 47 in 2 min 05 s; canary tables round-trip losslessly (checksums match)
- **Load rehearsal**: 1.5× peak writes during upgrade, zero deadlocks, max lock wait 340 ms
- **Waivers**: none; both directions tested

## 4. Rollout Record
- **Stage 1 staging soak**: 2026-09-16 09:00 UTC, 24 h, zero errors above baseline; promoted by Marina Duarte 2026-09-17 09:12 UTC
- **Stage 2 canary 5 percent**: 2 h, error rate 0.12 percent, p99 96 ms; promoted 11:20 UTC
- **Stage 3 half traffic**: 6 h, error rate 0.10 percent, p99 101 ms; promoted 17:35 UTC
- **Stage 4 full rollout**: 2026-09-17 17:40 UTC; production smoke 42/42 green by 18:05 UTC
- **Halt wiring**: error rate above 1 percent for 5 min pages; p99 above 200 ms for 10 min pages; neither triggered

## 5. Rollback Plan
- **Pre-staged**: 2.13.4 artifacts, configs, and 48-to-47 downgrade script verified before canary
- **Budget**: 15 min recovery; last quarterly drill measured 8 min 40 s on 2026-08-19
- **Decision authority**: on-call may roll back unilaterally on any halt trigger; no approval meeting required
- **Abort handling**: yank tags, mark registry version `broken`, file abort note with reason within 1 hour

## 6. Verification Evidence
- [ ] Signed tag verification log attached
- [ ] Checksum verification log from pristine VM attached
- [ ] Migration rehearsal logs both directions with row counts attached
- [ ] Rollout promotion timestamps and production smoke log attached
- [ ] `scripts/verify.sh` exits 0 (log attached)
