package storages

type urlStorage struct {
	m map[string]string
}

func NewUrlStorage() *urlStorage {
	return &urlStorage{make(map[string]string)}
}

func (u *urlStorage) Save(hash string, url string) {
	u.m[hash] = url
}

func (u *urlStorage) Get(hash string) string {
	url, ok := u.m[hash]
	if !ok {
		return ""
	}

	return url
}
