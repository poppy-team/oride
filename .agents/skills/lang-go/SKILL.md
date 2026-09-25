---
name: lang-go
description: Idiomatic Go (1.22+) engineering, explicit error handling with %w/errors.Is/errors.As, goroutine lifecycle ownership, context cancellation propagation, -race detector verification, sync.Pool recycling, and table-driven tests.
---

# Go Idiomatic Reliability & Concurrency Hygiene

## 1. Title and Description
**Go Idiomatic Reliability & Concurrency Hygiene (`lang-go`)**
Enforces idiomatic Go software engineering practices for modern Go (1.22+). Guarantees complete error accounting, zero goroutine leaks, strict happens-before memory visibility, race-clean concurrency, minimal heap allocations, and robust table-driven testing.

## 2. Purpose
Ensure all Go code produced or refactored within the Prumo framework adheres to production-grade Go standards, avoiding silent concurrency panics, memory bloat from uncollected goroutines, unhandled error escapes, and anti-idiomatic abstraction layers.

## 3. Prerequisites
- Go toolchain 1.22 or newer installed (`go version`).
- Project root containing a valid `go.mod` file.
- Race detector dependencies available (CGO or Go pure-race support where applicable).

## 4. Inputs
- Go source files (`*.go`), tests (`*_test.go`), and module dependencies (`go.mod`, `go.sum`).
- Target task acceptance criteria, concurrency budgets, and SLA latency/throughput profiles.

## 5. Outputs
- Idiomatic, formatted (`gofmt`/`goimports`) Go source code.
- Zero-leak concurrency structures with deterministic lifecycle termination.
- Comprehensive test suites with race detection clean passes (`go test -race ./...`).
- Verification report conforming to `templates/go-verification-report.md`.

## 6. Execution Steps

### Step 1: Explicit Error Handling and Wrapping
1. **Never swallow errors**: Never assign a fallible return to blank identifier `_` without an explicit comment and verified safety invariant.
2. **Contextual wrapping**: Wrap errors using `fmt.Errorf("action description: %w", err)` to preserve the causal error chain.
3. **Inspection**:
   - Use `errors.Is(err, target)` to compare sentinel errors (e.g. `io.EOF`, `fs.ErrNotExist`).
   - Use `errors.As(err, &targetStruct)` to extract structured error metadata.
   - Use `errors.Join(err1, err2)` when collecting independent multi-step failure records.
4. **Panic restriction**: Prohibit `panic()` and `recover()` in normal business logic. `panic()` is reserved exclusively for package initialization invariants (`must.go`) or programmer assertion failures that represent impossible states.

### Step 2: Goroutine Lifecycle and Ownership
1. **Every goroutine must have an explicit owner**: The spawning function or struct must maintain knowledge of how and when the spawned goroutine will exit.
2. **Context propagation**: Accept `ctx context.Context` as the very first parameter on all I/O, network, or long-running functions. Never store `context.Context` inside a struct.
3. **Cancellation compliance**: Every blocking channel receive or select block must include a `case <-ctx.Done(): return ctx.Err()` branch.
4. **Bounded concurrency**: Never launch unbounded goroutines inside a loop or HTTP handler. Use worker pools, semaphore channels (`chan struct{}`), or `golang.org/x/sync/errgroup`.

### Step 3: Channel and Synchronization Axioms
1. **Sender closes**: Only the producer/sender goroutine closes a channel. Never close a channel from the receiver, and never close a channel with multiple concurrent senders.
2. **Channel state semantics**:
   - Send to `nil` channel: blocks indefinitely.
   - Receive from `nil` channel: blocks indefinitely.
   - Send to closed channel: panics.
   - Close a closed channel: panics.
   - Receive from closed channel: immediately returns zero-value and `ok == false`.
3. **Mutex vs Channels**:
   - Use `sync.Mutex` or `sync.RWMutex` to protect internal state, in-memory caches, and data structures.
   - Use channels to transfer ownership of data, coordinate lifecycles, and distribute tasks.
4. **One-time initialization**: Use `sync.Once` or `sync.OnceValue` for lazy thread-safe initialization instead of custom atomic checks.

### Step 4: Memory Optimization & Allocations
1. **Slice & Map pre-allocation**: Always pre-allocate slices and maps when capacity is known: `make([]T, 0, expectedLen)` or `make(map[K]V, expectedKeys)`.
2. **Recycling with `sync.Pool`**:
   - Only pool frequently allocated, short-lived objects (e.g., serialization buffers, scratch byte slices).
   - **Crucial**: Always reset object state before returning it to the pool via `defer pool.Put(buf)`.
