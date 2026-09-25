---
name: refactoring
description: Detect smell -> establish behavior -> regression test -> minimal transformation -> verify -> repeat
---
# Refactoring Engineering

## 1. Smell Detection
Proactively identify code smells that indicate deeper design problems. Look for duplicated code, long methods, large classes, long parameter lists, divergent change, and shotgun surgery. Smells are heuristics, not strict rules; they guide you to areas needing investigation.

## 2. Establishing Behavior
Before changing any code, you must deeply understand its current behavior, including edge cases and undocumented quirks. Read the code, consult domain experts, and review version control history. Do not guess; prove the behavior.

## 3. Regression Testing
Ensure a robust suite of fast, deterministic regression tests exists before refactoring. If tests are missing, write characterization tests (golden master tests) that lock down the current output for given inputs, regardless of whether that output is 'correct'.

## 4. Minimal Transformation
Execute refactoring in tiny, safe, verifiable steps. Use proven automated refactoring tools provided by your IDE whenever possible (e.g., Extract Method, Rename, Inline). Avoid manual text editing for structural changes to minimize human error.

## 5. Verification
Run the test suite immediately after every small transformation. If a test fails, you have taken too large a step or made a mistake. Undo (revert) immediately, break the step down into smaller pieces, and try again. Never commit broken code.

## 6. Iterative Process
Refactoring is not a single large event; it is an iterative loop: smell -> test -> refactor -> verify -> commit. Repeat this loop constantly. Keep refactoring commits separate from feature commits to ensure bisectability and clear history.

## 7. Extract and Override
When dealing with legacy code, use the Extract and Override Call pattern to break dependencies. Extract the problematic dependency instantiation into a protected factory method, then subclass and override it in your test to inject a mock or fake.

## 8. Strangler Fig Application
For large-scale structural refactoring or rewriting, use the Strangler Fig pattern. Build the new system around the edges of the old system. Intercept calls, route them to the new system where implemented, and fall back to the old system otherwise. Gradually expand the new system.

## 9. Parallel Change
When changing APIs or database schemas, use the Parallel Change (Expand and Contract) pattern. 1. Expand: Add the new element alongside the old. 2. Migrate: Update clients to use the new element. 3. Contract: Remove the old element. This ensures backward compatibility during the transition.

## 10. Feature Flags for Refactoring
Use feature flags to hide large-scale refactorings in production until they are fully verified. Run both the old and new code paths concurrently in a 'dark launch' mode, compare their results, log discrepancies, and switch over only when confidence is high.

