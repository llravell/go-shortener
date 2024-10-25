package httpv1

import (
	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

func NewRouter(u *usecase.URLUseCase, l zerolog.Logger, baseAddr string) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware(l))
	r.Use(middleware.CompressMiddleware())

	newURLRoutes(r, u, l, baseAddr)

	return r
}
