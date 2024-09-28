package app

import (
	"io"
	"net/http"
)

func createShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	_, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func resolveShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	id := r.PathValue(`id`)
	if id == `` {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, `https://practicum.yandex.ru/`, http.StatusTemporaryRedirect)
}

func StartServer(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc(`/{id}`, resolveShortUrlHandler)
	mux.HandleFunc(`/`, createShortUrlHandler)

	return http.ListenAndServe(addr, mux)
}
