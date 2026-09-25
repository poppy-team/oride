# Security Review Audit Report

## Scope
- Component: pull request #842, `auth/session-cookie`
- Assessor: independent-security-reviewer
- Review date: 2026-09-23
- Risk: High — authentication boundary changed

## Review Method
- Compared the diff with the session threat model and checked every cookie path.
- Ran SAST, dependency checks, and focused integration tests.
- Traced token issuance, rotation, logout, and failure responses.

## Findings
- **HIGH**: The proposed cookie lacked `Secure` on the staging callback; blocked until the attribute is present in production and staging.
- **MEDIUM**: Logout did not revoke the refresh-token family; a new revocation integration test is required.
- **INFO**: The nonce rotation is correctly generated per request and is not logged.

## Verdict
Request changes. Re-review after the two blocking findings receive failing-first regression tests and the exact job is green. No security escape hatch or unregistered suppression is accepted.
