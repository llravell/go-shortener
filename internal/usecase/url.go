package usecase

import (
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

func (uc *URLUseCase) SaveURL(url string) *entity.URL {
	hash := uc.gen.Generate()
	urlObj := entity.NewURL(url, hash)

	uc.repo.Store(urlObj)

	return urlObj
}

func (uc *URLUseCase) ResolveURL(hash string) (*entity.URL, error) {
	return uc.repo.Get(hash)
}

func (uc *URLUseCase) BuildRedirectURL(url *entity.URL) string {
	return fmt.Sprintf("%s/%s", uc.baseRedirectURL, url.Short)
}
