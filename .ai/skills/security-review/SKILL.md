---
name: security-review
description: Threat model alignment, security diff analysis, SAST integration, independent verifier policy, escape hatch denial, and audit logging
---
# Security-Focused Code Review

## 1. Threat Model Alignment
Verify every pull request against the system threat model and trust boundaries. Evaluate whether proposed changes introduce new attack vectors, expand attack surfaces, or alter privilege boundaries.

## 2. Security-Focused Diff Prioritization
Prioritize scrutiny on code modifications touching authentication, authorization, cryptographic routines, session handling, input deserialization, and network egress.

## 3. Automated SAST & Security Linter Gates
Require 100% clean passes from static application security testing (SAST) tools (e.g. CodeQL, Semgrep, gosec) as mandatory pre-conditions for review approval.

## 4. Independent Verifier Policy
Enforce the Prumo Independent Verifier policy: any change categorized with High or Critical risk must be reviewed by a distinct, specialized security reviewer or persona.

## 5. Security Acceptance Criteria Verification
Verify that security-related pull requests include dedicated automated regression test cases demonstrating that the identified vulnerability is eliminated.

## 6. Denial of Unregistered Escape Hatches
Strictly reject pull requests introducing banned unsafe constructs (raw pointers, @trusted, memory reinterpret casts, or linter suppression directives) without formal registration in .prumo/escape-hatches.json.

## 7. Defense-in-Depth Verification
Ensure that security architecture does not rely on a single defensive layer. Verify multiple interlocking safeguards (e.g. database-level constraints combined with application-level authorization).

## 8. Security Regression Prevention
Mandate automated unit or integration tests for every security bug fix, ensuring the test fails against the vulnerable baseline and passes with the fix.

## 9. Structured Review Findings Format
Emit security review findings using standard severity ratings (Critical, High, Medium, Low) with explicit file links, line ranges, vulnerability explanations, and suggested code patches.

## 10. Audit Trail & Compliance Records
Log the final security review verdict, reviewer identity, timestamp, commit hash, and verification gate evidence permanently into the project compliance history.
