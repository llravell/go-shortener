package usecase

import (
	"context"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLRepo interface {
	Store(url *entity.URL)
	GetContext(ctx context.Context, hash string) (*entity.URL, error)
}

type HashGenerator interface {
	Generate() (string, error)
}
