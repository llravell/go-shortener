-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX idx_urls_url
ON urls(url);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX idx_urls_url;
-- +goose StatementEnd
