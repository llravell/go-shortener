package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/handlers"
	"github.com/llravell/go-shortener/internal/models"
	"github.com/llravell/go-shortener/internal/storages"
)

func buildRouter(s handlers.URLStorage, hg handlers.HashGenerator) chi.Router {
	saveURLHandler := handlers.SaveURLHandler(s, hg)
	resolveURLHandler := handlers.ResolveURLHandler(s)

	r := chi.NewRouter()

	r.Get("/{id}", resolveURLHandler)
	r.Post("/", saveURLHandler)

	return r
}

func StartServer(addr string) error {
	us := storages.NewURLStorage()
	rsg := models.NewRandomStringGenerator()

	r := buildRouter(us, rsg)

	return http.ListenAndServe(addr, r)
}
