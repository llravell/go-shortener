package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/llravell/go-shortener/internal/entity"
)

const (
	_queryTimeout     = 10 * time.Second
	_execTimeout      = 20 * time.Second
	_bootstrapTimeout = time.Minute
)

type URLPsqlRepo struct {
	conn *sql.DB
}

func NewURLPsqlRepo(conn *sql.DB) *URLPsqlRepo {
	return &URLPsqlRepo{conn: conn}
}

func (u *URLPsqlRepo) Store(ctx context.Context, url *entity.URL) (*entity.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, _queryTimeout)
	defer cancel()

	_, err := u.conn.ExecContext(ctx, `
		INSERT INTO urls (url, short)
		VALUES
			($1, $2);
	`, url.Original, url.Short)
	if err != nil {
		return nil, err
	}

	return u.Get(ctx, url.Short)
}

func (u *URLPsqlRepo) StoreMultiple(ctx context.Context, urls []*entity.URL) error {
	ctx, cancel := context.WithTimeout(ctx, _execTimeout)
	defer cancel()

	tx, err := u.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		err = tx.Rollback()
	}()

	for _, url := range urls {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO urls (url, short)
			VALUES
				($1, $2);
		`, url.Original, url.Short)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (u *URLPsqlRepo) Get(ctx context.Context, hash string) (*entity.URL, error) {
	ctx, cancel := context.WithTimeout(ctx, _queryTimeout)
	defer cancel()

	row := u.conn.QueryRowContext(
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
