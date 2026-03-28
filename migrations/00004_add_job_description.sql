-- +goose Up
ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS job_description TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE jobs
    DROP COLUMN IF EXISTS job_description;
