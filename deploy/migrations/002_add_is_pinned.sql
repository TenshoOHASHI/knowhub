-- +goose Up
ALTER TABLE articles ADD COLUMN is_pinned BOOLEAN NOT NULL DEFAULT 0 AFTER visibility;

-- +goose Down
ALTER TABLE articles DROP COLUMN is_pinned;
