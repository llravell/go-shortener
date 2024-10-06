package api

import (
	"github.com/go-chi/chi/v5"
)

func BuildRouter(s URLStorage, hg HashGenerator, baseAddr string) chi.Router {
	saveURLHandler := makeSaveURLHandler(s, hg, baseAddr)
	resolveURLHandler := makeResolveURLHandler(s)

	r := chi.NewRouter()

	r.Get("/{id}", resolveURLHandler)
	r.Post("/", saveURLHandler)

	return r
}
