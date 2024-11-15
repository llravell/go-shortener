package httpv1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/usecase"
	"github.com/rs/zerolog"
)

type healthRoutes struct {
	h   *usecase.HealthUseCase
	log zerolog.Logger
}

func NewHealthRoutes(r chi.Router, h *usecase.HealthUseCase, l zerolog.Logger) {
	routes := &healthRoutes{h, l}

	r.Get("/ping", routes.ping)
}

func (hr *healthRoutes) ping(w http.ResponseWriter, r *http.Request) {
	err := hr.h.PingContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}
