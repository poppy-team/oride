// Package main demonstrates production-grade Go concurrency, error handling, and memory reuse.
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Sentinel errors compared with errors.Is
var (
	ErrTaskFailed    = errors.New("task execution failed")
	ErrQueueExhausted = errors.New("work queue exhausted")
)

// Task represents a discrete unit of work.
type Task struct {
	ID      string
	Payload []byte
}

// Result represents the outcome of a processed Task.
type Result struct {
	TaskID    string
	BytesRead int
	Checksum  uint32
}

// bufferPool recycles scratch buffers to prevent garbage collection churn.
var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// Processor executes tasks concurrently with bounded workers and guaranteed lifecycle termination.
type Processor struct {
	maxWorkers int
	processed  atomic.Uint64
}

// NewProcessor constructs a Processor with validated bounds.
func NewProcessor(workers int) (*Processor, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("workers must be greater than 0, got %d", workers)
	}
	return &Processor{maxWorkers: workers}, nil
}

// ProcessBatch processes tasks concurrently while strictly respecting ctx cancellation.
func (p *Processor) ProcessBatch(ctx context.Context, tasks []Task) ([]Result, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	taskCh := make(chan Task, len(tasks))
	resultCh := make(chan Result, len(tasks))
	errCh := make(chan error, p.maxWorkers)

	// Populate work queue
	for _, t := range tasks {
		taskCh <- t
	}
	close(taskCh)

	var wg sync.WaitGroup
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Launch bounded worker pool
	for w := 0; w < p.maxWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-workerCtx.Done():
					return
				case task, ok := <-taskCh:
					if !ok {
						return // Work queue drained cleanly
					}

					res, err := p.processSingleTask(workerCtx, task)
					if err != nil {
						select {
						case errCh <- fmt.Errorf("worker %d: %w", workerID, err):
						default:
						}
						cancel() // Stop sibling workers on failure
						return
					}

					resultCh <- res
					p.processed.Add(1)
				}
			}
		}(w)
	}

	// Wait for workers to complete in a dedicated lifecycle owner goroutine
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(resultCh)
		close(done)
	}()

	select {
	case <-ctx.Done():
		cancel()
		<-done // Ensure all workers exited before returning
		return nil, fmt.Errorf("batch canceled: %w", ctx.Err())

	case err := <-errCh:
		<-done // Await clean shutdown
		return nil, fmt.Errorf("batch processing aborted: %w", err)

	case <-done:
		// Collect results
		results := make([]Result, 0, len(tasks))
		for r := range resultCh {
			results = append(results, r)
		}
		return results, nil
	}
}

// processSingleTask processes one task using a pooled buffer and value semantics.
func (p *Processor) processSingleTask(ctx context.Context, task Task) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}

	// Acquire scratch buffer from pool
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset() // Crucial: always reset state before reuse
	defer bufferPool.Put(buf)

	if _, err := buf.Write(task.Payload); err != nil {
		return Result{}, fmt.Errorf("writing payload for task %s: %w", task.ID, err)
	}

	// Compute mock checksum
	var checksum uint32
	for _, b := range buf.Bytes() {
		checksum += uint32(b)
	}

	return Result{
		TaskID:    task.ID,
		BytesRead: buf.Len(),
		Checksum:  checksum,
	}, nil
}

// ProcessedCount returns the cumulative count of successfully completed tasks.
func (p *Processor) ProcessedCount() uint64 {
	return p.processed.Load()
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	proc, err := NewProcessor(4)
	if err != nil {
		panic(err)
	}

	tasks := []Task{
		{ID: "T1", Payload: []byte("alpha")},
		{ID: "T2", Payload: []byte("beta")},
		{ID: "T3", Payload: []byte("gamma")},
	}

	results, err := proc.ProcessBatch(ctx, tasks)
	if err != nil {
		fmt.Printf("Batch error: %v\n", err)
		return
	}

	fmt.Printf("Successfully processed %d tasks (total: %d)\n", len(results), proc.ProcessedCount())
}
