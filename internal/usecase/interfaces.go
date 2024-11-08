package usecase

import (
	"context"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLRepo interface {
	Store(ctx context.Context, url *entity.URL) (*entity.URL, error)
	Get(ctx context.Context, hash string) (*entity.URL, error)
}

type HashGenerator interface {
	Generate() (string, error)
}
