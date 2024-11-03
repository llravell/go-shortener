package httpv1

import (
	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

func NewRouter(
	urlUseCase *usecase.URLUseCase,
	healthUseCase *usecase.HealthUseCase,
	log zerolog.Logger,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.LoggerMiddleware(log))

	newHealthRoutes(router, healthUseCase, log)
	newURLRoutes(router, urlUseCase, log)

	return router
}
