package app

import (
	"log"
	"net/http"

	"github.com/llravell/go-shortener/config"
	"github.com/llravell/go-shortener/internal/controller/httpv1"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/llravell/go-shortener/internal/usecase/repo"
)

func Run(cfg *config.Config) {
	urlUseCase := usecase.NewURLUseCase(
		repo.NewURLStorage(),
		entity.NewRandomStringGenerator(),
	)

	router := httpv1.NewRouter(
		urlUseCase,
		cfg.BaseAddr,
	)

	log.Fatal(http.ListenAndServe(cfg.Addr, router))
}
