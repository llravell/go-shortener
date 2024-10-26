package httpv1

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

type urlRoutes struct {
	u        *usecase.URLUseCase
	log      zerolog.Logger
	baseAddr string
}

type saveURLRequest struct {
	URL string `json:"url"`
}

type saveURLResponse struct {
	Result string `json:"result"`
}

func newURLRoutes(r chi.Router, u *usecase.URLUseCase, l zerolog.Logger, baseAddr string) {
	routes := &urlRoutes{u, l, baseAddr}

	r.Get("/{id}", routes.resolveURL)
	r.Post("/", routes.saveURLLegacy)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.CompressMiddleware())

		r.Post("/shorten", routes.saveURL)
	})
}

func (ur *urlRoutes) saveURLLegacy(w http.ResponseWriter, r *http.Request) {
	body := r.Body

	contentEncoding := r.Header.Get("Content-Encoding")
	sendsGzip := strings.Contains(contentEncoding, "gzip")
	ur.log.Info().Str("encoding", contentEncoding).Msg("[legacy] url encoding check")

	if sendsGzip {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "gzip decoding error", http.StatusInternalServerError)
			return
		}
		body = gz
	}

	res, err := io.ReadAll(body)
	url := string(res)
	if err != nil || url == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hash := ur.u.SaveURL(url)

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(fmt.Sprintf("%s/%s", ur.baseAddr, hash)))
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")
	}
}

func (ur *urlRoutes) saveURL(w http.ResponseWriter, r *http.Request) {
	body := r.Body

	contentEncoding := r.Header.Get("Content-Encoding")
	sendsGzip := strings.Contains(contentEncoding, "gzip")
	ur.log.Info().Str("encoding", contentEncoding).Msg("url encoding check")

	if sendsGzip {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "gzip decoding error", http.StatusInternalServerError)
			return
		}
		body = gz
	}

	var urlReq saveURLRequest

	if err := json.NewDecoder(body).Decode(&urlReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hash := ur.u.SaveURL(urlReq.URL)

	resp := saveURLResponse{
		Result: fmt.Sprintf("%s/%s", ur.baseAddr, hash),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")
	}
}

func (ur *urlRoutes) resolveURL(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue(`id`)

	url, err := ur.u.ResolveURL(hash)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	ur.log.Info().Str("url", url).Msg("redirect")

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
