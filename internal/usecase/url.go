package usecase

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

func (uc *URLUseCase) SaveURL(url string) string {
	hash := uc.gen.Generate(url)
	uc.repo.Store(hash, url)

	return hash
}

func (uc *URLUseCase) ResolveURL(hash string) (string, error) {
	return uc.repo.Get(hash)
}
