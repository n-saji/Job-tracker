-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM ('applied', 'interview', 'offer', 'rejected', 'withdrawn');
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_name TEXT NOT NULL,
    role_title TEXT NOT NULL,
    location TEXT NOT NULL,
    apply_link TEXT NOT NULL,
    linkedin_job_url TEXT NOT NULL DEFAULT '',
    resume_link TEXT NOT NULL DEFAULT '',
    status job_status NOT NULL,
    salary_text TEXT NOT NULL DEFAULT '',
    is_easy_apply BOOLEAN NOT NULL DEFAULT FALSE,
    applied_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_active ON jobs(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_company_name_active ON jobs(company_name) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_applied_at_active ON jobs(applied_at DESC) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS jobs;
DROP TYPE IF EXISTS job_status;
