package handlers

import (
	"fmt"
	"io"
	"net/http"
)

type urlStorage interface {
	Save(hash string, url string)
	Get(hash string) (string, error)
}

type hashGenerator interface {
	Generate(len int) string
}

func SaveUrlHandler(s urlStorage, hg hashGenerator) http.HandlerFunc {
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
		w.Write([]byte(fmt.Sprintf("http://localhost:8080/%s", hash)))
	}
}

func ResolveUrlHandler(s urlStorage) http.HandlerFunc {
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
