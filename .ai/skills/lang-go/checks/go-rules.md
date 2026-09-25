# Go Production Verification Checklist

## 1. Error Handling & Invariants
- [ ] Every fallible function call has its error return checked immediately.
- [ ] No blank identifiers used for error returns (`_ = fn()`) without explicit registered justification.
- [ ] Errors are wrapped with contextual information using `fmt.Errorf("...: %w", err)`.
- [ ] Sentinel errors compared via `errors.Is(err, TargetErr)` and custom types via `errors.As(err, &custom)`.
- [ ] No application panics in business logic (`panic()` only allowed in package `Must*` initializers).

## 2. Concurrency & Goroutine Ownership
- [ ] Every spawned goroutine (`go fn()`) has a designated owner and guaranteed termination path.
- [ ] `context.Context` is accepted as the first argument in all concurrent or I/O-bound functions.
- [ ] No `context.Context` is stored in struct fields.
- [ ] Every blocking channel operation includes a `case <-ctx.Done():` fallback.
- [ ] Concurrency is bounded (using worker pools, semaphores, or `errgroup.Group` with `SetLimit`).
- [ ] Channels are closed strictly by their sender, never by receivers.
- [ ] Data structures with concurrent access are guarded with `sync.Mutex`, `sync.RWMutex`, or atomics.

## 3. Memory & Resource Hygiene
- [ ] Slices and maps are pre-allocated with known capacity (`make([]T, 0, cap)`).
- [ ] Objects retrieved from `sync.Pool` are completely reset before `Put()`.
- [ ] Network and file bodies (`resp.Body`, `os.File`) are closed via `defer` immediately after error check.
- [ ] Structs containing `sync.Mutex` or `sync.WaitGroup` are passed by pointer, never by value.

## 4. API & Interface Idioms
- [ ] Functions accept interfaces and return concrete structs.
- [ ] Interfaces are defined in the consumer package with minimal method counts (1-2 methods).
- [ ] Constructors (`New*`) return concrete pointers and explicit errors.

## 5. Automated Testing & Verification
- [ ] Tests are table-driven with descriptive test case names.
- [ ] Subtests call `t.Parallel()` safely without loop closure traps.
- [ ] Entire test suite passes cleanly with race detector: `go test -race ./...`.
- [ ] Critical allocations are monitored via `go test -benchmem`.
