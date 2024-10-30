package usecase

import (
	"github.com/llravell/go-shortener/internal/entity"
)

type URLUseCase struct {
	repo URLRepo
	gen  HashGenerator
}

func NewURLUseCase(r URLRepo, g HashGenerator) *URLUseCase {
	return &URLUseCase{
		repo: r,
		gen:  g,
	}
}

func (uc *URLUseCase) SaveURL(url string) *entity.URL {
	hash := uc.gen.Generate(url)
	urlObj := entity.NewURL(url, hash)

	uc.repo.Store(urlObj)

	return urlObj
}

func (uc *URLUseCase) ResolveURL(hash string) (*entity.URL, error) {
	return uc.repo.Get(hash)
}
