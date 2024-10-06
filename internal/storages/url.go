package storages

import (
	"fmt"
)

type urlStorage struct {
	m map[string]string
}

type URLNotFoundError struct {
	hash string
}

func (err *URLNotFoundError) Error() string {
	return fmt.Sprintf(`Not found url with hash "%s"`, err.hash)
}

func NewURLStorage() *urlStorage {
	return &urlStorage{make(map[string]string)}
}

func (u *urlStorage) Save(hash string, url string) {
	u.m[hash] = url
}

func (u *urlStorage) Get(hash string) (string, error) {
	url, ok := u.m[hash]
	if !ok {
		return "", &URLNotFoundError{hash}
	}

	return url, nil
}
