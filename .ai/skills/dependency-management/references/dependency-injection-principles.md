# Dependency Injection Principles

## 1. Inversion of Control (IoC)
- High-level modules should not depend on low-level modules. Both should depend on abstractions (e.g., interfaces).
- Abstractions should not depend on details. Details (concrete implementations) should depend on abstractions.

## 2. Dependency Injection (DI)
- Dependencies are provided to the dependent object, rather than the object creating them internally.
- Methods of DI:
  - **Constructor Injection** (Preferred): Dependencies are passed through the class constructor. Ensures the object is always in a valid state.
  - **Property Injection**: Dependencies are set via public properties. Useful for optional dependencies.
  - **Method Injection**: Dependencies are passed to specific methods when needed.

## 3. The Composition Root
- A unique location in an application where modules are composed together.
- This is typically done at the application startup (e.g., `main.py`, application bootstrap).
- Only the Composition Root should reference the DI container directly. Avoid the Service Locator anti-pattern where business classes fetch dependencies from a global container.

## 4. Lifecycle Management
- **Transient**: A new instance is created every time the dependency is requested.
- **Singleton**: A single instance is created and shared throughout the application's lifetime.
- **Scoped**: An instance is created per request or context (e.g., per HTTP request in a web framework).
