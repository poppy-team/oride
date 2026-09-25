# Dependency and Supply Chain Security

Modern applications are built heavily on open-source libraries. If a dependency is compromised, your application is compromised.

## Key Risks
- **Vulnerable Dependencies**: Using components with known vulnerabilities (CVEs).
- **Typosquatting**: Attackers publish malicious packages with names similar to popular ones (e.g., `reqeusts` instead of `requests`).
- **Dependency Confusion**: Attackers publish a malicious package to a public registry with the same name as a private, internal package.
- **Compromised Maintainers**: An attacker steals a maintainer's credentials and publishes a malicious update to a legitimate package.

## Mitigation Strategies

### 1. Software Composition Analysis (SCA)
Use automated tools in your CI/CD pipeline to scan dependencies for known vulnerabilities and licensing issues.
- Node.js: `npm audit`, `Snyk`
- Python: `pip-audit`, `safety`
- Java/Maven: `OWASP Dependency-Check`
- General: `Dependabot`, `Renovate`

### 2. Lockfiles and Pinning
- Always commit lockfiles (`package-lock.json`, `Pipfile.lock`, `yarn.lock`, `Cargo.lock`).
- Lockfiles ensure reproducible builds by cryptographically hashing the exact versions of dependencies (and their sub-dependencies) that were installed.
- Do not use loose versioning (e.g., `^1.2.0` or `latest`) in production deployments without a lockfile.

### 3. Software Bill of Materials (SBOM)
- Generate an SBOM (in formats like CycloneDX or SPDX) for every release.
- An SBOM is a comprehensive inventory of all components, libraries, and modules required to build and run the software. It is crucial for rapid response when a new 0-day vulnerability is announced (e.g., Log4Shell).

### 4. Dependency Proxies
- Use a private package repository or proxy (e.g., Nexus, Artifactory) to cache external dependencies.
- This protects against public registry outages and left-pad incidents (where a package is deleted by the author).
- Configure the proxy to block known malicious packages.
