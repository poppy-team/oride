---
name: secrets-security
description: Zero hardcoded secrets, automated pre-commit scanning, centralized secret managers, ephemeral tokens, and automated log redaction
---
# Secrets & Credential Protection

## 1. Zero Hardcoded Secrets Invariant
Enforce an absolute ban on plaintext secrets (API keys, private keys, database passwords, webhook signing secrets) in source code, commit history, configuration files, and container images.

## 2. Automated Secret Scanning Gates
Run automated secret detection tools (Gitleaks, TruffleHog) in local git pre-commit hooks and as blocking CI gates. Analyze commit diffs for high-entropy strings and known vendor token patterns.

## 3. Environment Separation & Scoped Injection
Isolate credentials strictly across development, staging, and production environments. Follow Twelve-Factor App principles by injecting secrets via environment variables or mounted tmpfs files at runtime.

## 4. Centralized Secrets Management
Retrieve production secrets dynamically from centralized managers (HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager). Use short-lived, rotatable credentials instead of static long-lived strings.

## 5. Secret Rotation Playbooks
Design application systems to support zero-downtime secret rotation. Support dual-key verification windows during rotation (accepting both old and new signing keys while tokens migrate).

## 6. Ephemeral Developer Tokens
Provide developers with scoped, short-lived development tokens (validity < 30 days) rather than production-grade credentials. Enforce immediate revocation upon employee or collaborator offboarding.

## 7. Memory & Core Dump Protection
Overwrite sensitive byte buffers with zeros immediately after use where supported by the language runtime. Disable automatic core dumps in production environments to prevent credential extraction from memory.

## 8. Automated Log & Trace Redaction
Configure log formatters and OpenTelemetry exporters to automatically scrub patterns resembling JWTs, Bearer tokens, private keys, and passwords before persisting records to disk or central logging systems.

## 9. Compromise Response & History Scrubbing
In the event of an accidental secret commit, treat the secret as compromised immediately and revoke it at the provider first. Only after revocation, rewrite git history using git-filter-repo to purge the artifact.

## 10. Supply Chain & CI Secret Isolation
Use GitHub Actions encrypted repository secrets or OIDC Workload Identity Federation instead of long-lived cloud access keys. Restrict secret visibility to protected branches and prevent pull requests from forks from accessing secrets.
