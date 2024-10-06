package api

import (
	"fmt"
	"io"
	"net/http"
)

type UrlStorage interface {
	Save(hash string, url string)
	Get(hash string) (string, error)
}

type HashGenerator interface {
	Generate(len int) string
}

func saveUrlHandler(s UrlStorage, hg HashGenerator, baseAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := io.ReadAll(r.Body)
		url := string(res)
		if err != nil || url == "" {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		hash := hg.Generate(10)
		s.Save(hash, string(url))

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(fmt.Sprintf("%s/%s", baseAddr, hash)))
	}
}

func resolveUrlHandler(s UrlStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hash := r.PathValue(`id`)

		url, err := s.Get(hash)
		if err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
