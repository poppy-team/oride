# Code Review Security Guide

Security code reviews aim to identify vulnerabilities before code is merged and deployed. Reviewers should adopt an attacker's mindset.

## Key Focus Areas

### 1. Data Flow Analysis
- Trace user input from entry points (API endpoints, UI forms) to sinks (databases, file systems, OS commands).
- Ensure validation occurs immediately upon entry.
- Ensure sanitization/encoding occurs immediately before the sink.

### 2. Authentication and Session Management
- Look for custom authentication logic. Custom logic is often flawed; prefer established libraries.
- Check session expiration, revocation, and token handling (e.g., JWT signing algorithms, secret management).
- Ensure `Secure` and `HttpOnly` flags are used for session cookies.

### 3. Access Control (Authorization)
- Verify that every endpoint that modifies state or accesses sensitive data checks the user's permissions.
- Look out for Insecure Direct Object References (IDOR). E.g., `GET /user/123/profile` — Does the backend check if the logged-in user *is* user 123 or an admin?

### 4. Cryptography and Secrets
- Ensure no API keys, passwords, or tokens are hardcoded. Look for strings like `password=`, `secret=`, `api_key=`.
- Verify the use of strong cryptographic algorithms. Look for obsolete algorithms like MD5, SHA1, DES.

### 5. Error Handling and Logging
- Ensure `try...catch` blocks do not silently fail and hide security events.
- Check that stack traces or detailed database errors are not returned in HTTP responses.
- Verify that sensitive data (PII, passwords, tokens) is not logged.

### 6. Business Logic Flaws
- Does the code enforce business rules correctly? (e.g., can a user transfer a negative amount of money? Can a user buy an item for $0?)
- Look for race conditions in critical state changes (e.g., concurrent withdrawals).
