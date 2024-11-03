package app

import (
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/llravell/go-shortener/config"
	"github.com/llravell/go-shortener/internal/controller/httpv1"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/llravell/go-shortener/internal/usecase/repo"
	"github.com/llravell/go-shortener/logger"
)

func startServer(cfg *config.Config, handler http.Handler) error {
	server := http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	return server.ListenAndServe()
}

//nolint:funlen
func Run(cfg *config.Config) {
	log := logger.Get()
	defer logger.Close()

	db, err := sql.Open("pgx", cfg.DatabaseDsn)
	if err != nil {
		log.Error().Err(err).Msg("Database connection failed")
		panic(err)
	}

	defer db.Close()

	urlStorage := repo.NewURLStorage()

	backup, err := repo.NewURLBackup(cfg.FileStoragePath)
	if err != nil {
		log.Error().Err(err).Msg("Backup initialize failed")
		panic(err)
	}

	urls, err := backup.Restore()
	if err != nil {
		log.Error().Err(err).Msg("Backup restore failed")
	}

	urlStorage.Init(urls)

	defer func() {
		err := backup.Store(urlStorage.GetList())
		if err != nil {
			log.Error().Err(err).Msg("Backup restore failed")
		}

		backup.Close()
	}()

	healthUseCase := usecase.NewHealthUseCase(db)

	urlUseCase := usecase.NewURLUseCase(
		urlStorage,
		entity.NewRandomStringGenerator(),
		cfg.BaseAddr,
	)

	router := httpv1.NewRouter(
		urlUseCase,
		healthUseCase,
		log,
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	serverNorify := make(chan error, 1)
	go func() {
		serverNorify <- startServer(cfg, router)
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
}
