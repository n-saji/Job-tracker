-- +goose NO TRANSACTION
-- +goose Up
ALTER TYPE job_status ADD VALUE IF NOT EXISTS 'discarded';

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'discard_reason') THEN
        CREATE TYPE discard_reason AS ENUM (
            'high_applicants',
            'security_clearance',
            'less_experience',
            'citizenship',
            'not_fit'
        );
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE jobs
    ADD COLUMN IF NOT EXISTS discard_reason discard_reason NULL;

ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS chk_jobs_discard_reason;

ALTER TABLE jobs
    ADD CONSTRAINT chk_jobs_discard_reason
    CHECK (
        (status = 'discarded' AND discard_reason IS NOT NULL)
        OR
        (status <> 'discarded' AND discard_reason IS NULL)
    );

-- +goose Down
ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS chk_jobs_discard_reason;

ALTER TABLE jobs
    DROP COLUMN IF EXISTS discard_reason;

DROP TYPE IF EXISTS discard_reason;