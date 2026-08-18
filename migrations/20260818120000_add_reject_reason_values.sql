-- +goose Up
ALTER TYPE discard_reason ADD VALUE IF NOT EXISTS 'seniority_mismatch';
ALTER TYPE discard_reason ADD VALUE IF NOT EXISTS 'non_technical';
ALTER TYPE discard_reason ADD VALUE IF NOT EXISTS 'employment_type';
ALTER TYPE discard_reason ADD VALUE IF NOT EXISTS 'empty_jd';

ALTER TABLE jobs
    ALTER COLUMN reject_reason TYPE discard_reason USING reject_reason::discard_reason;

-- +goose Down
ALTER TABLE jobs
    ALTER COLUMN reject_reason TYPE TEXT USING reject_reason::TEXT;
