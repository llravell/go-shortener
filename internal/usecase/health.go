package usecase

import (
	"context"
	"errors"
	"reflect"
	"time"
)

const pingTimeout = time.Second * 30

var ErrHasNotConnection = errors.New("has not db connection")

type HealthUseCase struct {
	repo HealthRepo
}

func NewHealthUseCase(repo HealthRepo) *HealthUseCase {
	return &HealthUseCase{repo}
}

func (h HealthUseCase) PingContext(ctx context.Context) error {
	v := reflect.ValueOf(h.repo)
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return ErrHasNotConnection
	}

	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	return h.repo.PingContext(ctx)
}
