-- +goose Up
ALTER TABLE groups ADD COLUMN country_code TEXT;

-- +goose Down
ALTER TABLE groups DROP COLUMN country_code; 