# Secrets & Credential Protection Checklist

- [ ] Gitleaks/TruffleHog configured in pre-commit and blocking in CI pipeline
- [ ] Zero plaintext credentials or private keys committed in source or configuration
- [ ] Secrets injected via process environment or secure vault, never baked into container images
- [ ] Log formatters configured with automatic redaction filters for tokens and credentials
- [ ] Production credentials separated completely from development/staging environments
- [ ] CI/CD uses OIDC Workload Identity Federation rather than permanent cloud keys
- [ ] Revocation and rotation playbooks tested and verified for all external API keys
