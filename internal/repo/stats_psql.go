package repo

import (
	"context"
	"database/sql"
)

// StatsDatabaseRepo репозиторий для получения статистики из базы данных.
type StatsDatabaseRepo struct {
	conn *sql.DB
}

// NewStatsDatabaseRepo создает репозиторий.
func NewStatsDatabaseRepo(conn *sql.DB) *StatsDatabaseRepo {
	return &StatsDatabaseRepo{conn: conn}
}

// GetURLsAmount возвращает количество сокращенных урлов.
func (r *StatsDatabaseRepo) GetURLsAmount(ctx context.Context) (int, error) {
	var amount int

	row := r.conn.QueryRowContext(ctx, "SELECT count(*) FROM urls;")

	err := row.Scan(&amount)
	if err != nil {
		return 0, err
	}

	return amount, nil
}

// GetUsersAmount возвращает количество пользователей.
func (r *StatsDatabaseRepo) GetUsersAmount(ctx context.Context) (int, error) {
	var amount int

	row := r.conn.QueryRowContext(ctx, `
		SELECT count(DISTINCT user_uuid)
		FROM urls
		WHERE user_uuid IS NOT NULL;
	`)

	err := row.Scan(&amount)
	if err != nil {
		return 0, err
	}

	return amount, nil
}
