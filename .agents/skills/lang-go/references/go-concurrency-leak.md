# Diagnosing and Preventing Goroutine Leaks in Go

## 1. Anatomy of a Goroutine Leak
A goroutine leak occurs when a goroutine is allocated on the heap (initial stack 2KB to 8KB) and enters a permanent blocked state on an uncoordinated channel, infinite mutex wait, or un-timed I/O operation. Because the runtime scheduler cannot preempt an indefinitely blocked goroutine that has no pending wake-up events, the goroutine remains alive for the entire process lifetime, leading to cumulative memory exhaustion.

## 2. Common Leak Vectors & Patterns

### Pattern A: Unbuffered Channel with Abandoned Receiver
```go
// LEAK: If ctx cancels before work finishes, the worker goroutine hangs on ch <- result forever!
func worker(ctx context.Context) <-chan int {
    ch := make(chan int) // unbuffered!
    go func() {
        res := computeHeavy()
        ch <- res // BLOCKS FOREVER if receiver timed out and stopped listening!
    }()
    return ch
}

// FIX: Buffer channel by 1, or select on ctx.Done()
func workerSafe(ctx context.Context) <-chan int {
    ch := make(chan int, 1) // buffered: write never blocks
    go func() {
        res := computeHeavy()
        select {
        case ch <- res:
        case <-ctx.Done():
        }
    }()
    return ch
}
```

### Pattern B: Leaking HTTP Response Body
Failing to read to completion and close `resp.Body` keeps the underlying TCP connection and its TLS buffers pinned:
```go
// FIX:
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()
// To reuse HTTP/1.1 or HTTP/2 keep-alive connections:
_, _ = io.Copy(io.Discard, resp.Body)
```

### Pattern C: Ticker Without Stop
`time.Tick()` creates a ticker whose channel is never closed and cannot be garbage collected.
```go
// LEAK:
for range time.Tick(1 * time.Second) { ... }

// FIX:
ticker := time.NewTicker(1 * time.Second)
defer ticker.Stop()
for range ticker.C { ... }
```

## 3. Automated Detection in Unit Tests

Integrate `go.uber.org/goleak` in test packages to automatically detect dangling goroutines:

```go
package mypkg_test

import (
    "testing"
    "go.uber.org/goleak"
)

func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m)
}

func TestConcurrentWorkflow(t *testing.T) {
    defer goleak.VerifyNone(t)
    // run test workflow here...
}
```

## 4. Production Profiling with pprof
To analyze live goroutines under load:
```bash
# Capture goroutine stack trace dump
curl -s http://localhost:6060/debug/pprof/goroutine?debug=2 > goroutines.txt

# Inspect call stacks with highest allocation counts
go tool pprof http://localhost:6060/debug/pprof/goroutine
(pprof) top
(pprof) traces
```
Look for stacks blocked in `runtime.gopark`, `chanrecv`, or `chansend`.
