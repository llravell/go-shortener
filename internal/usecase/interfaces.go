package usecase

import (
	"context"

	"github.com/llravell/go-shortener/internal/entity"
)

// Интерфейсы сторонних зависимостей.
//
//go:generate ../../bin/mockgen -destination=../mocks/mock_usecase.go -package=mocks . URLRepo,HealthRepo,StatsRepo,HashGenerator
type (
	URLRepo interface {
		Store(ctx context.Context, url *entity.URL) (*entity.URL, error)
		StoreMultipleURLs(ctx context.Context, urls []*entity.URL) error
		GetURL(ctx context.Context, hash string) (*entity.URL, error)
		GetUserURLS(ctx context.Context, userUUID string) ([]*entity.URL, error)
		DeleteMultipleURLs(ctx context.Context, userUUID string, urlHashes []string) error
	}

	HealthRepo interface {
		PingContext(ctx context.Context) error
	}

	StatsRepo interface {
		GetURLsAmount(ctx context.Context) (int, error)
		GetUsersAmount(ctx context.Context) (int, error)
	}

	HashGenerator interface {
		Generate() (string, error)
	}
)
