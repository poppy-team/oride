# Pragmatic Clean Code Verification Checklist

## 1. Domain Naming & Intent
- [ ] Variables are descriptive nouns; functions are descriptive action verbs.
- [ ] Booleans use positive predicate prefixes (`is_`, `has_`, `can_`, `should_`).
- [ ] No cryptic single-letter names outside standard loop indices (`i`) or coordinates (`x`, `y`).
- [ ] Avoid generic noise words (`Data`, `Info`, `Manager`, `Processor`, `Object`).

## 2. Function Granularity & Responsibilities
- [ ] Every function adheres to Single Level of Abstraction Principle (SLAP).
- [ ] Functions fit comfortably on a single screen (<= 35 lines).
- [ ] Parameter arity is <= 3 arguments (otherwise bundled into a cohesive parameter struct).
- [ ] Guard clauses and early returns are used to eliminate deep nesting.

## 3. Anti-Premature Abstraction & Architecture
- [ ] Rule of Three respected: No abstraction created for only two occurrences of similar code.
- [ ] No single-implementation interfaces (concrete structs used until polymorphism is required).
- [ ] Composition favored over deep class inheritance hierarchies.
- [ ] Pure calculation functions isolated from I/O and side effects (Command-Query Separation).

## 4. Smells & Dead Code
- [ ] Primitive Obsession replaced with domain Value Objects where validation or invariants exist.
- [ ] Zero commented-out dead code blocks.
- [ ] Cyclomatic complexity of functions is <= 10.
- [ ] All unit and regression tests pass deterministically before and after refactoring.
