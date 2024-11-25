package repo

import (
	"context"
	"fmt"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLMemoRepo struct {
	m map[string]*entity.URL
}

type URLNotFoundError struct {
	hash string
}

func (err *URLNotFoundError) Error() string {
	return fmt.Sprintf(`Not found url with hash "%s"`, err.hash)
}

func NewURLMemoRepo() *URLMemoRepo {
	return &URLMemoRepo{make(map[string]*entity.URL)}
}

func (u *URLMemoRepo) Store(_ context.Context, url *entity.URL) (*entity.URL, error) {
	u.m[url.Short] = url

	return url, nil
}

func (u *URLMemoRepo) StoreMultiple(_ context.Context, urls []*entity.URL) error {
	for _, url := range urls {
		u.m[url.Short] = url
	}

	return nil
}

func (u *URLMemoRepo) Get(_ context.Context, hash string) (*entity.URL, error) {
	url, ok := u.m[hash]
	if !ok {
		return nil, &URLNotFoundError{hash}
	}

	return url, nil
}

func (u *URLMemoRepo) GetByUserUUID(_ context.Context, userUUID string) ([]*entity.URL, error) {
	urls := make([]*entity.URL, 0)

	for _, url := range u.m {
		if url.UserUUID == userUUID {
			urls = append(urls, url)
		}
	}

	return urls, nil
}

func (u *URLMemoRepo) GetList() []*entity.URL {
	list := make([]*entity.URL, 0, len(u.m))

	for _, url := range u.m {
		list = append(list, url)
	}

	return list
}

func (u *URLMemoRepo) Init(urls []*entity.URL) {
	for _, url := range urls {
		u.m[url.Short] = url
	}
}

func (u *URLMemoRepo) DeleteMultiple(ctx context.Context, userUUID string, urlHashes []string) error {
	urlHashesToDelete := make(map[string]struct{}, len(urlHashes))

	for _, hash := range urlHashes {
		urlHashesToDelete[hash] = struct{}{}
	}

	for _, url := range u.m {
		if url.UserUUID != userUUID {
			continue
		}

		_, shouldBeDeleted := urlHashesToDelete[url.Short]
		if shouldBeDeleted {
			url.Deleted = true
		}
	}

	return nil
}
