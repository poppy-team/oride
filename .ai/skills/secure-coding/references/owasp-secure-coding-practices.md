# OWASP Secure Coding Practices Quick Reference

This guide outlines fundamental secure coding practices that should be integrated into the development lifecycle.

## 1. Input Validation
- **Rule**: Never trust user input.
- **Practice**: Validate all input against a rigorous allowlist (positive validation). Check length, type, syntax, and business rules.
- **Avoid**: Relying solely on deny-lists (negative validation).

## 2. Output Encoding
- **Rule**: Encode data before rendering it in a browser or passing it to an interpreter.
- **Practice**: Use context-aware output encoding (e.g., HTML entity encoding, JavaScript encoding, URL encoding) to prevent Cross-Site Scripting (XSS).

## 3. Parameterized Queries
- **Rule**: Separate code from data in database queries.
- **Practice**: Always use parameterized queries or prepared statements for SQL/NoSQL databases to prevent Injection flaws.

## 4. Authentication and Password Management
- **Rule**: Implement robust authentication mechanisms.
- **Practice**: Use strong password hashing algorithms (e.g., Argon2, bcrypt). Implement MFA. Enforce password complexity and expiration policies reasonably. Do not roll your own crypto.

## 5. Session Management
- **Rule**: Protect session identifiers.
- **Practice**: Generate new session identifiers on login. Invalidate sessions on logout or timeout. Set `HttpOnly`, `Secure`, and `SameSite` flags on session cookies.

## 6. Access Control
- **Rule**: Enforce authorization checks on every request.
- **Practice**: Implement the principle of least privilege. Use role-based (RBAC) or attribute-based (ABAC) access controls. Fail securely (default deny).

## 7. Cryptographic Practices
- **Rule**: Protect sensitive data at rest and in transit.
- **Practice**: Use TLS 1.2+ for all data in transit. Encrypt sensitive data at rest using strong, industry-standard algorithms (e.g., AES-256). Store keys securely (e.g., KMS, Vault).

## 8. Error Handling and Logging
- **Rule**: Log security events without exposing sensitive data.
- **Practice**: Log authentication successes/failures, access control failures, and exceptions. Do not log passwords, session IDs, or PII. Provide generic error messages to users.
