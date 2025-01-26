package rest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/llravell/go-shortener/internal/entity"
	"github.com/llravell/go-shortener/internal/rest/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
)

type URLUseCase interface {
	SaveURL(ctx context.Context, url string, userUUID string) (*entity.URL, error)
	SaveURLMultiple(ctx context.Context, urls []string, userUUID string) ([]*entity.URL, error)
	ResolveURL(ctx context.Context, hash string) (*entity.URL, error)
	GetUserURLS(ctx context.Context, userUUID string) ([]*entity.URL, error)
	BuildRedirectURL(url *entity.URL) string
	QueueDelete(item *entity.URLDeleteItem) error
}

type URLRoutes struct {
	urlUC URLUseCase
	auth  *middleware.Auth
	log   *zerolog.Logger
}

type saveURLRequest struct {
	URL string `json:"url"`
}

type saveURLResponse struct {
	Result string `json:"result"`
}

type URLBatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLBatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UserURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewURLRoutes(
	urlUC URLUseCase,
	auth *middleware.Auth,
	log *zerolog.Logger,
) *URLRoutes {
	return &URLRoutes{
		urlUC: urlUC,
		auth:  auth,
		log:   log,
	}
}

func (ur *URLRoutes) getUserUUIDFromRequest(r *http.Request) string {
	v := r.Context().Value(middleware.UserUUIDContextKey)
	userUUID, ok := v.(string)

	if !ok {
		return ""
	}

	return userUUID
}

func (ur *URLRoutes) saveURLLegacy(w http.ResponseWriter, r *http.Request) {
	res, err := io.ReadAll(r.Body)
	url := string(res)

	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	if url == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	userUUID := ur.getUserUUIDFromRequest(r)
	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	statusCode := http.StatusCreated

	urlObj, err := ur.urlUC.SaveURL(r.Context(), url, userUUID)
	if err != nil {
		if errors.Is(err, usecase.ErrURLDuplicate) {
			statusCode = http.StatusConflict
		} else {
			http.Error(w, "saving url failed", http.StatusInternalServerError)

			return
		}
	}

	w.WriteHeader(statusCode)

	_, err = w.Write([]byte(ur.urlUC.BuildRedirectURL(urlObj)))
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")

		return
	}
}

func (ur *URLRoutes) saveURL(w http.ResponseWriter, r *http.Request) {
	var urlReq saveURLRequest

	if err := json.NewDecoder(r.Body).Decode(&urlReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	userUUID := ur.getUserUUIDFromRequest(r)
	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	urlObj, err := ur.urlUC.SaveURL(r.Context(), urlReq.URL, userUUID)
	if err != nil {
		if errors.Is(err, usecase.ErrURLDuplicate) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
		} else {
			http.Error(w, "saving url failed", http.StatusInternalServerError)

			return
		}
	}

	resp := saveURLResponse{
		Result: ur.urlUC.BuildRedirectURL(urlObj),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")

		return
	}
}

func (ur *URLRoutes) saveURLMultiple(w http.ResponseWriter, r *http.Request) {
	var batchItems []URLBatchRequestItem

	if err := json.NewDecoder(r.Body).Decode(&batchItems); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	userUUID := ur.getUserUUIDFromRequest(r)
	if userUUID == "" {
		w.WriteHeader(http.StatusUnauthorized)

		return
	}

	urls := make([]string, 0, len(batchItems))
	for _, item := range batchItems {
		urls = append(urls, item.OriginalURL)
	}

	urlObjs, err := ur.urlUC.SaveURLMultiple(r.Context(), urls, userUUID)
	if err != nil {
		http.Error(w, "saving url failed", http.StatusInternalServerError)

		return
	}

	responseItems := make([]URLBatchResponseItem, 0, len(batchItems))

	for i, urlObj := range urlObjs {
		item := URLBatchResponseItem{
			CorrelationID: batchItems[i].CorrelationID,
			ShortURL:      ur.urlUC.BuildRedirectURL(urlObj),
		}

		responseItems = append(responseItems, item)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(responseItems)
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")

		return
	}
}

func (ur *URLRoutes) resolveURL(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue(`id`)

	url, err := ur.urlUC.ResolveURL(r.Context(), hash)
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	if url.Deleted {
		w.WriteHeader(http.StatusGone)

		return
	}

	ur.log.Info().Str("url", url.Original).Msg("redirect")

	http.Redirect(w, r, url.Original, http.StatusTemporaryRedirect)
}

func (ur *URLRoutes) getUserURLS(w http.ResponseWriter, r *http.Request) {
	userUUID := ur.getUserUUIDFromRequest(r)

	userURLS, err := ur.urlUC.GetUserURLS(r.Context(), userUUID)
	if err != nil {
		http.Error(w, "searching urls failed", http.StatusInternalServerError)

		return
	}

	if len(userURLS) == 0 {
		w.WriteHeader(http.StatusNoContent)

		return
	}

	responseItems := make([]UserURLItem, 0, len(userURLS))

	for _, urlObj := range userURLS {
		item := UserURLItem{
			OriginalURL: urlObj.Original,
			ShortURL:    ur.urlUC.BuildRedirectURL(urlObj),
		}

		responseItems = append(responseItems, item)
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(responseItems)
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")

		return
	}
}

func (ur *URLRoutes) deleteUserURLS(w http.ResponseWriter, r *http.Request) {
	var urlHashes []string

	if err := json.NewDecoder(r.Body).Decode(&urlHashes); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	err := ur.urlUC.QueueDelete(&entity.URLDeleteItem{
		UserUUID: ur.getUserUUIDFromRequest(r),
		Hashes:   urlHashes,
	})
	if err != nil {
		http.Error(w, "delete urls failed", http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (ur *URLRoutes) Apply(r chi.Router) {
	r.Get("/{id}", ur.resolveURL)
	r.With(middleware.DecompressMiddleware()).
		With(ur.auth.ProvideJWTMiddleware).
		Post("/", ur.saveURLLegacy)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.CompressMiddleware("application/json"))
		r.Use(middleware.DecompressMiddleware())

		r.Route("/shorten", func(r chi.Router) {
			r.Use(ur.auth.ProvideJWTMiddleware)

			r.Post("/", ur.saveURL)
			r.Post("/batch", ur.saveURLMultiple)
		})

		r.Route("/user", func(r chi.Router) {
			r.Route("/urls", func(r chi.Router) {
				r.Use(ur.auth.CheckJWTMiddleware)

				r.Get("/", ur.getUserURLS)
				r.Delete("/", ur.deleteUserURLS)
			})
		})
	})
}
