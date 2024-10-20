package app

import (
	"net/http"

	"github.com/llravell/go-shortener/config"
	"github.com/llravell/go-shortener/internal/controller/httpv1"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/llravell/go-shortener/internal/usecase/repo"
	"github.com/llravell/go-shortener/logger"
)

func Run(cfg *config.Config) {
	log := logger.Get()

	urlUseCase := usecase.NewURLUseCase(
		repo.NewURLStorage(),
		entity.NewRandomStringGenerator(),
	)

	router := httpv1.NewRouter(
		urlUseCase,
		log,
		cfg.BaseAddr,
	)

	log.Info().
		Str("addr", cfg.Addr).
		Msgf("Starting shortener server on '%s'", cfg.Addr)

	log.Fatal().
		Err(http.ListenAndServe(cfg.Addr, router)).
		Msg("Shortener server has been closed")

	logger.Shutdown()
}
