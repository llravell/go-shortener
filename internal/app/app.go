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
	"github.com/rs/zerolog"
)

type urlUnionRepo interface {
	usecase.URLRepo
	usecase.URLDeleteRepo
}

func startServer(cfg *config.Config, handler http.Handler) error {
	server := http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	return server.ListenAndServe()
}

func prepareMemoryURLRepo(
	memoRepo *repo.URLMemoRepo,
	cfg *config.Config,
	log zerolog.Logger,
) func() error {
	backup, err := repo.NewURLBackup(cfg.FileStoragePath)
	if err != nil {
		log.Error().Err(err).Msg("backup initialize failed")
		os.Exit(1)
	}

	urls, err := backup.Restore()
	if err != nil {
		log.Error().Err(err).Msg("backup restore failed")
	}

	memoRepo.Init(urls)

	return func() error {
		err := backup.Store(memoRepo.GetList())
		if err != nil {
			log.Error().Err(err).Msg("backup store failed")
		}

		return backup.Close()
	}
}

//nolint:funlen
func Run(cfg *config.Config, db *sql.DB) {
	log := logger.Get()
	defer logger.Close()

	var err error

	var urlRepo urlUnionRepo

	if cfg.DatabaseDsn != "" {
		urlRepo = repo.NewURLDatabaseRepo(db)
	} else {
		memoRepo := repo.NewURLMemoRepo()
		cancel := prepareMemoryURLRepo(memoRepo, cfg, log)
		urlRepo = memoRepo

		defer func() {
			err = cancel()
			if err != nil {
				log.Error().Err(err).Msg("backup cancel failed")
			}
		}()
	}

	healthUseCase := usecase.NewHealthUseCase(db)

	urlUseCase := usecase.NewURLUseCase(
		urlRepo,
		entity.NewRandomStringGenerator(),
		cfg.BaseAddr,
	)
	urlDeleteUseCase := usecase.NewURLDeleteUseCase(
		urlRepo,
		log,
	)

	router := httpv1.NewRouter(
		urlUseCase,
		urlDeleteUseCase,
		healthUseCase,
		cfg.JWTSecret,
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
		Msgf("starting shortener server on '%s'", cfg.Addr)

	go urlDeleteUseCase.ProcessQueue()
	defer urlDeleteUseCase.Close()

	select {
	case s := <-interrupt:
		log.Info().Str("signal", s.String()).Msg("interrupt")
	case err = <-serverNorify:
		log.Error().Err(err).Msg("shortener server has been closed")
	}
}
