package main

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/llravell/go-shortener/config"
	"github.com/llravell/go-shortener/internal/app"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	app.Run(cfg)
}
