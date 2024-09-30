package app

import (
	"net/http"

	"github.com/llravell/go-shortener/internal/handlers"
	"github.com/llravell/go-shortener/internal/models"
	"github.com/llravell/go-shortener/internal/storages"
)

func StartServer(addr string) error {
	mux := http.NewServeMux()
	us := storages.NewUrlStorage()
	rsg := models.NewRandomStringGenerator()

	saveUrlHandler := handlers.SaveUrlHandler(us, rsg)
	resolveUrlHandler := handlers.ResolveUrlHandler(us)

	mux.HandleFunc(`GET /{id}`, resolveUrlHandler)
	mux.HandleFunc(`POST /`, saveUrlHandler)

	return http.ListenAndServe(addr, mux)
}
