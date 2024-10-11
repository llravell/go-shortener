package repo

import "fmt"

type urlRepo struct {
	m map[string]string
}

type URLNotFoundError struct {
	hash string
}

func (err *URLNotFoundError) Error() string {
	return fmt.Sprintf(`Not found url with hash "%s"`, err.hash)
}

func NewURLStorage() *urlRepo {
	return &urlRepo{make(map[string]string)}
}

func (u *urlRepo) Store(hash string, url string) {
	u.m[hash] = url
}

func (u *urlRepo) Get(hash string) (string, error) {
	url, ok := u.m[hash]
	if !ok {
		return "", &URLNotFoundError{hash}
	}

	return url, nil
}
