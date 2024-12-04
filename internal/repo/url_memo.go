package repo

import (
	"context"
	"fmt"
	"sync"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLMemoRepo struct {
	m  map[string]*entity.URL
	mu sync.Mutex
}

type URLNotFoundError struct {
	hash string
}

func (err *URLNotFoundError) Error() string {
	return fmt.Sprintf(`Not found url with hash "%s"`, err.hash)
}

func NewURLMemoRepo() *URLMemoRepo {
	return &URLMemoRepo{
		m: make(map[string]*entity.URL),
	}
}

func (r *URLMemoRepo) Store(_ context.Context, url *entity.URL) (*entity.URL, error) {
	r.mu.Lock()
	r.m[url.Short] = url
	r.mu.Unlock()

	return url, nil
}

func (r *URLMemoRepo) StoreMultipleURLs(_ context.Context, urls []*entity.URL) error {
	r.mu.Lock()
	for _, url := range urls {
		r.m[url.Short] = url
	}
	r.mu.Unlock()

	return nil
}

func (r *URLMemoRepo) GetURL(_ context.Context, hash string) (*entity.URL, error) {
	r.mu.Lock()
	url, ok := r.m[hash]
	r.mu.Unlock()

	if !ok {
		return nil, &URLNotFoundError{hash}
	}

	return url, nil
}

func (r *URLMemoRepo) GetUserURLS(_ context.Context, userUUID string) ([]*entity.URL, error) {
	urls := make([]*entity.URL, 0)

	r.mu.Lock()
	for _, url := range r.m {
		if url.UserUUID == userUUID {
			urls = append(urls, url)
		}
	}
	r.mu.Unlock()

	return urls, nil
}

func (r *URLMemoRepo) GetList() []*entity.URL {
	list := make([]*entity.URL, 0, len(r.m))

	r.mu.Lock()
	for _, url := range r.m {
		list = append(list, url)
	}
	r.mu.Unlock()

	return list
}

func (r *URLMemoRepo) Init(urls []*entity.URL) {
	r.mu.Lock()
	for _, url := range urls {
		r.m[url.Short] = url
	}
	r.mu.Unlock()
}

func (r *URLMemoRepo) DeleteMultipleURLs(_ context.Context, userUUID string, urlHashes []string) error {
	urlHashesToDelete := make(map[string]struct{}, len(urlHashes))

	for _, hash := range urlHashes {
		urlHashesToDelete[hash] = struct{}{}
	}

	r.mu.Lock()
	for _, url := range r.m {
		if url.UserUUID != userUUID {
			continue
		}

		_, shouldBeDeleted := urlHashesToDelete[url.Short]
		if shouldBeDeleted {
			url.Deleted = true
		}
	}
	r.mu.Unlock()

	return nil
}
