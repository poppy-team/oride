# Pragmatic Clean Code Principles & Design Metrics

## 1. Clean Code in Prumo: The Pragmatic Contract
Clean code in Prumo is defined as code that makes intent explicit while incurring minimal cognitive and architectural overhead. It is anchored on five pillars:
1. **Explicit Responsibilities**: Every package, module, struct, and function has one clear reason to change.
2. **Low Coupling**: Modules interact through narrow, stable boundaries without reaching into each other's internal state.
3. **High Cohesion**: Data and the operations that mutate that data live together.
4. **Domain Naming**: Symbols speak the language of the business domain, not incidental infrastructure.
5. **No Speculative Abstraction**: We build for the concrete present, not an imaginary future.

## 2. Software Design Metrics

### Cyclomatic Complexity (McCabe)
- Measures the number of linearly independent paths through program source code.
- Formula: $M = E - N + 2P$ (where $E$ is edges, $N$ is nodes, $P$ is connected components).
- **Target**: Functions should maintain complexity <= 10. Anything > 15 must be refactored into smaller sub-methods or table lookups.

### Lack of Cohesion of Methods (LCOM)
- Measures the dissimilarity of methods in a class/struct based on shared field usage.
- High LCOM indicates that a class is doing multiple unrelated jobs.
- **Action**: Split structs with high LCOM into independent, focused components.

### Coupling Metrics: Afferent vs Efferent
- **Afferent Coupling ($C_a$)**: Number of external classes that depend on this class (responsibility).
- **Efferent Coupling ($C_e$)**: Number of external classes this class depends on (dependency).
- **Instability ($I$)**: $I = \frac{C_e}{C_a + C_e}$. Stable core modules must have $I \to 0$, depending on few volatile things.

## 3. Command-Query Separation (CQS)
Formulated by Bertrand Meyer:
- **Command**: Changes the state of a system but does not return a value (e.g. `ActivateGoal()`).
- **Query**: Returns a value but does not alter observable state (e.g. `GetActiveGoal()`).
Mixing queries with hidden state mutations is the leading cause of spooky action-at-a-distance bugs in large codebases.
