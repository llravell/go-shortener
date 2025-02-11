package rest

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

// HealthUseCase юзкейс пинга приложения.
type HealthUseCase interface {
	PingContext(ctx context.Context) error
}

// HealthRoutes роуты для проверки приложения.
type HealthRoutes struct {
	healthUC HealthUseCase
	log      *zerolog.Logger
}

// NewHealthRoutes создает роуты.
func NewHealthRoutes(healthUC HealthUseCase, log *zerolog.Logger) *HealthRoutes {
	return &HealthRoutes{
		healthUC: healthUC,
		log:      log,
	}
}

func (hr *HealthRoutes) ping(w http.ResponseWriter, r *http.Request) {
	err := hr.healthUC.PingContext(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.WriteHeader(http.StatusOK)
}

// Apply добавляет роуты к роутеру.
func (hr *HealthRoutes) Apply(r chi.Router) {
	r.Get("/ping", hr.ping)
}
