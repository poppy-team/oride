---
name: testing-quality
description: Testing pyramid, deterministic tests, regression tests, behavior assertions, test isolation, flaky detection, representative fixtures
---
# Testing Quality Engineering

## 1. Testing Pyramid
Structure the test suite according to the Testing Pyramid: a broad base of fast, isolated unit tests, a smaller middle layer of integration tests, and a narrow peak of slow, end-to-end (E2E) tests. Push tests down the pyramid wherever possible.

## 2. Deterministic Execution
Tests must be perfectly deterministic. A test that passes sometimes and fails others is worse than no test. Eliminate sources of non-determinism: network calls, uncontrolled timers, random number generators, and shared mutable state.

## 3. Regression Test Coverage
Ensure every discovered bug is accompanied by a regression test before fixing. The test must fail before the fix and pass after. This builds an unbreakable safety net and prevents the same bug from reappearing.

## 4. Behavior Assertions
Assert against observable behavior, not internal implementation details. Test the public API of a component. Avoid mocking internal methods or checking private state, as this leads to fragile tests that break during refactoring.

## 5. Strict Test Isolation
Each test must run in complete isolation. Tests must not depend on the execution order or the side effects of other tests. Reset the database, clear caches, and recreate objects before each test run.

## 6. Flaky Test Detection
Implement automated mechanisms to detect and quarantine flaky tests. Re-run failed tests locally or in CI to identify non-determinism. Treat a flaky test as a failing test and prioritize its repair or removal.

## 7. Representative Fixtures
Use realistic, representative data for test fixtures. Avoid overly simplistic 'foo/bar' data if domain-specific constraints exist. Use factories or builders to generate valid, complex objects easily without cluttering the test setup.

## 8. Mutation Testing
Use mutation testing tools to evaluate the quality of your test suite. These tools introduce small defects (mutations) into your code and check if your tests catch them. High coverage doesn't guarantee good assertions; mutation testing does.

## 9. Test Driven Development (TDD)
Encourage TDD as a design tool. Writing tests first forces you to think about the API from the consumer's perspective, leading to simpler, more decoupled, and more testable code architectures.

## 10. Performance Testing
Integrate performance tests into the CI pipeline to catch regressions. Assert on response times, memory usage, and CPU load under simulated load. Establish baselines and fail the build if performance degrades beyond acceptable thresholds.

