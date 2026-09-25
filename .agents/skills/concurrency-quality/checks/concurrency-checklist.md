# Concurrency Quality & Invariants Verification Checklist

## 1. Data Race Prevention & Memory Consistency
- [ ] All shared mutable state is protected by explicit synchronization primitives (Mutex, RWMutex, Atomics, Channels, or Actor mailboxes).
- [ ] Dynamic race detector passes without race warnings (`go test -race ./...`, Clang/GCC `-fsanitize=thread`, or equivalent).
- [ ] Atomics use the correct memory ordering:
  - [ ] Relaxed ordering (`memory_order_relaxed`) used ONLY for independent monotonic counters.
  - [ ] Acquire-Release pairs (`memory_order_acquire` / `memory_order_release`) used for message passing and flag handshakes.
  - [ ] Sequential Consistency (`memory_order_seq_cst`) used when a single global total order of operations across all threads is mandatory.
- [ ] Contended atomic variables are cache-line aligned (64 bytes) to prevent false sharing and cache-line invalidation storms.

## 2. Deadlock Avoidance & Lock Hierarchy
- [ ] Lock hierarchy is strictly acyclic; all multi-lock acquisitions adhere to a documented global ordering.
- [ ] Critical section scopes are minimized: zero disk I/O, network calls, serialization, or unbounded loops while holding locks.
- [ ] No recursive mutex locks without explicit re-entrancy guarantees.
- [ ] Condition variables (`sync.Cond`) always evaluate predicates inside a loop (`while (!predicate) wait();`) to guard against spurious wakeups.
- [ ] Reader-Writer locks are protected against writer starvation under high reader throughput.

## 3. Structured Concurrency & Task Lifecycles
- [ ] Every background worker, thread, or goroutine is bound to a parent lifecycle scope (`sync.WaitGroup`, Nursery, TaskGroup, or JoinHandle).
- [ ] Zero orphaned tasks: all spawned tasks have guaranteed completion or cancellation paths.
- [ ] Panics or unhandled exceptions inside background tasks are recovered, logged, and propagated to the supervisor without crashing the runtime silently.
- [ ] Shutdown sequence implements orderly draining: stop receiving new tasks $\to$ signal active workers $\to$ wait with timeout $\to$ force terminate if timeout exceeded.

## 4. Backpressure & Queue Boundaries
- [ ] All channels, work queues, and task buffers have explicit, bounded capacities. Zero unbounded in-memory queues.
- [ ] Queue buffer sizes are dimensioned using Little's Law ($L = \lambda W$) and burst capacity requirements.
- [ ] Explicit backpressure policies are defined and tested:
  - [ ] Blocking push when downstream is saturated (rate-matching).
  - [ ] Drop oldest / Drop newest with metric increment for real-time sensor/telemetry streams.
  - [ ] Immediate HTTP 429 / 503 rejection with `Retry-After` header for user-facing API endpoints.

## 5. Cancellation & Timeout Propagation
- [ ] All blocking calls (network connections, database queries, channel receives, semaphore acquisitions) honor cancellation contexts or timeouts.
- [ ] Cancellation signals propagate downwards to all sub-tasks and background dependencies immediately.
- [ ] Clean-up routines (`defer`, RAII destructors, `finally` blocks) use an un-cancelled context or fallback timeout to ensure release of file descriptors and locks.

## 6. Stress Testing & Verification Evidence
- [ ] Automated stress tests run with high concurrency ($N \ge 100$ concurrent workers, $M \ge 10{,}000$ iterations).
- [ ] Verification script `scripts/verify.sh` executes with 0 errors and 0 race detector findings.
