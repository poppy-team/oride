---
name: secure-coding
description: Input validation, output encoding, buffer safety, integer overflow, format strings
---
# Secure Coding Practices

## 1. Input Validation
Implement strict allow-list validation for all external input. Validate data type, length, format, and range. Reject any input that does not conform to expected patterns. Never rely solely on client-side validation.

## 2. Output Encoding
Contextually encode all user-controllable data before rendering it in a browser, injecting it into a command, or inserting it into a database query. This is the primary defense against injection attacks like XSS.

## 3. Buffer Safety
When using memory-unsafe languages (C/C++), rigidly control buffer boundaries. Avoid unsafe functions like strcpy or sprintf. Use safe alternatives (strncpy, snprintf) and strictly validate lengths to prevent buffer overflows.

## 4. Integer Overflows
Be vigilant against integer overflows and underflows, which can bypass logic checks or cause memory corruption. Use safe math libraries or compiler flags that trap on overflow, especially when calculating buffer sizes or array indices.

## 5. Format String Vulnerabilities
Never pass user-controlled data directly as the format string parameter to logging or formatting functions (e.g., printf in C, logging in Python). Always use static format strings and pass user data as arguments.

## 6. Parameterized Queries
Prevent SQL Injection by exclusively using parameterized queries or prepared statements. Never concatenate untrusted data into SQL strings, regardless of any escaping mechanisms applied.

## 7. Safe Deserialization
Avoid deserializing complex objects from untrusted sources if possible. If required, use safe serialization formats (JSON, Protobuf) rather than language-native binary formats (Java Serialization, Python Pickle) which can lead to Remote Code Execution.

## 8. Cryptographic Agility
Do not roll your own crypto. Use standard, well-vetted cryptographic libraries. Implement cryptographic agility: design systems so that algorithms and key sizes can be easily upgraded if a vulnerability is discovered.

## 9. Fail Securely
Ensure systems fail into a secure state. If an authorization check fails or throws an exception, the default action must be to deny access. Never fail open.

## 10. Least Privilege Enforcement
Apply the Principle of Least Privilege at the code level. Drop unnecessary privileges immediately after use. Run processes with the minimum permissions required. Restrict file system and network access to only what is strictly necessary.

