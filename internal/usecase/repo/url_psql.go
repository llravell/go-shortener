package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/llravell/go-shortener/internal/entity"
)

const _queryTimeout = 10 * time.Second

type URLPsqlRepo struct {
	db *sql.DB
}

func NewURLPsqlRepo(db *sql.DB) *URLPsqlRepo {
	return &URLPsqlRepo{db}
}

func (u *URLPsqlRepo) Store(_ *entity.URL) {}

func (u *URLPsqlRepo) GetContext(ctx context.Context, hash string) (*entity.URL, error) {
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
