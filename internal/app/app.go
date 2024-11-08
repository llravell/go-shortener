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
		log.Error().Err(err).Msg("Backup initialize failed")
		panic(err)
	}

	urls, err := backup.Restore()
	if err != nil {
		log.Error().Err(err).Msg("Backup restore failed")
	}

	memoRepo.Init(urls)

	return func() error {
		err := backup.Store(memoRepo.GetList())
		if err != nil {
			log.Error().Err(err).Msg("Backup store failed")
		}

		return backup.Close()
	}
}

//nolint:funlen
func Run(cfg *config.Config) {
	log := logger.Get()
	defer logger.Close()

	var db *sql.DB

	var err error

	var urlRepo usecase.URLRepo

	if cfg.DatabaseDsn != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			log.Error().Err(err).Msg("Database connection failed")
			panic(err)
		}

		urlRepo = repo.NewURLPsqlRepo(db)
		defer db.Close()
	} else {
		memoRepo := repo.NewURLMemoRepo()
		cancel := prepareMemoryURLRepo(memoRepo, cfg, log)
		urlRepo = memoRepo

		defer func() {
			err = cancel()
		}()
	}

	healthUseCase := usecase.NewHealthUseCase(db)

	urlUseCase := usecase.NewURLUseCase(
		urlRepo,
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
