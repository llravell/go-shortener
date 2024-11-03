package repo

import (
	"fmt"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLRepo struct {
	m map[string]*entity.URL
}

type URLNotFoundError struct {
	hash string
}

func (err *URLNotFoundError) Error() string {
	return fmt.Sprintf(`Not found url with hash "%s"`, err.hash)
}

func NewURLStorage() *URLRepo {
	return &URLRepo{make(map[string]*entity.URL)}
}

func (u *URLRepo) Store(url *entity.URL) {
	u.m[url.Short] = url
}

func (u *URLRepo) Get(hash string) (*entity.URL, error) {
	url, ok := u.m[hash]
	if !ok {
		return nil, &URLNotFoundError{hash}
	}

	return url, nil
}

func (u *URLRepo) GetList() []*entity.URL {
	list := make([]*entity.URL, 0, len(u.m))

	for _, url := range u.m {
		list = append(list, url)
	}

	return list
}

func (u *URLRepo) Init(urls []*entity.URL) {
	for _, url := range urls {
		u.m[url.Short] = url
	}
}
