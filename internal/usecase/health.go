package usecase

import (
	"context"
	"time"
)

const pintTimeout = time.Second * 30

type storage interface {
	PingContext(ctx context.Context) error
}

type HealthUseCase struct {
	s storage
}

func NewHealthUseCase(s storage) *HealthUseCase {
	return &HealthUseCase{s}
}

func (h HealthUseCase) PingContext(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pintTimeout)
	defer cancel()

	return h.s.PingContext(ctx)
}
