---
name: clean-code
description: Pragmatic Clean Code engineering, domain-expressive naming, function single-responsibility, cohesion & coupling metrics, language-specific idiomatic design, and elimination of premature abstraction.
---

# Pragmatic Clean Code & Design Ergonomics

## 1. Title and Description
**Pragmatic Clean Code & Design Ergonomics (`clean-code`)**
Enforces the Prumo canonical clean code discipline: explicit responsibilities, low coupling, high cohesion, domain-driven naming, small cohesive functions, composition over inheritance, deterministic error returns, and zero speculative abstractions.

## 2. Purpose
Maintain a codebase that is immediately understandable, easily testable, and cheap to modify, while strictly guarding against both "spaghetti code" and its opposite hazard: "enterprise astronautics" (premature interfaces, excessive wrapper layers, and over-engineering).

## 3. Prerequisites
- Target source code in any language supported by Prumo (Go, Rust, TypeScript, Python, C++, Zig).
- Formatting and static analysis tools for the target stack.

## 4. Inputs
- Existing source files, pull request diffs, or newly authored modules.
- Domain models and architectural boundary definitions.

## 5. Outputs
- Clean, readable, idiomatic code adhering to language-specific conventions.
- Explicit refactoring diffs removing code smells and dead code.
- Clean code review findings conforming to `templates/clean-code-review.md`.

## 6. Execution Steps

### Step 1: Domain-Expressive Naming
1. **Intention-Revealing Names**: Names must state why an entity exists, what it does, and how it is used. Avoid noise words (`Data`, `Info`, `Manager`, `Processor`, `Object`) unless formally defined in the domain glossary.
2. **Grammar Conventions**:
   - Variables/Types: Nouns or noun phrases (`Invoice`, `TransactionBatch`, `TokenEnvelope`).
   - Functions/Methods: Action verbs or predicates (`calculateTax()`, `isAuthorized()`, `parsePayload()`).
   - Booleans: Formulate as yes/no predicates (`is_valid`, `has_children`, `can_retry`). Never use negative names (`is_not_empty`).
3. **Searchability and Pronounceability**: Forbid single-letter variables except for standard loop indices (`i`, `j`), coordinate mathematics (`x`, `y`), or standard idiom conventions (Go receiver `s *Server`).

### Step 2: Function Responsibility & Granularity
1. **Single Level of Abstraction Principle (SLAP)**: Statements within a function should belong to the same level of conceptual abstraction. High-level orchestrators should not mix with low-level bitwise masking or string manipulation.
2. **Small & Cohesive**: Limit function size to what comfortably fits on a single viewport (typically <= 35 lines). If a function requires comments dividing it into logical "phases", extract each phase into a private helper function.
3. **Minimize Argument Arity**:
   - 0–2 arguments: Ideal.
   - 3 arguments: Acceptable when strictly related.
   - 4+ arguments: Code smell. Bundle parameters into a cohesive configuration struct or options object.

### Step 3: Anti-Premature-Abstraction Rule (Pragmatic Discipline)
1. **The Rule of Three**: Do not extract a shared abstraction upon the second occurrence of code. Wait until the pattern appears a third time with demonstrably identical *behavioral* intent (not just coincidental structural similarity).
2. **No Single-Implementation Interfaces**: In Go, Rust, and TypeScript, never declare an interface if there is only one concrete implementation, unless explicitly required for external mock boundaries in unit tests.
3. **Composition Over Inheritance**: Avoid deep class inheritance hierarchies. Favor composing flat structs containing discrete capability traits or delegates.

### Step 4: Cohesion & Coupling Optimization
1. **High Cohesion**: Every method in a struct/class should access or modify a significant portion of its fields. If a subset of methods only touches a subset of fields, split the struct into two distinct types (Lack of Cohesion of Methods - LCOM).
2. **Low Coupling**: Depend on abstract, stable capabilities rather than volatile concrete modules. Dependencies must point inwards toward domain logic.
3. **No Hidden Temporal Coupling**: If method `B()` must strictly be called after method `A()`, enforce this through the type system (e.g., `A()` returns a `ConfiguredState` token required as a parameter by `B()`).

