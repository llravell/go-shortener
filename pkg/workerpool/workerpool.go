package workerpool

import (
	"errors"
	"sync"
	"sync/atomic"
)

const (
	_defaultWorksChanSize = 64
)

var ErrHasBeenAlreadyClosed = errors.New("worker pool has been already closed")

type Work interface {
	Do()
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

func (wp *WorkerPool[W]) worker() {
	defer wp.wg.Done()

	for work := range wp.worksChan {
		work.Do()
	}
}

func (wp *WorkerPool[W]) ProcessQueue() {
	if wp.closed.Load() {
		return
	}

	wp.processOnce.Do(func() {
		for range wp.workersAmount {
			wp.wg.Add(1)
			go wp.worker()
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
