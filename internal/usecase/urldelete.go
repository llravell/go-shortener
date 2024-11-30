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
	log      *zerolog.Logger
	UserUUID string
	Hashes   []string
}

func (w *URLDeleteWork) Do(ctx context.Context) {
	err := w.repo.DeleteMultiple(ctx, w.UserUUID, w.Hashes)
	if err != nil {
		w.log.Error().
			Err(err).
			Str("userUUID", w.UserUUID).
			Msg("delete urls failed")
	} else {
		w.log.Info().
			Str("userUUID", w.UserUUID).
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
		log:      &uc.log,
		UserUUID: deleteItem.UserUUID,
		Hashes:   deleteItem.Hashes,
	}

	return uc.wp.QueueWork(deleteWork)
}
