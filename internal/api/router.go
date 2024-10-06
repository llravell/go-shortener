package api

import (
	"github.com/go-chi/chi/v5"
)

func BuildRouter(s UrlStorage, hg HashGenerator, baseAddr string) chi.Router {
	saveUrlHandler := saveUrlHandler(s, hg, baseAddr)
	resolveUrlHandler := resolveUrlHandler(s)

	r := chi.NewRouter()

	r.Get("/{id}", resolveUrlHandler)
	r.Post("/", saveUrlHandler)

	return r
}
