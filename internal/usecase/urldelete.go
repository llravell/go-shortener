package usecase

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog"
)

const (
	urlDeleteItemsChanSize = 64
	urlDeleteWorkersAmount = 4
)

type URLDeleteItem struct {
	UserUUID string
	Hashes   []string
}

type URLDeleteUseCase struct {
	repo             URLDeleteRepo
	urlDeleteItemsCh chan *URLDeleteItem
	cancelled        atomic.Bool
	processOnce      sync.Once
	log              zerolog.Logger
	wg               sync.WaitGroup
}

func NewURLDeleteUseCase(repo URLDeleteRepo, log zerolog.Logger) *URLDeleteUseCase {
	return &URLDeleteUseCase{
		repo:             repo,
		log:              log,
		urlDeleteItemsCh: make(chan *URLDeleteItem, urlDeleteItemsChanSize),
	}
}

func (ud *URLDeleteUseCase) Cancel() error {
	hasBeenCanceled := ud.cancelled.Swap(true)

	if !hasBeenCanceled {
		close(ud.urlDeleteItemsCh)
		ud.wg.Wait()
	}

	return nil
}

func (ud *URLDeleteUseCase) worker() {
	defer ud.wg.Done()

	for item := range ud.urlDeleteItemsCh {
		err := ud.repo.DeleteMultiple(context.Background(), item.UserUUID, item.Hashes)
		if err != nil {
			ud.log.Error().
				Err(err).
				Str("userUUID", item.UserUUID).
				Msg("delete urls failed")
		} else {
			ud.log.Info().
				Str("userUUID", item.UserUUID).
				Msg("delete urls successeded")
		}
	}
}

func (ud *URLDeleteUseCase) QueueDelete(item *URLDeleteItem) {
	ud.urlDeleteItemsCh <- item
}

func (ud *URLDeleteUseCase) ProcessQueue() {
	if ud.cancelled.Load() {
		return
	}

	ud.processOnce.Do(func() {
		for range urlDeleteWorkersAmount {
			ud.wg.Add(1)

			go ud.worker()
		}
	})
}
