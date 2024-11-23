package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/usecase/repo"
)

var ErrURLDuplicate = errors.New("duplicate url")

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

func (uc *URLUseCase) SaveURL(ctx context.Context, url string, userUUID string) (*entity.URL, error) {
	hash, err := uc.gen.Generate()
	if err != nil {
		return nil, err
	}

	urlObj := &entity.URL{Original: url, Short: hash, UserUUID: userUUID}
	fmt.Println(urlObj)

	storedURL, err := uc.repo.Store(ctx, urlObj)
	if errors.Is(err, repo.ErrOriginalURLConflict) {
		return storedURL, ErrURLDuplicate
	}

	return storedURL, err
}

func (uc *URLUseCase) SaveURLMultiple(ctx context.Context, urls []string, userUUID string) ([]*entity.URL, error) {
	urlObjs := make([]*entity.URL, 0, len(urls))

	if len(urls) == 0 {
		return urlObjs, nil
	}

	for _, url := range urls {
		hash, err := uc.gen.Generate()
		if err != nil {
			return urlObjs, err
		}

		urlObjs = append(urlObjs, &entity.URL{Original: url, Short: hash, UserUUID: userUUID})
	}

	return urlObjs, uc.repo.StoreMultiple(ctx, urlObjs)
}

func (uc *URLUseCase) ResolveURL(ctx context.Context, hash string) (*entity.URL, error) {
	return uc.repo.Get(ctx, hash)
}

func (uc *URLUseCase) GetUserURLS(ctx context.Context, userUUID string) ([]*entity.URL, error) {
	return uc.repo.GetByUserUUID(ctx, userUUID)
}

func (uc *URLUseCase) BuildRedirectURL(url *entity.URL) string {
	return fmt.Sprintf("%s/%s", uc.baseRedirectURL, url.Short)
}
