package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/llravell/go-shortener/internal/entity"
)

const (
	_queryTimeout     = 10 * time.Second
	_bootstrapTimeout = time.Minute
)

type URLPsqlRepo struct {
	db *sql.DB
}

func NewURLPsqlRepo(db *sql.DB) *URLPsqlRepo {
	return &URLPsqlRepo{db}
}

func (u *URLPsqlRepo) Store(_ *entity.URL) {}

func (u *URLPsqlRepo) Get(ctx context.Context, hash string) (*entity.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, _queryTimeout)
	defer cancel()

	row := u.db.QueryRowContext(
		ctx,
		"SELECT uuid, url, short FROM urls WHERE short=$1",
		hash,
	)

	var url entity.URL

	err := row.Scan(&url.UUID, &url.Original, &url.Short)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (u *URLPsqlRepo) Bootstrap(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, _bootstrapTimeout)
	defer cancel()

	_, err := u.db.ExecContext(ctx, `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		CREATE TABLE IF NOT EXISTS urls (
			uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			url VARCHAR(2048) NOT NULL,
			short VARCHAR(50) UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)

	return err
}
