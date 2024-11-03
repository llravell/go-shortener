package usecase

import "github.com/llravell/go-shortener/internal/entity"

type URLRepo interface {
	Store(url *entity.URL)
	Get(hash string) (*entity.URL, error)
}

type HashGenerator interface {
	Generate() (string, error)
}
