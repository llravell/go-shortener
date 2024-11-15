package repo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLDatabaseRepo struct {
	conn *sql.DB
}

var ErrOriginalURLConflict = errors.New("url already exists")

func NewURLDatabaseRepo(conn *sql.DB) *URLDatabaseRepo {
	return &URLDatabaseRepo{conn: conn}
}

func (u *URLDatabaseRepo) Store(ctx context.Context, url *entity.URL) (*entity.URL, error) {
	storedURL, err := u.getByOriginalURL(ctx, url.Original)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if storedURL != nil {
		return storedURL, ErrOriginalURLConflict
	}

	row := u.conn.QueryRowContext(ctx, `
		INSERT INTO urls (url, short)
		VALUES
			($1, $2)
		RETURNING uuid, url, short;
	`, url.Original, url.Short)

	var returnedURL entity.URL

	err = row.Scan(&returnedURL.UUID, &returnedURL.Original, &returnedURL.Short)
	if err != nil {
		return nil, err
	}

	return &returnedURL, nil
}

func (u *URLDatabaseRepo) StoreMultiple(ctx context.Context, urls []*entity.URL) error {
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

func (u *URLDatabaseRepo) Get(ctx context.Context, hash string) (*entity.URL, error) {
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

func (u *URLDatabaseRepo) getByOriginalURL(ctx context.Context, originalURL string) (*entity.URL, error) {
	row := u.conn.QueryRowContext(
		ctx,
		"SELECT uuid, url, short FROM urls WHERE url=$1",
		originalURL,
	)

	var url entity.URL

	err := row.Scan(&url.UUID, &url.Original, &url.Short)
	if err != nil {
		return nil, err
	}

	return &url, nil
}
