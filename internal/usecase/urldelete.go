package usecase

import (
	"context"

	"github.com/llravell/go-shortener/internal/entity"
	"github.com/rs/zerolog"
)

type URLDeleteWorkerPool interface {
	QueueWork(w *URLDeleteWork) error
}

type URLDeleteWork struct {
	repo     URLRepo
	log      zerolog.Logger
	userUUID string
	hashes   []string
}

func (w *URLDeleteWork) Do(ctx context.Context) {
	err := w.repo.DeleteMultiple(ctx, w.userUUID, w.hashes)
	if err != nil {
		w.log.Error().
			Err(err).
			Str("userUUID", w.userUUID).
			Msg("delete urls failed")
	} else {
		w.log.Info().
			Str("userUUID", w.userUUID).
			Msg("delete urls successeded")
	}
}

type URLDeleteUseCase struct {
	repo URLRepo
	wp   URLDeleteWorkerPool
	log  zerolog.Logger
}

func NewURLDeleteUseCase(repo URLRepo, wp URLDeleteWorkerPool, log zerolog.Logger) *URLDeleteUseCase {
	return &URLDeleteUseCase{
		repo: repo,
		wp:   wp,
		log:  log,
	}
}

func (uc *URLDeleteUseCase) QueueDelete(deleteItem *entity.URLDeleteItem) error {
	deleteWork := &URLDeleteWork{
		repo:     uc.repo,
		log:      uc.log,
		userUUID: deleteItem.UserUUID,
		hashes:   deleteItem.Hashes,
	}

	return uc.wp.QueueWork(deleteWork)
}
