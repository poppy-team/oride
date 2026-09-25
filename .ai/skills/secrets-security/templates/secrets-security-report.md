# Secrets & Credential Security Audit Report

## Scope
- Component: `billing-api` production configuration
- Assessor: secrets-review-agent
- Review date: 2026-09-23
- Evidence: repository scan, CI policy, and redaction test output

## Exposure Review
- Source tree: no plaintext keys or private keys found.
- CI: OIDC workload identity replaces the legacy `AWS_SECRET_ACCESS_KEY` secret.
- Runtime: `prod/Billing/Database` is read at startup through IAM and is not logged.
- Development: temporary token expires after 7 days and is scoped to the sandbox account.

## Findings
- **HIGH**: A staging token was present in an old CI log; it was revoked before the report was archived.
- **MEDIUM**: One error formatter exposed a password field when the upstream service returned a validation error.
- **LOW**: The rotation ledger lacked an owner for the payment webhook key.

## Remediation
- Added redaction for `password`, `token`, `authorization`, and private-key headers.
- Replaced long-lived cloud credentials with the GitHub OIDC role.
- Assigned rotation owners and dates to all 11 production secrets.
