# Dependency Management Audit — API Gateway (Q3)

## 1. Metadata
- **Skill**: dependency-management (Dependency Management)
- **Date**: 2026-09-10
- **Author / Agent**: supply-chain-bot
- **Target Goal / Phase**: GOAL-087 q3-dependency-audit

## 2. Executive Summary
Audited the API gateway's 214 direct + transitive dependencies (npm). Pinned 3 floating ranges, removed 2 abandoned packages (`leftpad-slim`, `http-unwrap`), upgraded `fastify` 4.21 → 4.28 to close CVE-2026-4182 (high), and allowlisted licenses (MIT/Apache-2.0/ISC/BSD). Advisory scan is clean; lockfile reproducible across 3 CI runs.

## 3. Inputs & Scope
- **Inputs Evaluated**: `package.json`, `package-lock.json` (lockfileVersion 3), GHSA + OSV advisories as of 2026-09-10
- **Artifacts Modified**: `package.json` (3 pins, 2 removals, 1 upgrade), `LICENSES.md` (regenerated), `.github/dependabot.yml` (weekly cadence)

## 4. Key Findings & Implementation Details
- **Vulnerabilities**: 1 high (`fastify` CVE-2026-4182, fixed by upgrade), 0 medium, 0 low remaining. SLA met (high patched within 48 h of disclosure).
- **Surface reduction**: Replaced `leftpad-slim` with 12-line internal util; replaced `http-unwrap` with native `fetch` wrapper. Direct deps 38 → 36.
- **Transitive risk**: Deepest chain 7 levels (`gateway → auth → jwks → cache → ...`); flagged `expired-cache@1.2.0` (unmaintained 3 y) for Q4 migration to `lru-map`.
- **Pinning**: All ranges converted to exact versions; lockfile hash `sha512:9f2c…` reproduced on linux/mac/windows runners.
- **Licenses**: 214/214 allowlisted; 0 GPL/AGPL. `LICENSES.md` committed as build artifact check.

## 5. Verification & Evidence
- **Evidence Type**: security-scan, build
- **Test Results**: Passed — `npm audit` 0 vulnerabilities; `license-checker` 0 violations; full test suite 312/312 green post-upgrade
- **Static Analysis Status**: Pass — build reproducible; no floating ranges (`npm ls` exact)

## 6. Next Steps & Handoff
- Migrate `expired-cache@1.2.0` → `lru-map` in Q4 (tracked as GOAL-091); owner: backend-team.
- Weekly Dependabot PRs auto-merge on green CI; owner: supply-chain-bot.
