-- +goose Up
CREATE TABLE IF NOT EXISTS resume_generation_queue (
    job_id UUID PRIMARY KEY REFERENCES jobs(id) ON DELETE CASCADE,
    apply_link TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'added',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_resume_generation_queue_apply_link
ON resume_generation_queue (apply_link);

CREATE INDEX IF NOT EXISTS idx_resume_generation_queue_status
ON resume_generation_queue (status);

-- +goose Down
DROP TABLE IF EXISTS resume_generation_queue;
