# Supply Chain Security Audit Report

## Scope
- Component: `payments-api` dependency and build pipeline
- Assessor: supply-chain-review-agent
- Review date: 2026-09-23
- Inputs: `package-lock.json`, container base image, and CI release workflow

## Inventory
- 412 production packages are represented in the CycloneDX SBOM.
- `fastify@4.18.2` and `undici@6.19.8` are pinned by integrity hashes.
- Node install scripts run with `--ignore-scripts` in the release job.
- The base image is signed by the organization’s Cosign identity.

## Findings
- **HIGH**: `lodash@4.17.20` was reachable through a transitive path; the lockfile update removes it.
- **MEDIUM**: A direct dependency used a caret range; it is pinned to `4.18.2` until the next reviewed upgrade.
- **LOW**: The SBOM omitted license metadata for two test-only packages; the release workflow now blocks missing fields.

## Remediation
- Freeze the lockfile, regenerate `sbom.cdx.json`, and run the dependency audit.
- Reject High and Critical CVEs in CI and document the emergency rollback owner.
