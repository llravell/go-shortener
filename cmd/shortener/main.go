package main

import (
	"database/sql"
	"embed"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/llravell/go-shortener/config"
	"github.com/llravell/go-shortener/internal/app"
	"github.com/pressly/goose/v3"
)

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

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	var db *sql.DB

	if cfg.DatabaseDsn != "" {
		db, err = sql.Open("pgx", cfg.DatabaseDsn)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		runMigrations(db)
	}

	app.Run(cfg, db)
}
