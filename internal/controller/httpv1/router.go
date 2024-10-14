package httpv1

import (
	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/usecase"
)

func NewRouter(u *usecase.URLUseCase, baseAddr string) chi.Router {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		newURLRoutes(r, u, baseAddr)
	})

	return r
}
