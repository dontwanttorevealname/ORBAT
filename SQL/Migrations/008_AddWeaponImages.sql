-- +goose Up
ALTER TABLE weapons ADD COLUMN image_url TEXT;

-- +goose Down
ALTER TABLE weapons DROP COLUMN image_url;