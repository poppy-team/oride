---
name: supply-chain-security
description: Cryptographic lockfiles, exact dependency pinning, automated vulnerability scanning, package provenance, and typosquatting defense
---
# Supply Chain & Dependency Security

## 1. Cryptographic Lockfile Verification
Commit lockfiles (go.sum, package-lock.json, Cargo.lock, prumo.lock) to version control. Validate cryptographic package hashes during CI builds to guarantee build reproducibility and tamper detection.

## 2. Exact Version Pinning
Pin exact dependency versions in production manifests. Prohibit dynamic version wildcards (*, ^, ~, latest) in deployable services to prevent unexpected upstream breaking changes or malicious minor updates.

## 3. Automated Vulnerability Scanning
Integrate automated CVE scanners (govulncheck, npm audit, cargo audit, pip-audit) as blocking pre-merge checks in CI pipelines. Fail builds immediately upon detection of High or Critical vulnerabilities.

## 4. Package Provenance & SLSA Verification
Verify cryptographic signatures and SLSA build provenance for external packages, libraries, and container base images using Sigstore and Cosign before deployment.

## 5. Typosquatting & Namespace Defense
Enforce strict registry configurations using scoped private namespaces (@company/package). Inspect new dependencies for suspiciously recent publication dates or typo-similarity to popular packages.

## 6. Dependency Minimization & Tree Pruning
Routinely audit the dependency graph for transitive bloat. Prefer well-tested standard library capabilities over importing single-function third-party micro-packages.

## 7. Safe Installation Flags
Run package managers with script-execution disabled during build (npm ci --ignore-scripts) to prevent malicious install hooks from executing arbitrary shell commands during dependency fetching.

## 8. Vendoring for Critical Systems
For security-critical or mission-critical projects, vendor third-party source dependencies into the repository, ensuring complete autonomy from external registry downtime or upstream repository deletion.

## 9. License Compliance Governance
Scan dependency licenses automatically using tools like go-licenses or FOSSA. Reject packages with incompatible or viral licenses (e.g. GPL in proprietary closed-source distributions).

## 10. Upstream Compromise Incident Response
Maintain an emergency playbook for upstream dependency compromise: steps for immediately freezing lockfiles, revoking affected tokens, and publishing patched releases.
