---
name: architecture-quality
description: Dependency direction, cycle detection, boundary leakage, transport/domain coupling, god modules, shared mutable state, hidden globals, unstable interfaces
---
# Architecture Quality

## 1. Dependency Direction
Dependencies must point inwards toward higher-level policies. The domain layer should not depend on infrastructure or UI layers. This Dependency Inversion Principle ensures that core business logic remains isolated from volatile implementation details.

## 2. Cycle Detection
Cyclic dependencies create a single tightly coupled super-module. Detect and break dependency cycles using Dependency Inversion, extracting common abstractions, or merging modules if they fundamentally represent the same concept. Use static analysis tools to prevent cycles.

## 3. Boundary Leakage
Ensure that domain models do not leak into transport or persistence layers, and vice versa. Use DTOs (Data Transfer Objects) and mappers at system boundaries. A change in the database schema should not necessitate a change in the core domain logic.

## 4. Transport/Domain Coupling
Decouple HTTP/gRPC handlers from domain services. The transport layer is responsible for decoding requests and encoding responses, nothing else. The domain layer should be completely unaware of the transport mechanism used to invoke it.

## 5. God Modules Mitigation
Identify and decompose God Modules—classes or packages that know too much or do too much. Break them apart along business capability lines. Establish clear, narrow interfaces for the resulting smaller modules to interact with each other.

## 6. Shared Mutable State
Avoid shared mutable state across architectural boundaries. When components must share data, prefer message passing, event sourcing, or immutable shared memory. If shared mutable state is unavoidable, encapsulate it strictly behind a synchronizing boundary.

## 7. Hidden Globals Management
Identify and eliminate Hidden Globals (e.g., Singletons that maintain state, implicit context objects). They make unit testing difficult and introduce hidden coupling. Replace them with explicit dependency injection and scoped contexts.

## 8. Unstable Interfaces
Identify interfaces that change frequently and stabilize them. Apply the Stable Dependencies Principle: depend in the direction of stability. Isolate unstable, experimental modules behind stable adapter layers to protect the rest of the system.

## 9. Architectural Fitness Functions
Implement automated architectural fitness functions using tools like ArchUnit. Assert that specific packages do not depend on forbidden packages. Fail the build if architectural constraints (e.g., layer isolation) are violated.

## 10. System Observability Hooks
Design the architecture to be inherently observable. Embed correlation IDs at the system edges and propagate them through all internal boundaries. Define clear, consistent logging contexts and tracing spans at every architectural transition point.

