package rest

import (
	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/rest/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

func NewRouter(
	urlUseCase *usecase.URLUseCase,
	urlDeleteUseCase *usecase.URLDeleteUseCase,
	healthUseCase *usecase.HealthUseCase,
	jwtSecret string,
	log zerolog.Logger,
) *chi.Mux {
	router := chi.NewRouter()

	router.Use(middleware.LoggerMiddleware(log))

	auth := middleware.NewAuth(jwtSecret)

	NewHealthRoutes(router, healthUseCase, log)
	NewURLRoutes(router, urlUseCase, urlDeleteUseCase, auth, log)

	return router
}
