// Пакет workerpool представляет реализацию одноименного паттерна.
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

// ErrHasBeenAlreadyClosed ошибка повторного закрытия WorkerPool.
var ErrHasBeenAlreadyClosed = errors.New("worker pool has been already closed")

// Work определяет интерфейс выполняемых задач.
type Work interface {
	Do(ctx context.Context)
}

// WorkerPool структура, предоставляющая интерфейс для распараллеливания задач.
type WorkerPool[W Work] struct {
	workersAmount int
	worksChan     chan W
	doneChan      chan struct{}
	closed        atomic.Bool
	processOnce   sync.Once
	wg            sync.WaitGroup
}

// New создает инстанс WorkerPool'а, дает возможность задать количество воркеров.
func New[W Work](workersAmount int) *WorkerPool[W] {
	return &WorkerPool[W]{
		workersAmount: workersAmount,
		worksChan:     make(chan W, _defaultWorksChanSize),
		doneChan:      make(chan struct{}),
	}
}

// QueueWork добавляет задачу в очередь на обработку.
func (wp *WorkerPool[W]) QueueWork(work W) error {
	if wp.closed.Load() {
		return ErrHasBeenAlreadyClosed
	}

	wp.worksChan <- work

	return nil
}

func (wp *WorkerPool[W]) worker(ctx context.Context) {
	defer wp.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case work, ok := <-wp.worksChan:
			if !ok {
				return
			}

			work.Do(ctx)
		}
	}
}

// ProcessQueue запускает цикл выполнения задач из очереди.
func (wp *WorkerPool[W]) ProcessQueue() {
	if wp.closed.Load() {
		return
	}

	wp.processOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			<-wp.doneChan
			cancel()
		}()

		for range wp.workersAmount {
			wp.wg.Add(1)
			go wp.worker(ctx)
		}
	})
}

// Close оповещает воркеров об окончании работ, закрывает каналы.
func (wp *WorkerPool[W]) Close() error {
	hasBeenCanceled := wp.closed.Swap(true)

	if !hasBeenCanceled {
		close(wp.doneChan)
		close(wp.worksChan)
	}

	return nil
}

// Wait дожидается окончания работ всех воркеров.
func (wp *WorkerPool[W]) Wait() {
	wp.wg.Wait()
}
