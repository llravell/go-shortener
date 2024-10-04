package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/handlers"
	"github.com/llravell/go-shortener/internal/models"
	"github.com/llravell/go-shortener/internal/storages"
)

func buildRouter(s handlers.UrlStorage, hg handlers.HashGenerator) chi.Router {
	saveUrlHandler := handlers.SaveUrlHandler(s, hg)
	resolveUrlHandler := handlers.ResolveUrlHandler(s)

	r := chi.NewRouter()

	r.Get("/{id}", resolveUrlHandler)
	r.Post("/", saveUrlHandler)

	return r
}

func StartServer(addr string) error {
	us := storages.NewUrlStorage()
	rsg := models.NewRandomStringGenerator()

	r := buildRouter(us, rsg)

	return http.ListenAndServe(addr, r)
}
