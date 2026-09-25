# Concurrency Engineering & Parallel Systems Reference Guide

## 1. Theoretical Foundations of Concurrency & Parallelism

### 1.1 Bernstein Conditions
For two concurrent tasks $T_1$ and $T_2$ with input sets $I_1, I_2$ (memory read) and output sets $O_1, O_2$ (memory written), deterministic, race-free parallel execution is guaranteed if and only if all three Bernstein conditions hold:
$$I_1 \cap O_2 = \emptyset \quad (\text{No Read-After-Write conflict})$$
$$O_1 \cap I_2 = \emptyset \quad (\text{No Write-After-Read conflict})$$
$$O_1 \cap O_2 = \emptyset \quad (\text{No Write-After-Write conflict})$$
If any intersection is non-empty, access must be synchronized via mutual exclusion, channel sequencing, or lock-free atomics.

### 1.2 Scaling Limits: Amdahl, Gustafson & Gunther's USL
1. **Amdahl's Law** (Fixed problem size, variable processors $s$):
   $$S(s) = \frac{1}{(1 - p) + \frac{p}{s}}$$
   where $p$ is the parallelizable fraction of execution time. As $s \to \infty$, maximum speedup is bounded asymptotically by $\frac{1}{1 - p}$.
2. **Gustafson's Law** (Scaled speedup with workload expanding with core count):
   $$S(s) = s - \alpha(s - 1)$$
   where $\alpha = 1 - p$ is the serial execution proportion.
3. **Gunther's Universal Scalability Law (USL)**:
   Accounts for both contention (queuing for shared resources) and coherency delay (cache-line invalidation traffic across interconnects):
   $$C(N) = \frac{N}{1 + \sigma(N - 1) + \kappa N(N - 1)}$$
   - $N$: Concurrency level (threads/workers).
   - $\sigma \ge 0$: Contention parameter (serialization / lock wait time).
   - $\kappa \ge 0$: Coherency parameter (crosstalk / cache coherence overhead).
   - When $\kappa > 0$, throughput peaks at $N^* = \sqrt{\frac{1 - \sigma}{\kappa}}$ and subsequently degrades (retrograde scaling).

### 1.3 Queue Sizing via Little's Law
For any stable queuing system, the average number of items in the queue $L$ equals the average arrival rate $\lambda$ multiplied by the average time an item spends in the system $W$:
$$L = \lambda \cdot W$$
To absorb bursts of duration $\Delta t_{\text{burst}}$ without dropping requests or overflowing memory:
$$C_{\text{buffer}} \ge \lambda_{\text{peak}} \cdot (W_{\text{p99}} - W_{\text{avg}}) + \int_{0}^{\Delta t_{\text{burst}}} (\lambda(t) - \mu(t)) \, dt$$
where $\mu(t)$ is consumer processing service rate.

---

## 2. Hardware Architecture & Memory Consistency

### 2.1 The Cache Coherence Hierarchy & False Sharing
Modern multi-core processors maintain cache coherence via protocols such as MESI (Modified, Exclusive, Shared, Invalid) or MOESI.
- Cache lines are typically **64 bytes** wide.
- **False Sharing**: Occurs when two threads running on different cores modify independent variables that reside within the same 64-byte cache line. The hardware treats the entire line as contended, continually invalidating and bouncing the cache line across CPU sockets, causing severe performance collapse ($10\times - 50\times$ slowdown).
- **Remedy**: Align independent per-thread or atomic counters to 64-byte boundaries:
  ```go
  // In Go:
  type WorkerStats struct {
      OpsCompleted uint64
      _pad         [56]byte // 8 + 56 = 64 bytes (1 cache line)
  }
  ```
  ```cpp
  // In C++20:
  struct alignas(hardware_destructive_interference_size) WorkerStats {
      std::atomic<uint64_t> ops_completed;
  };
  ```

### 2.2 Memory Ordering Semantics
Hardware reordering (Store Buffers, Invalid Queues) and compiler optimizations require explicit memory order specifications:

| Memory Order | Read / Write | Guarantee | Hardware Cost (x86) |
|---|---|---|---|
| `Relaxed` (`memory_order_relaxed`) | Either | Atomicity only; zero ordering constraints. | Free (standard MOV) |
| `Acquire` (`memory_order_acquire`) | Read | No subsequent reads/writes can be reordered before this read. | Free on x86 (implied) |
| `Release` (`memory_order_release`) | Write | No preceding reads/writes can be reordered after this write. | Free on x86 (implied) |
| `AcqRel` (`memory_order_acq_rel`) | RMW | Combines acquire and release semantics. | Free on x86 |
| `SeqCst` (`memory_order_seq_cst`) | Either | Globally consistent total order across all threads. | Requires `MFENCE` / `LOCK` prefix |

---

## 3. Deadlock Prevention & Lock Hierarchies

### 3.1 Coffman Conditions
A deadlock can arise if and only if all four conditions hold simultaneously:
1. **Mutual Exclusion**: Resources cannot be shared simultaneously.
2. **Hold and Wait**: A task holds at least one resource while awaiting another.
3. **No Preemption**: Resources cannot be forcibly taken from a holding task.
4. **Circular Wait**: A closed chain of tasks exists where each task holds a resource needed by the next.

### 3.2 Lock Ordering DAG
The only scalable programmatic defense against deadlocks is eliminating **Circular Wait** by enforcing a strict partial or total order over all locks:
$$\text{Rank}(L_1) < \text{Rank}(L_2) < \dots < \text{Rank}(L_n)$$
Rule: A thread holding lock $L_i$ may only acquire lock $L_j$ if $\text{Rank}(L_j) > \text{Rank}(L_i)$.

```
+-------------------------------------------------------------+
|                STRICT LOCK HIERARCHY DAG                    |
|                                                             |
|   [ Global Registry Lock ] (Rank 10)                        |
|              |                                              |
|              v                                              |
|   [ Session State Lock ]   (Rank 20)                        |
|              |                                              |
|              v                                              |
|   [ Channel Buffer Lock ]  (Rank 30)                        |
|                                                             |
|   * Rule: Never acquire upstream locks while holding        |
|           downstream locks (e.g. Never 30 -> 10).           |
+-------------------------------------------------------------+
```

---

## 4. Structured Concurrency & Graceful Shutdown

### 4.1 Structured Concurrency Principles
- Concurrency constructs must mirror lexical call scopes: task lifecycles are strictly nested within their spawning scope.
- If a child task fails or panics, the failure is automatically propagated to sibling tasks (cancelling them) and reported to the parent supervisor.
- A parent function cannot return until all its child tasks have terminated cleanly.

### 4.2 Four-Stage Graceful Shutdown Protocol
1. **Quiesce (Stop Ingestion)**: Close listeners, return HTTP 503 / gRPC `UNAVAILABLE` to new traffic.
2. **Drain (Process In-Flight)**: Allow existing work queues to empty up to a configured drain deadline $T_{\text{drain}}$.
3. **Cancel (Broadcast Termination)**: Trigger `context.CancelFunc()` or cancellation tokens to abort long-running tasks.
4. **Reap (Join & Flush)**: Await `WaitGroup.Wait()` or join handles, flush write-ahead logs, and release OS descriptors.
