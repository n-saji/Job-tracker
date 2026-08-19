-- +goose Up
ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS company_size TEXT;

-- +goose Down
ALTER TABLE jobs
    DROP COLUMN IF EXISTS company_size;
