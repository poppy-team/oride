# Secrets Management Best Practices

Hardcoded secrets (API keys, database passwords, TLS certificates) are one of the most common and severe security vulnerabilities in modern applications.

## Key Principles

### 1. Never Hardcode Secrets in Source Code
- Do not store secrets in `.py`, `.js`, `.java`, or configuration files (`config.json`, `application.yml`) that are committed to version control.
- Even if a repository is private, hardcoded secrets are a massive risk (insider threat, accidental open-sourcing, compromised developer machine).

### 2. Do Not Commit `.env` Files
- Environment variables are the standard way to pass configuration to 12-factor apps.
- However, `.env` files should be added to `.gitignore`.
- Provide a `.env.example` with dummy values instead.

### 3. Use a Dedicated Secrets Manager
- For production, use a purpose-built KMS/Secrets Manager:
  - AWS Secrets Manager / AWS Systems Manager Parameter Store
  - HashiCorp Vault
  - Azure Key Vault
  - Google Cloud Secret Manager
- Applications should authenticate to the Secrets Manager (using IAM roles, Kubernetes Service Accounts, etc.) and retrieve the secrets at runtime into memory.

### 4. Scan for Secrets in CI/CD
- Use tools like `git-secrets`, `trufflehog`, or `gitleaks` as pre-commit hooks and in your CI pipeline.
- If a secret is committed, the build must fail immediately.

### 5. Rotate Secrets Regularly
- If a secret is compromised (or even suspected to be), it must be rotated immediately.
- Implement automated rotation policies for database passwords and API keys where the provider supports it.
- **Revoke** the old secret; do not just generate a new one.

### 6. Principle of Least Privilege for Secrets
- A web service should only have access to the specific secrets it needs.
- Do not use a single "god" database password for all microservices. Create separate database users with restricted permissions.
