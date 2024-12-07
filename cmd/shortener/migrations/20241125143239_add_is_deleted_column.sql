-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls
ADD is_deleted boolean DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls
DROP COLUMN is_deleted;
-- +goose StatementEnd
