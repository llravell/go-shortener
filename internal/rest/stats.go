package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/llravell/go-shortener/internal/entity"
	"github.com/rs/zerolog"
)

// StatsUseCase юзкейс внутренней статистики приложения.
type StatsUseCase interface {
	GetStats(ctx context.Context) (*entity.Stats, error)
}

// StatsRoutes роуты внутренней статистики приложения.
type StatsRoutes struct {
	statsUC StatsUseCase
	log     *zerolog.Logger
}

// NewStatsRoutes создает роуты.
func NewStatsRoutes(statsUC StatsUseCase, log *zerolog.Logger) *StatsRoutes {
	return &StatsRoutes{
		statsUC: statsUC,
		log:     log,
	}
}

func (sr *StatsRoutes) stats(w http.ResponseWriter, r *http.Request) {
	stats, err := sr.statsUC.GetStats(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(stats)
	if err != nil {
		sr.log.Err(err).Msg("response write has been failed")
	}
}

// Apply добавляет роуты к роутеру.
func (sr *StatsRoutes) Apply(r chi.Router) {
	r.Get("/stats", sr.stats)
}
