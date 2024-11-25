package httpv1

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/controller/httpv1/middleware"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

type urlRoutes struct {
	uuc *usecase.URLUseCase
	udc *usecase.URLDeleteUseCase
	log zerolog.Logger
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
	r chi.Router,
	uuc *usecase.URLUseCase,
	udc *usecase.URLDeleteUseCase,
	jwtSecret string,
	l zerolog.Logger,
) {
	routes := &urlRoutes{uuc, udc, l}
	auth := middleware.NewAuth(jwtSecret)

	r.Get("/{id}", routes.resolveURL)
	r.With(middleware.DecompressMiddleware()).
		With(auth.ProvideJWTMiddleware).
		Post("/", routes.saveURLLegacy)

	r.Route("/api", func(r chi.Router) {
		r.Use(middleware.CompressMiddleware("application/json"))
		r.Use(middleware.DecompressMiddleware())

		r.Route("/shorten", func(r chi.Router) {
			r.Use(auth.ProvideJWTMiddleware)

			r.Post("/", routes.saveURL)
			r.Post("/batch", routes.saveURLMultiple)
		})

		r.Route("/user", func(r chi.Router) {
			r.Use(auth.CheckJWTMiddleware)

			r.Route("/urls", func(r chi.Router) {
				r.Get("/", routes.getUserURLS)
				r.Delete("/", routes.deleteUserURLS)
			})
		})
	})
}

func (ur *urlRoutes) getUserUUIDFromRequest(r *http.Request) string {
	v := r.Context().Value(middleware.UserUUIDContextKey)
	userUUID, ok := v.(string)

	if !ok {
		return ""
	}

	return userUUID
}

func (ur *urlRoutes) saveURLLegacy(w http.ResponseWriter, r *http.Request) {
	res, err := io.ReadAll(r.Body)
	url := string(res)

	if err != nil || url == "" {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	urlObj, err := ur.uuc.SaveURL(r.Context(), url, ur.getUserUUIDFromRequest(r))
	if err != nil {
		if errors.Is(err, usecase.ErrURLDuplicate) {
			w.WriteHeader(http.StatusConflict)
		} else {
			http.Error(w, "saving url failed", http.StatusInternalServerError)

			return
		}
	}

	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(ur.uuc.BuildRedirectURL(urlObj)))
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")
	}
}

func (ur *urlRoutes) saveURL(w http.ResponseWriter, r *http.Request) {
	var urlReq saveURLRequest

	if err := json.NewDecoder(r.Body).Decode(&urlReq); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	urlObj, err := ur.uuc.SaveURL(r.Context(), urlReq.URL, ur.getUserUUIDFromRequest(r))
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
		Result: ur.uuc.BuildRedirectURL(urlObj),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")
	}
}

func (ur *urlRoutes) saveURLMultiple(w http.ResponseWriter, r *http.Request) {
	var batchItems []URLBatchRequestItem

	if err := json.NewDecoder(r.Body).Decode(&batchItems); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	urls := make([]string, 0, len(batchItems))
	for _, item := range batchItems {
		urls = append(urls, item.OriginalURL)
	}

	urlObjs, err := ur.uuc.SaveURLMultiple(r.Context(), urls, ur.getUserUUIDFromRequest(r))
	if err != nil {
		http.Error(w, "saving url failed", http.StatusInternalServerError)

		return
	}

	responseItems := make([]URLBatchResponseItem, 0, len(batchItems))

	for i, urlObj := range urlObjs {
		item := URLBatchResponseItem{
			CorrelationID: batchItems[i].CorrelationID,
			ShortURL:      ur.uuc.BuildRedirectURL(urlObj),
		}

		responseItems = append(responseItems, item)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(responseItems)
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")
	}
}

func (ur *urlRoutes) resolveURL(w http.ResponseWriter, r *http.Request) {
	hash := r.PathValue(`id`)

	url, err := ur.uuc.ResolveURL(r.Context(), hash)
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

func (ur *urlRoutes) getUserURLS(w http.ResponseWriter, r *http.Request) {
	userUUID := ur.getUserUUIDFromRequest(r)

	userURLS, err := ur.uuc.GetUserURLS(r.Context(), userUUID)
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
			ShortURL:    ur.uuc.BuildRedirectURL(urlObj),
		}

		responseItems = append(responseItems, item)
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(responseItems)
	if err != nil {
		ur.log.Err(err).Msg("response write has been failed")
	}
}

func (ur *urlRoutes) deleteUserURLS(w http.ResponseWriter, r *http.Request) {
	var urlHashes []string

	if err := json.NewDecoder(r.Body).Decode(&urlHashes); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)

		return
	}

	ur.udc.QueueDelete(&usecase.URLDeleteItem{
		UserUUID: ur.getUserUUIDFromRequest(r),
		Hashes:   urlHashes,
	})

	w.WriteHeader(http.StatusAccepted)
}
