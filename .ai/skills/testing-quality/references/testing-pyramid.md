# The Testing Pyramid

## 1. Unit Tests (The Base)
- **Scope**: Testing individual functions, classes, or modules in isolation.
- **Characteristics**: Fast, isolated, highly reliable, easy to debug.
- **Volume**: Should make up the vast majority of your test suite.
- **Mocking**: Extensive use of mocks and stubs to isolate the unit from its dependencies (databases, external APIs).

## 2. Integration Tests (The Middle)
- **Scope**: Testing how different parts of the system work together.
- **Characteristics**: Slower than unit tests, may involve real databases (e.g., test DBs), file systems, or network calls to internal services.
- **Volume**: Fewer than unit tests, but still a significant portion.
- **Mocking**: Only mock external systems that you don't own or that are too slow/unreliable to include.

## 3. End-to-End (E2E) Tests (The Top)
- **Scope**: Testing the application as a whole from the user's perspective.
- **Characteristics**: Slow, brittle, hard to debug, testing through the UI or top-level API.
- **Volume**: The smallest number of tests. Focus on critical user journeys.
- **Mocking**: Minimal to no mocking. Real environments where possible.

## Best Practices
- **Write tests that are resilient to refactoring**: Test behavior, not implementation details.
- **Avoid Test Contamination**: Ensure each test starts with a clean slate. Use `setUp`/`tearDown` or Pytest fixtures effectively.
- **Arrange, Act, Assert**: Structure your tests clearly using the AAA pattern.
