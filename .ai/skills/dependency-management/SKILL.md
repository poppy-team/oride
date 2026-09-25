---
name: dependency-management
description: Minimal dependency surface, provenance, pinning, advisories, maintenance, transitive risk, licensing
---
# Dependency Management

## 1. Minimal Dependency Surface
Every dependency is a liability. Strictly evaluate the need for a new dependency against building the functionality in-house. Minimize the total number of dependencies to reduce the attack surface, maintenance burden, and build times.

## 2. Provenance Verification
Verify the source and integrity of all dependencies. Use package managers that support cryptographic hashing and signature verification. Ensure dependencies are downloaded from trusted, authoritative registries.

## 3. Strict Dependency Pinning
Pin all dependencies to exact versions (e.g., major.minor.patch) in application projects. Use lockfiles to ensure reproducible builds across all environments. Avoid floating versions or ranges that can introduce unexpected breaking changes.

## 4. Security Advisories Auditing
Continuously monitor dependency trees for known vulnerabilities using automated tools (e.g., npm audit, dependabot, Trivy). Establish a strict SLA for patching critical and high-severity vulnerabilities in dependencies.

## 5. Maintenance and Upgrades
Treat dependency upgrades as regular maintenance, not an afterthought. Schedule regular cycles to update dependencies to their latest stable versions. Test upgrades thoroughly to catch subtle breaking changes or behavioral shifts.

## 6. Transitive Risk Assessment
Understand that you depend not just on your direct dependencies, but on their dependencies as well (transitive dependencies). Analyze the deep dependency graph to identify risky, unmaintained, or bloated packages brought in indirectly.

## 7. Licensing Compliance
Automate the checking of dependency licenses. Ensure all transitive dependencies have licenses compatible with your project's distribution model. Maintain an explicit allowlist of acceptable licenses and block builds containing incompatible ones.

## 8. Vendor Branching / Vendoring
For critical systems, consider vendoring dependencies directly into the source control repository. This guarantees availability, prevents upstream removal (left-pad scenario), and allows for targeted local patching if upstream is unresponsive.

## 9. Deprecation Management
Actively track the deprecation status of dependencies. Plan migrations away from abandoned or deprecated packages before they become security risks or block upgrades of other components.

## 10. Private Registries
Utilize private package registries or proxies for internal code sharing and to cache external dependencies. This provides a buffer against external registry outages and allows for tighter control over approved package versions.

