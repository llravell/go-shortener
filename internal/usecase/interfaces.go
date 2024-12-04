package usecase

import (
	"context"

	"github.com/llravell/go-shortener/internal/entity"
)

//go:generate ../../bin/mockgen -destination=../mocks/mock_urlrepo.go -package=mocks . URLRepo
type URLRepo interface {
	Store(ctx context.Context, url *entity.URL) (*entity.URL, error)
	StoreMultipleURLs(ctx context.Context, urls []*entity.URL) error
	GetURL(ctx context.Context, hash string) (*entity.URL, error)
	GetUserURLS(ctx context.Context, userUUID string) ([]*entity.URL, error)
	DeleteMultipleURLs(ctx context.Context, userUUID string, urlHashes []string) error
}

//go:generate ../../bin/mockgen -destination=../mocks/mock_healthrepo.go -package=mocks . HealthRepo
type HealthRepo interface {
	PingContext(ctx context.Context) error
}

//go:generate ../../bin/mockgen -destination=../mocks/mock_hashgen.go -package=mocks . HashGenerator
type HashGenerator interface {
	Generate() (string, error)
}
