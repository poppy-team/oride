package concurrency

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestBoundedWorkerPool_BasicExecution(t *testing.T) {
	ctx := context.Background()
	pool := NewBoundedWorkerPool(ctx, 4, 100)

	var counter uint64
	numTasks := 1000

	for i := 0; i < numTasks; i++ {
		err := pool.Submit(ctx, func(c context.Context) error {
			atomic.AddUint64(&counter, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("unexpected submit error: %v", err)
		}
	}

	if err := pool.Shutdown(5 * time.Second); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	if val := atomic.LoadUint64(&counter); val != uint64(numTasks) {
		t.Fatalf("expected %d tasks completed, got %d", numTasks, val)
	}

	if pool.Metrics.TasksCompleted != uint64(numTasks) {
		t.Fatalf("expected %d metrics completed, got %d", numTasks, pool.Metrics.TasksCompleted)
	}
}

func TestBoundedWorkerPool_BackpressureAndSaturation(t *testing.T) {
	ctx := context.Background()
	// 1 worker, 1 queue slot
	pool := NewBoundedWorkerPool(ctx, 1, 1)

	blockedTask := make(chan struct{})
	// Fill worker
	_ = pool.Submit(ctx, func(c context.Context) error {
		<-blockedTask
		return nil
	})

	// Fill queue slot
	_ = pool.Submit(ctx, func(c context.Context) error {
		return nil
	})

	// Third task with TrySubmit must fail with ErrQueueSaturated
	err := pool.TrySubmit(func(c context.Context) error {
		return nil
	})
	if !errors.Is(err, ErrQueueSaturated) {
		t.Fatalf("expected ErrQueueSaturated, got %v", err)
	}

	close(blockedTask)
	if err := pool.Shutdown(2 * time.Second); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}

func TestBoundedWorkerPool_PanicRecovery(t *testing.T) {
	ctx := context.Background()
	pool := NewBoundedWorkerPool(ctx, 2, 10)

	_ = pool.Submit(ctx, func(c context.Context) error {
		panic("simulated worker failure")
	})

	_ = pool.Submit(ctx, func(c context.Context) error {
		return nil
	})

	if err := pool.Shutdown(2 * time.Second); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	if pool.Metrics.PanicsRecovered != 1 {
		t.Fatalf("expected 1 panic recovered, got %d", pool.Metrics.PanicsRecovered)
	}
	if pool.Metrics.TasksCompleted != 1 {
		t.Fatalf("expected 1 task completed, got %d", pool.Metrics.TasksCompleted)
	}
}
