package workerpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

const (
	_defaultWorksChanSize = 64
)

var ErrHasBeenAlreadyClosed = errors.New("worker pool has been already closed")

type Work interface {
	Do(ctx context.Context)
}

type WorkerPool[W Work] struct {
	workersAmount int
	worksChan     chan W
	closed        atomic.Bool
	processOnce   sync.Once
	wg            sync.WaitGroup
}

func New[W Work](workersAmount int) *WorkerPool[W] {
	return &WorkerPool[W]{
		workersAmount: workersAmount,
		worksChan:     make(chan W, _defaultWorksChanSize),
	}
}

func (wp *WorkerPool[W]) QueueWork(work W) error {
	if wp.closed.Load() {
		return ErrHasBeenAlreadyClosed
	}

	wp.worksChan <- work

	return nil
}

func (wp *WorkerPool[W]) worker(ctx context.Context) {
	wp.wg.Add(1)
	defer wp.wg.Done()

	for work := range wp.worksChan {
		work.Do(ctx)
	}
}

func (wp *WorkerPool[W]) ProcessQueue(ctx context.Context) {
	if wp.closed.Load() {
		return
	}

	wp.processOnce.Do(func() {
		for range wp.workersAmount {
			go wp.worker(ctx)
		}
	})
}

func (wp *WorkerPool[W]) Close() error {
	hasBeenCanceled := wp.closed.Swap(true)

	if !hasBeenCanceled {
		close(wp.worksChan)
	}

	return nil
}

func (wp *WorkerPool[W]) Wait() {
	wp.wg.Wait()
}

func (wp *WorkerPool[W]) Reset() *WorkerPool[W] {
	if !wp.closed.Load() {
		return wp
	}

	wp.Wait()

	wp.worksChan = make(chan W, _defaultWorksChanSize)
	wp.processOnce = sync.Once{}
	wp.wg = sync.WaitGroup{}

	wp.closed.Store(false)

	return wp
}
