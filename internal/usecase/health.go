package usecase

import (
	"context"
	"errors"
	"reflect"
	"time"
)

const pingTimeout = time.Second * 30

// ErrHasNotConnection ошибка отсутствия подключения к БД.
var ErrHasNotConnection = errors.New("has not db connection")

// HealthUseCase юзкейс пинга приложения.
type HealthUseCase struct {
	repo HealthRepo
}

// NewHealthUseCase создает юзкейс.
func NewHealthUseCase(repo HealthRepo) *HealthUseCase {
	return &HealthUseCase{repo}
}

// PingContext проверяет подключение к БД.
func (h HealthUseCase) PingContext(ctx context.Context) error {
	v := reflect.ValueOf(h.repo)
	if !v.IsValid() || (v.Kind() == reflect.Ptr && v.IsNil()) {
		return ErrHasNotConnection
	}

	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	return h.repo.PingContext(ctx)
}
