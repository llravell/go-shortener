-- +goose Up
-- +goose StatementBegin
ALTER TABLE urls
ADD user_uuid UUID DEFAULT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE urls
DROP COLUMN user_uuid;
-- +goose StatementEnd
