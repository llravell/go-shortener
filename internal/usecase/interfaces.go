package usecase

type URLRepo interface {
	Store(hash string, url string)
	Get(hash string) (string, error)
}

type HashGenerator interface {
	Generate(url string) string
}
