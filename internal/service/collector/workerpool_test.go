package collector

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	pool := NewWorkerPool(5, 50)
	defer pool.Stop()

	var counter int64
	numTasks := 20

	for range numTasks {
		ok := pool.Submit(func(ctx context.Context) {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt64(&counter, 1)
		})
		if !ok {
			t.Fatal("failed to submit task")
		}
	}

	time.Sleep(200 * time.Millisecond)

	finalVal := atomic.LoadInt64(&counter)
	if finalVal != int64(numTasks) {
		t.Fatalf("expected counter %d, got %d", numTasks, finalVal)
	}
}
