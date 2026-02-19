package collector

import (
	"context"
	"sync"
)

type Task func(ctx context.Context)

type WorkerPool struct {
	workerCount int
	tasksChan   chan Task
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewWorkerPool(workerCount int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workerCount: workerCount,
		tasksChan:   make(chan Task, queueSize),
		ctx:         ctx,
		cancel:      cancel,
	}

	for range workerCount {
		pool.wg.Go(pool.worker)
	}

	return pool
}

func (p *WorkerPool) worker() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case task, ok := <-p.tasksChan:
			if !ok {
				return
			}
			task(p.ctx)
		}
	}
}

func (p *WorkerPool) Submit(task Task) bool {
	select {
	case <-p.ctx.Done():
		return false
	case p.tasksChan <- task:
		return true
	}
}

func (p *WorkerPool) Stop() {
	p.cancel()
	close(p.tasksChan)
	p.wg.Wait()
}
