-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
  uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  url VARCHAR(2048) NOT NULL,
  short VARCHAR(50) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE urls;
-- +goose StatementEnd
