package usecase

import (
	"context"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLRepo interface {
	Store(ctx context.Context, url *entity.URL) (*entity.URL, error)
	StoreMultiple(ctx context.Context, urls []*entity.URL) error
	Get(ctx context.Context, hash string) (*entity.URL, error)
	GetByUserUUID(ctx context.Context, userUUID string) ([]*entity.URL, error)
}

type HealthRepo interface {
	PingContext(ctx context.Context) error
}

type HashGenerator interface {
	Generate() (string, error)
}
