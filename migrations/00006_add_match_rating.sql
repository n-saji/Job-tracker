-- +goose Up
ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS match_rating DOUBLE PRECISION NULL;

ALTER TABLE jobs
DROP CONSTRAINT IF EXISTS jobs_match_rating_range;

ALTER TABLE jobs
ADD CONSTRAINT jobs_match_rating_range
CHECK (match_rating IS NULL OR (match_rating >= 0 AND match_rating <= 10));

-- +goose Down
ALTER TABLE jobs
DROP CONSTRAINT IF EXISTS jobs_match_rating_range;

ALTER TABLE jobs
DROP COLUMN IF EXISTS match_rating;
