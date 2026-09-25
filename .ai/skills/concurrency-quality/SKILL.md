# Concurrency Quality & Parallel Execution Invariants

## Purpose
Enforce provable correctness, data race freedom, deadlock prevention, bounded resource allocation, and predictable throughput across multi-threaded, asynchronous, and distributed concurrent systems. Guarantee deterministic cancellation trees, backpressure propagation, and zero thread/goroutine leakage across system lifecycles.

## Use when
- Designing, implementing, or auditing concurrent subsystems, worker pools, async pipelines, and event loops.
- Synchronizing access to shared mutable state via mutexes, read-write locks, lock-free atomics, or channels.
- Mitigating concurrency hazards: data races, deadlocks, livelocks, thread starvation, priority inversion, and false sharing.
- Establishing cooperative task cancellation, graceful termination sequences, and backpressure mechanisms.
- Validating concurrent code under automated race detectors (`go test -race`, ThreadSanitizer `tsan`, Loom, Tokio console).

## Do not use when
- Implementing purely sequential, single-threaded algorithms with no async boundaries or background tasks.
- Designing high-level relational database schema migrations without connection pooling or transactional isolation analysis (use `database-review`).
- Measuring static web asset loading waterfalls without concurrent runtime workers (use `performance-web`).

## Required context
- Target runtime concurrency model: Go CSP (goroutines/channels), Rust (ownership/Send/Sync/Tokio), C++20 (`std::jthread`/atomics), or Python (asyncio/multiprocessing).
- Memory consistency model: Acquire-Release semantics, Sequential Consistency, or runtime happens-before rules.
- Hardware concurrency constraints: CPU core count, NUMA architecture, cache line size (typically 64 bytes).
- System throughput, latency SLAs, and queue backpressure boundaries.

## Procedure
1. **Model Shared State & Enforce Bernstein Conditions**:
   - For all concurrent operations $T_j$ and $T_k$ ($j \neq k$), mathematically verify Bernstein's Conditions:
     $$I_j \cap O_k = \emptyset, \quad O_j \cap I_k = \emptyset, \quad O_j \cap O_k = \emptyset$$
   - Where shared mutable state is unavoidable, encapsulate state within a protective synchronization boundary (mutex, actor, or atomic register).
2. **Establish Strict Global Lock Hierarchy**:
   - Construct a directed acyclic graph (DAG) of all lock acquisitions.
   - Enforce a strict total order: if lock $L_A$ precedes $L_B$ in the hierarchy, any thread acquiring both must acquire $L_A$ before $L_B$.
   - Never invoke dynamic callbacks, trait methods, or external blocking I/O while holding a mutex lock.
3. **Enforce Bounded Buffers & Backpressure**:
   - Forbid unbounded queues, unbounded channels, and unbounded goroutine/thread spawning.
   - Size queues using Little's Law ($L = \lambda W$) where buffer capacity accommodates peak burst without exhausting system memory.
   - Implement explicit rejection, shedding, or upstream throttling when buffer thresholds are saturated.
4. **Structure Cooperative Cancellation & Task Lifecycles**:
   - Establish parent-child cancellation trees using structured concurrency (e.g., `context.Context`, `CancellationToken`, or nursery scopes).
   - Require all blocking primitives (channel reads, socket polls, condition variables) to listen for cancellation signals.
   - Bind all spawned tasks to a lifecycle tracker (`sync.WaitGroup`, task nursery, or join handle) guaranteeing deterministic joining before shutdown.
5. **Optimize Hardware & Cache Synchronization**:
   - Align atomic variables and contended counters to distinct 64-byte cache line boundaries (`alignas(64)` or padding fields) to eliminate false sharing.
   - Minimize lock hold times: compute results outside the critical section, acquire the lock only to swap pointers/state, and release immediately.
6. **Verify Under Dynamic & Static Race Analyzers**:
   - Execute test suites under dynamic race detection (`-race` / `-fsanitize=thread`).
   - Subject concurrent loops to high-iteration stress tests ($N \ge 10{,}000$) with aggressive context switching to reveal edge interleavings.

## Decision rules
- **Zero Data Races**: Any data race reported by ThreadSanitizer or Go `-race` is a critical severity defect that blocks release.
- **Bounded Resources**: Every channel, queue, and worker pool must have an explicit capacity limit. Unbounded growth is an anti-pattern.
- **Minimal Critical Sections**: Critical sections must contain only memory operations; no network calls, disk I/O, or foreign library calls may execute under a lock.
- **Hierarchical Locking**: Acquiring locks out of hierarchy order is strictly prohibited. If two locks cannot be statically ordered, use `TryLock` with backoff and retry.
- **Graceful Draining**: Applications must drain pending in-flight tasks within a configurable timeout before forcing SIGKILL or aborting.

## Evidence required
- Concurrency architecture specification documenting the lock hierarchy DAG and cancellation propagation.
- Test logs with race detection enabled (`go test -race` or `tsan`) demonstrating 0 race warnings.
- Stress benchmark results demonstrating throughput scaling without lock convoying or thread starvation.
- Verification script execution log from `scripts/verify.sh` exiting with code 0.

## Output contract
- Concurrency design document adhering to `templates/concurrency-architecture-spec.md`.
- Concurrency checklist verification in `checks/concurrency-checklist.md`.
- Reproducible, automated race-detection test harness.

## Stop conditions
- Complete concurrency verification passed with zero data races and zero deadlocks.
- All tasks bounded, structured, and guaranteed to terminate cleanly upon cancellation.
- Token budget exhausted or irreconcilable deadlocks in third-party runtime dependencies.

## Escalation rules
- Escalate to Architecture Lead if third-party libraries require holding locks across asynchronous boundaries.
- Escalate to Infrastructure Lead if memory consumption under peak backpressure exceeds host container cgroup limits.
