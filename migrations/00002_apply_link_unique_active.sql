-- +goose Up
CREATE UNIQUE INDEX IF NOT EXISTS uq_jobs_apply_link_active
ON jobs(apply_link)
WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS uq_jobs_apply_link_active;