3. **Pointer vs Value semantics**:
   - Return structs by value for small types (<= 64 bytes) to encourage stack allocation and eliminate GC pressure.
   - Use pointer receivers on methods if the method modifies the receiver, or if the struct contains synchronization primitives (`sync.Mutex` must never be copied).

### Step 5: Interface Design
1. **Accept interfaces, return structs**: Functions should receive the minimal interface required (`io.Reader` rather than `*os.File`) and return concrete types.
2. **Interface location**: Define interfaces in the consumer package, not in the producer package.
3. **Keep interfaces small**: Single-method interfaces (`io.Reader`, `io.Writer`, `fmt.Stringer`) compose naturally and are easily mocked in tests.

### Step 6: Table-Driven Testing & Race Detection
1. **Table-driven structure**: Formulate test cases as slices of structs containing `name`, `inputs`, `expected`, and `wantErr`.
2. **Subtests and Parallelism**:
   - Run test cases via `t.Run(tc.name, func(t *testing.T) { ... })`.
   - Call `t.Parallel()` inside the subtest. (Note: In Go 1.22+, loop variables have per-iteration scope, eliminating the `tc := tc` trap).
3. **Execution Gate**: Execute `go test -race -count=1 ./...` on every change. A single data race is an immediate release blocker.

## 7. Verification
Run the verification suite:
```bash
# 1. Format and import hygiene
gofmt -s -w .
go vet ./...

# 2. Race-detector test execution
go test -race -cover ./...

# 3. Static analysis
if command -v staticcheck >/dev/null 2>&1; then
  staticcheck ./...
fi

# 4. Allocation profiling (on critical hot paths)
go test -benchmem -bench=. ./...
```

## 8. Fallbacks & Failure Recovery
| Failure Class | Cause | Deterministic Remediation |
|---|---|---|
| Data Race Detected | Unsynchronized concurrent read/write to shared variable | Identify access points via `-race` stack trace. Wrap in `sync.Mutex` or replace with `sync/atomic` if primitive. |
| Goroutine Leak | Goroutine blocked on channel send/receive with no active reader/writer | Ensure channel buffer size matches expectations, or select on `ctx.Done()` with `defer cancel()`. Use `uber-go/goleak` in tests. |
| Memory Bloat from `sync.Pool` | Stale or huge buffers retained in pool | Limit maximum size of buffers returned to pool (drop if `cap > 64KB`). |
| Nil Pointer Dereference | Untyped nil in interface comparison | Remember `var p *T = nil`; interface holding `p` is not nil (`val != nil` is true). Return untyped `nil` explicitly. |

## 9. Constraints
- **Zero raw blank errors**: Code that uses `_ = fallibleFunc()` without explicit recorded rationale is rejected.
- **No global mutable singletons**: All dependencies must be passed explicitly via constructor functions (`NewService(...)`).
- **No context stored in structs**: `context.Context` must remain ephemeral and flow through call stacks.
- **Pass with `-race`**: Code failing `go test -race` will not pass the quality gate under any circumstance.

## 10. Examples

### Anti-Pattern: Goroutine Leak & Ignored Error
```go
// BAD: Leaks goroutine if caller aborts, ignores channel error, no context cancellation
func FetchAll(urls []string) []string {
    ch := make(chan string) // unbuffered!
    for _, u := range urls {
        go func(url string) {
            resp, err := http.Get(url)
            if err != nil {
                return // ERROR: if receiver stopped reading, channel send will hang forever!
            }
            body, _ := io.ReadAll(resp.Body) // ERROR: ignored error, leaks body
            ch <- string(body)
        }(u)
    }
    var results []string
    for range urls {
        results = append(results, <-ch)
    }
    return results
}
```

### Idiomatic Pattern: Structured Concurrency with errgroup
```go
// GOOD: Bounded concurrency, context propagation, resource cleanup, clean errors
func FetchAll(ctx context.Context, client *http.Client, urls []string, maxConcurrency int) ([]string, error) {
    g, ctx := errgroup.WithContext(ctx)
    g.SetLimit(maxConcurrency)

    results := make([]string, len(urls))

    for i, u := range urls {
        g.Go(func() error {
            req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
            if err != nil {
                return fmt.Errorf("create request for %s: %w", u, err)
            }

            resp, err := client.Do(req)
            if err != nil {
                return fmt.Errorf("execute request to %s: %w", u, err)
            }
            defer resp.Body.Close()

            if resp.StatusCode != http.StatusOK {
                return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, u)
            }

            body, err := io.ReadAll(resp.Body)
            if err != nil {
                return fmt.Errorf("read response body from %s: %w", u, err)
            }

            results[i] = string(body)
            return nil
        })
    }

    if err := g.Wait(); err != nil {
        return nil, fmt.Errorf("fetching URLs: %w", err)
    }
    return results, nil
}
```
