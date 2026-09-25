# Secure Coding — Verification Checklist

## 1. Input Validation & Injection Defense
- [ ] All external inputs validated against strict allow-lists (type, length, format, range).
- [ ] Parameterized queries / prepared statements used exclusively for database access.
- [ ] No shell-constructed commands from user data; argument vectors used where processes spawn.

## 2. Output Encoding & Data Protection
- [ ] Context-aware output encoding applied before rendering (HTML/JS/URL contexts distinguished).
- [ ] Sensitive data encrypted in transit (TLS 1.2+) and at rest (AES-256 or equivalent); keys from KMS/Vault.
- [ ] No sensitive data in logs, errors, or client-visible messages (generic user-facing errors).

## 3. Memory & Type Safety
- [ ] Buffer boundaries enforced; unsafe functions (`strcpy`, `sprintf`, `gets`) absent in C/C++ paths.
- [ ] Integer arithmetic on sizes/indices checked for overflow or built with safe-math primitives.
- [ ] Format strings static; user data passed only as arguments, never as the format parameter.

## 4. Authentication, Session & Access Control
- [ ] Passwords hashed with Argon2/bcrypt; MFA available; no custom crypto.
- [ ] Session IDs rotated on login, invalidated on logout/timeout; cookies carry `HttpOnly`, `Secure`, `SameSite`.
- [ ] Authorization checked on every state-changing endpoint; default-deny on failure.

## 5. Verification & Evidence
- [ ] SAST run clean on touched code; findings triaged with suppressions justified or removed.
- [ ] Security regression tests added for fixed classes of bug.
- [ ] `scripts/verify.sh` exits 0 from the repository root.
