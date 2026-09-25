package concurrency

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed   = errors.New("worker pool is closed")
	ErrQueueSaturated = errors.New("queue is saturated (backpressure)")
)

// Task represents an executable work unit.
type Task func(ctx context.Context) error

// WorkerMetrics demonstrates cache-line padding (64 bytes) to eliminate false sharing
// on heavily contended atomic counters.
type WorkerMetrics struct {
	TasksCompleted uint64
	_pad1          [56]byte // 8 + 56 = 64 bytes (Cache Line 1)

	TasksFailed    uint64
	_pad2          [56]byte // 8 + 56 = 64 bytes (Cache Line 2)

	PanicsRecovered uint64
	_pad3           [56]byte // 8 + 56 = 64 bytes (Cache Line 3)
}

// BoundedWorkerPool implements structured concurrency with bounded buffering and graceful draining.
type BoundedWorkerPool struct {
	workerCount int
	taskQueue   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	closed      atomic.Bool
	Metrics     WorkerMetrics
}

// NewBoundedWorkerPool instantiates a pool with bounded capacity and parent context.
func NewBoundedWorkerPool(parentCtx context.Context, workers int, queueCapacity int) *BoundedWorkerPool {
	if workers <= 0 {
		workers = 1
	}
	if queueCapacity <= 0 {
		queueCapacity = 1
	}

	ctx, cancel := context.WithCancel(parentCtx)
	pool := &BoundedWorkerPool{
		workerCount: workers,
		taskQueue:   make(chan Task, queueCapacity),
		ctx:         ctx,
		cancel:      cancel,
	}

	pool.start()
	return pool
}

func (p *BoundedWorkerPool) start() {
	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.workerLoop(i)
	}
}

func (p *BoundedWorkerPool) workerLoop(workerID int) {
	defer p.wg.Done()

	for {
		select {
		case <-p.ctx.Done():
			// Context cancelled; drain remaining tasks if needed or terminate
			return
		case task, ok := <-p.taskQueue:
			if !ok {
				// Queue channel closed; orderly exit
				return
			}
			p.executeTaskSafely(task)
		}
	}
}

func (p *BoundedWorkerPool) executeTaskSafely(task Task) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddUint64(&p.Metrics.PanicsRecovered, 1)
		}
	}()

	if err := task(p.ctx); err != nil {
		atomic.AddUint64(&p.Metrics.TasksFailed, 1)
	} else {
		atomic.AddUint64(&p.Metrics.TasksCompleted, 1)
	}
}

// Submit blocks if queue is full, applying natural backpressure until timeout or context cancel.
func (p *BoundedWorkerPool) Submit(ctx context.Context, task Task) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}

	select {
	case <-p.ctx.Done():
		return p.ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	case p.taskQueue <- task:
		return nil
	}
}

// TrySubmit attempts to submit immediately without blocking; returns ErrQueueSaturated if full.
func (p *BoundedWorkerPool) TrySubmit(task Task) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}

	select {
	case p.taskQueue <- task:
		return nil
	default:
		return ErrQueueSaturated
	}
}

// Shutdown gracefully closes the task queue and awaits worker completion within drainTimeout.
func (p *BoundedWorkerPool) Shutdown(drainTimeout time.Duration) error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil // Already shutting down
	}

	close(p.taskQueue)

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.cancel()
		return nil
	case <-time.After(drainTimeout):
		// Drain timeout expired; force cancel workers
		p.cancel()
		<-done
		return fmt.Errorf("shutdown drain timeout exceeded (%v)", drainTimeout)
	}
}
