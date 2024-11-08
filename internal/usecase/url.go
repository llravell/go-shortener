package usecase

import (
	"context"
	"fmt"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLUseCase struct {
	repo            URLRepo
	gen             HashGenerator
	baseRedirectURL string
}

func NewURLUseCase(repo URLRepo, gen HashGenerator, baseRedirectURL string) *URLUseCase {
	return &URLUseCase{
		repo:            repo,
		gen:             gen,
		baseRedirectURL: baseRedirectURL,
	}
}

func (uc *URLUseCase) SaveURL(ctx context.Context, url string) (*entity.URL, error) {
	hash, err := uc.gen.Generate()
	if err != nil {
		return nil, err
	}

	urlObj := &entity.URL{Original: url, Short: hash}

	return uc.repo.Store(ctx, urlObj)
}

func (uc *URLUseCase) ResolveURL(ctx context.Context, hash string) (*entity.URL, error) {
	return uc.repo.Get(ctx, hash)
}

func (uc *URLUseCase) BuildRedirectURL(url *entity.URL) string {
	return fmt.Sprintf("%s/%s", uc.baseRedirectURL, url.Short)
}
