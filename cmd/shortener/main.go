package main

import (
	"context"
	"database/sql"
	"embed"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/llravell/go-shortener/config"
	"github.com/llravell/go-shortener/internal/app"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/repo"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/llravell/go-shortener/logger"
	"github.com/llravell/go-shortener/pkg/workerpool"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

const urlDeleteWorkersAmount = 4

//go:embed migrations/*.sql
var embedMigrations embed.FS

func runMigrations(db *sql.DB) {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		panic(err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic(err)
	}
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
func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("config error: %s", err)
	}

	var db *sql.DB

	if cfg.DatabaseDsn != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			log.Fatalf("open db error: %s", err)
		}
		defer db.Close()

		runMigrations(db)
	}

	log := logger.Get()
	defer logger.Close()

	var urlRepo usecase.URLRepo

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

	urlDeleteWorkerPool := workerpool.New[*usecase.URLDeleteWork](urlDeleteWorkersAmount)

	urlUseCase := usecase.NewURLUseCase(
		urlRepo,
		entity.NewRandomStringGenerator(),
		cfg.BaseAddr,
	)
	urlDeleteUseCase := usecase.NewURLDeleteUseCase(
		urlRepo,
		urlDeleteWorkerPool,
		log,
	)
	healthUseCase := usecase.NewHealthUseCase(db)

	urlDeleteWorkerPool.ProcessQueue(context.Background())

	defer func() {
		urlDeleteWorkerPool.Close()

		log.Info().Msg("delete worker pool finish waiting...")
		urlDeleteWorkerPool.Wait()
	}()

	app.New(
		urlUseCase,
		urlDeleteUseCase,
		healthUseCase,
		log,
		app.Addr(cfg.Addr),
		app.JWTSecret(cfg.JWTSecret),
	).Run()
}
