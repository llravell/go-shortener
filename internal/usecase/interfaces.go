package usecase

import "github.com/llravell/go-shortener/internal/entity"

type URLRepo interface {
	Store(url *entity.URL)
	Get(hash string) (*entity.URL, error)
}

type URLBackup interface {
	Store([]entity.URL) error
	Restore() ([]entity.URL, error)
}

type HashGenerator interface {
	Generate(url string) string
}
