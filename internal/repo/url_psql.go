package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/llravell/go-shortener/internal/entity"
)

type URLDatabaseRepo struct {
	conn *sql.DB
}

var ErrOriginalURLConflict = errors.New("url already exists")

func NewURLDatabaseRepo(conn *sql.DB) *URLDatabaseRepo {
	return &URLDatabaseRepo{conn: conn}
}

func (r *URLDatabaseRepo) getNullableUserUUID(url *entity.URL) sql.NullString {
	return sql.NullString{
		String: url.UserUUID,
		Valid:  url.UserUUID != "",
	}
}

func (r *URLDatabaseRepo) Store(ctx context.Context, url *entity.URL) (*entity.URL, error) {
	storedURL, err := r.getByOriginalURL(ctx, url.Original)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if storedURL != nil {
		return storedURL, ErrOriginalURLConflict
	}

	row := r.conn.QueryRowContext(ctx, `
		INSERT INTO urls (url, short, user_uuid)
		VALUES
			($1, $2, $3)
		RETURNING uuid, url, short;
	`, url.Original, url.Short, r.getNullableUserUUID(url))

	var returnedURL entity.URL

	err = row.Scan(&returnedURL.UUID, &returnedURL.Original, &returnedURL.Short)
	if err != nil {
		return nil, err
	}

	return &returnedURL, nil
}

func (r *URLDatabaseRepo) StoreMultipleURLs(ctx context.Context, urls []*entity.URL) error {
	queryTemplateBase := 3
	queryTemplateParams := make([]string, len(urls))

	for i := range urls {
		offset := i * queryTemplateBase

		//nolint
		queryTemplateParams[i] = fmt.Sprintf("($%d,$%d,$%d)", offset+1, offset+2, offset+3)
	}

	args := make([]any, 0, len(urls))
	for _, url := range urls {
		args = append(args, url.Original, url.Short, r.getNullableUserUUID(url))
	}

	//nolint:gosec
	query := `
		INSERT INTO urls (url, short, user_uuid)
			VALUES
	` + " " + strings.Join(queryTemplateParams, ",") + ";"

	_, err := r.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *URLDatabaseRepo) GetURL(ctx context.Context, hash string) (*entity.URL, error) {
	row := r.conn.QueryRowContext(
		ctx,
		"SELECT uuid, url, short, is_deleted FROM urls WHERE short=$1",
		hash,
	)

	var url entity.URL

	err := row.Scan(&url.UUID, &url.Original, &url.Short, &url.Deleted)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLDatabaseRepo) GetUserURLS(ctx context.Context, userUUID string) ([]*entity.URL, error) {
	urls := make([]*entity.URL, 0)

	rows, err := r.conn.QueryContext(
		ctx,
		"SELECT uuid, url, short FROM urls WHERE user_uuid=$1 AND NOT is_deleted",
		userUUID,
	)
	if err != nil {
		return urls, err
	}

	defer rows.Close()

	for rows.Next() {
		var url entity.URL

		err = rows.Scan(&url.UUID, &url.Original, &url.Short)
		if err != nil {
			return urls, err
		}

		urls = append(urls, &url)
	}

	err = rows.Err()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return urls, nil
		}

		return urls, err
	}

	return urls, nil
}

func (r *URLDatabaseRepo) getByOriginalURL(ctx context.Context, originalURL string) (*entity.URL, error) {
	row := r.conn.QueryRowContext(
		ctx,
		"SELECT uuid, url, short FROM urls WHERE url=$1 AND NOT is_deleted",
		originalURL,
	)

	var url entity.URL

	err := row.Scan(&url.UUID, &url.Original, &url.Short)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (r *URLDatabaseRepo) DeleteMultipleURLs(ctx context.Context, userUUID string, urlHashes []string) error {
	queryTemplateBase := 2
	queryTemplateParams := make([]string, len(urlHashes))

	for i := range urlHashes {
		queryTemplateParams[i] = fmt.Sprintf("$%d", queryTemplateBase+i)
	}

	args := make([]any, 0, len(urlHashes)+1)
	args = append(args, userUUID)

	for _, hash := range urlHashes {
		args = append(args, hash)
	}

	//nolint:gosec
	query := `
		UPDATE urls
		SET is_deleted=TRUE
		WHERE user_uuid=$1 AND short IN
	` + " (" + strings.Join(queryTemplateParams, ",") + ");"

	_, err := r.conn.ExecContext(ctx, query, args...)

	return err
}
