package rest

import (
	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/rest/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

type Router struct {
	Mux           chi.Router
	urlUseCase    *usecase.URLUseCase
	healthUseCase *usecase.HealthUseCase
	jwtSecret     string
	log           zerolog.Logger
}

func NewRouter(
	urlUseCase *usecase.URLUseCase,
	healthUseCase *usecase.HealthUseCase,
	jwtSecret string,
	log zerolog.Logger,
) *Router {
	router := &Router{
		Mux:           chi.NewRouter(),
		urlUseCase:    urlUseCase,
		healthUseCase: healthUseCase,
		jwtSecret:     jwtSecret,
		log:           log,
	}

	router.Mux.Use(middleware.LoggerMiddleware(log))

	auth := middleware.NewAuth(jwtSecret, log)

	healthRoutes := NewHealthRoutes(healthUseCase, log)
	urlRoutes := NewURLRoutes(urlUseCase, auth, log)

	healthRoutes.Apply(router.Mux)
	urlRoutes.Apply(router.Mux)

	return router
}