### Step 5: Side Effects & State Mutation
1. **Pure Core, Imperative Shell**: Isolate pure calculation functions (no I/O, no mutation, deterministic output from input) from the imperative edges that handle filesystem, database, and network I/O.
2. **Command-Query Separation (CQS)**: A function should either perform an action (mutation) or answer a query (return data), but never both, unless returning state transition receipts (e.g. `Pop()`).
3. **Immutability by Default**: Mark variables, parameters, and fields immutable wherever the language supports it (`const`, `readonly`, `let` without `mut`).

### Step 6: Code Smell Audit & Elimination
Audit the codebase systematically for classic Martin Fowler code smells:
- **Long Method / Large Class**: Extract Class / Extract Method.
- **Primitive Obsession**: Replace raw primitives with typed Value Objects (e.g. `UserId`, `Money`, `EmailAddress`).
- **Feature Envy**: Method frequently accesses data of another class; move the method to the data owner.
- **Dead Code**: Delete commented-out code and unreferenced functions immediately. Version control preserves history.

## 7. Verification
```bash
# 1. Complexity & Cohesion Analysis
./scripts/analyze_complexity.sh .

# 2. Linter & Static Analysis (Language-specific)
# Go: golangci-lint run
# Rust: cargo clippy -- -D clippy::pedantic
# TypeScript: eslint .
```

## 8. Fallbacks & Failure Recovery
| Code Smell / Hazard | Root Cause | Deterministic Remediation |
|---|---|---|
| Cyclomatic Complexity > 15 | Nested conditionals, large switch statements | Apply "Replace Conditional with Polymorphism" or early-return guards (`guard clauses`). |
| Shotgun Surgery | Single feature change forces edits across 6+ distinct files | Group related responsibilities into a single cohesive domain module. |
| Divergent Change | One class is repeatedly changed for multiple unrelated reasons | Apply Single Responsibility Principle: separate concerns into dedicated structs. |
| Premature Generic Boilerplate | Interfaces with only 1 implementation, deep factories | Inline factories and replace interface with concrete struct until polymorphism is proven necessary. |

## 9. Constraints
- **Preserve behavior**: Refactoring must never alter external observable behavior. All unit tests must pass before and after changes.
- **Do not invent abstractions**: Never create generic plugins, base classes, or factories "just in case" future requirements need them.
- **No commented-out code**: Dead code must be removed, not commented out.

## 10. Examples

### Anti-Pattern: Primitive Obsession & Premature Complexity
```typescript
// BAD: Primitives, temporal coupling, side effects, cryptic naming
class OrderProc {
    private db: any;
    public o: any;

    // Temporal coupling: must call init() before calc()!
    init(orderId: string) {
        this.o = this.db.fetch(orderId);
    }

    calc(d: number, t: string) {
        // Deep nested logic, magic numbers, side-effect inside query
        if (this.o != null) {
            if (t === "US") {
                this.o.total = this.o.subtotal * 1.08 - d;
                this.db.save(this.o); // Unexpected side-effect in calculate!
                return this.o.total;
            }
        }
        return -1; // Magic error code!
    }
}
```

### Idiomatic Pattern: Domain Types, Immutability & CQS
```typescript
// GOOD: Value objects, pure calculations, explicit errors, clear naming
export class CurrencyAmount {
    constructor(public readonly cents: number, public readonly currency: "USD" | "EUR") {
        if (!Number.isInteger(cents) || cents < 0) {
            throw new Error(`Invalid monetary amount: ${cents}`);
        }
    }
}

export interface TaxPolicy {
    calculateTax(subtotal: CurrencyAmount): CurrencyAmount;
}

export class OrderBillingService {
    // Pure calculation function: zero I/O, deterministic, easily testable
    public calculateFinalTotal(
        subtotal: CurrencyAmount,
        discount: CurrencyAmount,
        taxPolicy: TaxPolicy
    ): CurrencyAmount {
        if (discount.cents > subtotal.cents) {
            throw new Error("Discount cannot exceed order subtotal");
        }

        const discountedSubtotal = new CurrencyAmount(subtotal.cents - discount.cents, subtotal.currency);
        const tax = taxPolicy.calculateTax(discountedSubtotal);

        return new CurrencyAmount(discountedSubtotal.cents + tax.cents, subtotal.currency);
    }
}
```
