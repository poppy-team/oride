# Supply Chain & Dependency Security Checklist

- [ ] Lockfiles committed and verified with cryptographic checksums in CI
- [ ] Production dependency versions strictly pinned; wildcards banned
- [ ] Automated vulnerability scanning (govulncheck/audit) blocks CI on High/Critical CVEs
- [ ] Builds run with script execution disabled (--ignore-scripts) where applicable
- [ ] Dependency graph audited for unnecessary transitive packages
- [ ] Licenses scanned automatically to prevent legal contamination
- [ ] Emergency rollback and lockfile freeze playbook documented
