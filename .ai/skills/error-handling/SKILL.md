---
name: error-handling
description: Error taxonomy, recoverable/fatal, retries, contextual errors, user-facing vs internal, logging, propagation, panic/exception boundaries
---
# Error Handling Strategy

## 1. Error Taxonomy
Establish a clear taxonomy of errors within the domain. Distinguish between expected domain errors (e.g., 'Insufficient Funds') and unexpected technical faults (e.g., 'Database Connection Lost'). Use strong types for domain errors to enable programmatic handling.

## 2. Recoverable vs. Fatal Errors
Define strict boundaries between recoverable and fatal errors. Recoverable errors should be handled gracefully, often returning an error result to the caller. Fatal errors (e.g., out of memory, corrupted internal state) should trigger an immediate fail-fast mechanism (panic/crash) to prevent cascading failures.

## 3. Retry Strategies
Implement intelligent retry mechanisms for transient technical faults (e.g., network timeouts). Use exponential backoff with jitter to avoid thundering herd problems. Ensure operations are idempotent before retrying. Define a maximum retry limit to prevent infinite loops.

## 4. Contextual Errors
Errors must carry sufficient context for debugging. Wrap low-level errors with high-level context as they propagate up the stack. Include relevant IDs, state information, and the attempted operation. Never return a bare 'Not Found' without specifying *what* was not found.

## 5. User-Facing vs. Internal Errors
Strictly separate internal error representations from user-facing error messages. Internal errors contain technical details, stack traces, and sensitive data. User-facing errors must be safe, sanitized, localized, and actionable, never leaking system internals.

## 6. Structured Error Logging
Log errors systematically using structured formats (e.g., JSON). Include error severity, context fields, correlation IDs, and stack traces. Ensure error logs are easily searchable and aggregate-able in observability platforms. Alert on unexpected spikes in error rates.

## 7. Error Propagation
Establish consistent rules for error propagation. Choose either explicit error returns (e.g., Go, Rust) or structured exception handling (e.g., Java, Python). Avoid mixing paradigms. Ensure errors are propagated to a layer capable of making a meaningful decision about them.

## 8. Panic and Exception Boundaries
Define clear boundaries where exceptions must be caught and converted to standard responses (e.g., HTTP middleware catching panics to return 500s). Prevent internal exceptions from leaking across process boundaries or API endpoints.

## 9. Circuit Breakers
Implement circuit breaker patterns to prevent repeated failures from overwhelming downstream services. When a service fails repeatedly, open the circuit to fail fast locally. Periodically probe the service and close the circuit only when health is restored.

## 10. Error Rate Monitoring
Continuously monitor error rates as a key service level indicator (SLI). Define acceptable error budgets. Automatically trigger alerts or rollbacks if error budgets are exhausted. Differentiate between baseline background noise and anomalous error spikes.

