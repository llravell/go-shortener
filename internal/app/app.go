package app

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/llravell/go-shortener/config"
	"github.com/llravell/go-shortener/internal/controller/httpv1"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/llravell/go-shortener/internal/usecase/repo"
	"github.com/llravell/go-shortener/logger"
)

func Run(cfg *config.Config) {
	log := logger.Get()

	urlStorage := repo.NewURLStorage()
	backup, err := repo.NewURLBackup(cfg.FileStoragePath)

	if err != nil {
		log.Error().Err(err).Msg("Backup initialize failed")
	} else {
		urls, err := backup.Restore()
		if err != nil {
			log.Error().Err(err).Msg("Backup restore failed")
		}

		urlStorage.Init(urls)
	}

	urlUseCase := usecase.NewURLUseCase(
		urlStorage,
		entity.NewRandomStringGenerator(),
	)

	router := httpv1.NewRouter(
		urlUseCase,
		log,
		cfg.BaseAddr,
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	serverNorify := make(chan error, 1)
	go func() {
		serverNorify <- http.ListenAndServe(cfg.Addr, router)
		close(serverNorify)
	}()

	log.Info().
		Str("addr", cfg.Addr).
		Msgf("Starting shortener server on '%s'", cfg.Addr)

	select {
	case s := <-interrupt:
		log.Info().Str("signal", s.String()).Msg("interrupt")
	case err = <-serverNorify:
		log.Error().Err(err).Msg("Shortener server has been closed")
	}

	if backup != nil {
		err = backup.Store(urlStorage.GetList())
		if err != nil {
			log.Error().Err(err).Msg("Backup store failed")
		}

		err = backup.Close()
		if err != nil {
			log.Error().Err(err).Msg("Backup close failed")
		}
	}

	logger.Shutdown()
}
