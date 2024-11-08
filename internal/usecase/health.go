package usecase

import (
	"context"
	"errors"
	"reflect"
	"time"
)

const pintTimeout = time.Second * 30

var ErrHasNotConnection = errors.New("has not db connection")

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
	v := reflect.ValueOf(h.s)
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return ErrHasNotConnection
	}

	ctx, cancel := context.WithTimeout(ctx, pintTimeout)
	defer cancel()

	return h.s.PingContext(ctx)
}
