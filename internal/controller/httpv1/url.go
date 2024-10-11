package httpv1

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/usecase"
)

type urlRoutes struct {
	u        *usecase.URLUseCase
	baseAddr string
}

func newURLRoutes(r chi.Router, u *usecase.URLUseCase, baseAddr string) {
	routes := &urlRoutes{u, baseAddr}

	r.Get("/{id}", routes.resolveURL)
	r.Post("/", routes.saveURL)
}

func (ur *urlRoutes) saveURL(w http.ResponseWriter, r *http.Request) {
	res, err := io.ReadAll(r.Body)
	url := string(res)
	if err != nil || url == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hash := ur.u.SaveURL(url)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("%s/%s", ur.baseAddr, hash)))
}

func (ur *urlRoutes) resolveURL(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue(`id`)

	url, err := ur.u.ResolveURL(hash)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
