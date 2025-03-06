package usecase

import (
	"context"

	"github.com/llravell/go-shortener/internal/entity"
	"golang.org/x/sync/errgroup"
)

// StatsUseCase юзкейс внутренней статистики приложения.
type StatsUseCase struct {
	repo StatsRepo
}

// NewHealthUseCase создает юзкейс.
func NewStatsUseCase(repo StatsRepo) *StatsUseCase {
	return &StatsUseCase{repo}
}

// GetStats возвращает статистику приложения.
func (s *StatsUseCase) GetStats(ctx context.Context) (*entity.Stats, error) {
	var stats entity.Stats

	group, ctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		urlsAmount, err := s.repo.GetURLsAmount(ctx)
		if err != nil {
			return err
		}

		stats.URLs = urlsAmount

		return nil
	})

	group.Go(func() error {
		usersAmount, err := s.repo.GetUsersAmount(ctx)
		if err != nil {
			return err
		}

		stats.Users = usersAmount

		return nil
	})

	err := group.Wait()
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
