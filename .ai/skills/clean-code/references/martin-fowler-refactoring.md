# Martin Fowler's Refactoring Catalog & Core Transformations

## 1. The Refactoring Loop
Refactoring is a disciplined technique for restructuring an existing body of code, altering its internal structure without changing its external behavior.
```
Identify Smell -> Verify Test Coverage -> Apply Small Transformation -> Verify Tests Pass -> Commit
```

## 2. Essential Refactoring Transformations

### 1. Extract Function / Method
- **When**: A code fragment can be grouped together and named according to its purpose.
- **Why**: Enhances readability, eliminates duplicated logic, and simplifies debugging.
- **Rule**: If you have to spend effort looking at a fragment of code and thinking "what is this doing?", extract it into a function named after the *what*.

### 2. Replace Primitive with Object (Value Object)
- **When**: A primitive data item (e.g. string for phone number, integer for cents) needs special validation, formatting, or behavior.
- **Why**: Eliminates primitive obsession, encapsulates validation in one place, and provides type safety.

### 3. Replace Conditional with Polymorphism / Strategy
- **When**: A `switch` or chain of `if-else` statements checks for type codes or modes across multiple functions.
- **Why**: Eliminates divergent change; adding a new case only requires implementing a new struct/class without editing existing switch statements.

### 4. Introduce Parameter Object
- **When**: A group of parameters frequently travel together through multiple function calls.
- **Why**: Shortens parameter lists, improves naming, and creates a natural home for validation logic.

### 5. Decompose Conditional (Guard Clauses)
- **When**: Deeply nested `if-else` blocks obscure the primary happy path.
- **Why**: Early returns for edge cases and errors allow the main flow to read linearly down the left margin of the editor.
