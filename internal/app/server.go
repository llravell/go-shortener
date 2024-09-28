package app

import (
	"fmt"
	"io"
	"net/http"
)

const HASH = `EwHXdJfB`

var urlMap = make(map[string]string)

func createShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	url, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	urlMap[HASH] = string(url)

	w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", HASH)))
	w.WriteHeader(http.StatusCreated)
}

func resolveShortUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	id := r.PathValue(`id`)
	fmt.Println(id)

	url, ok := urlMap[id]
	if !ok {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	fmt.Println(url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func StartServer(addr string) error {
	mux := http.NewServeMux()

	mux.HandleFunc(`/{id}`, resolveShortUrlHandler)
	mux.HandleFunc(`/`, createShortUrlHandler)

	return http.ListenAndServe(addr, mux)
}
