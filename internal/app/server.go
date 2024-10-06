package app

import (
	"net/http"

	"github.com/llravell/go-shortener/internal/handlers"
	"github.com/llravell/go-shortener/internal/models"
	"github.com/llravell/go-shortener/internal/storages"
)

func StartServer(addr string) error {
	mux := http.NewServeMux()
	us := storages.NewURLStorage()
	rsg := models.NewRandomStringGenerator()

	saveURLHandler := handlers.SaveURLHandler(us, rsg)
	resolveURLHandler := handlers.ResolveURLHandler(us)

	mux.HandleFunc(`GET /{id}`, resolveURLHandler)
	mux.HandleFunc(`POST /`, saveURLHandler)

	return http.ListenAndServe(addr, mux)
}
