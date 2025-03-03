-- +goose Up
ALTER TABLE weapons ADD COLUMN image_url TEXT;

-- +goose Down
