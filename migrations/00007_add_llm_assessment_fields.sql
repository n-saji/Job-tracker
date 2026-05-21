-- +goose Up
ALTER TABLE jobs
ADD COLUMN IF NOT EXISTS verdict TEXT NULL,
ADD COLUMN IF NOT EXISTS total_score INTEGER NULL,
ADD COLUMN IF NOT EXISTS reject_reason TEXT NULL,
ADD COLUMN IF NOT EXISTS section_scores JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS extracted JSONB NOT NULL DEFAULT '{}'::jsonb,
ADD COLUMN IF NOT EXISTS flags JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE jobs
DROP CONSTRAINT IF EXISTS jobs_total_score_range;

ALTER TABLE jobs
ADD CONSTRAINT jobs_total_score_range
CHECK (total_score IS NULL OR (total_score >= 0 AND total_score <= 100));

CREATE INDEX IF NOT EXISTS idx_jobs_verdict_active ON jobs(verdict) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_total_score_active ON jobs(total_score DESC) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_section_skills_match_active ON jobs (((section_scores ->> 'skills_match')::int)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_section_years_of_experience_active ON jobs (((section_scores ->> 'years_of_experience')::int)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_section_location_active ON jobs (((section_scores ->> 'location')::int)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_section_title_alignment_active ON jobs (((section_scores ->> 'title_alignment')::int)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_section_employment_type_active ON jobs (((section_scores ->> 'employment_type')::int)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_jobs_section_domain_relevance_active ON jobs (((section_scores ->> 'domain_relevance')::int)) WHERE deleted_at IS NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_jobs_section_domain_relevance_active;
DROP INDEX IF EXISTS idx_jobs_section_employment_type_active;
DROP INDEX IF EXISTS idx_jobs_section_title_alignment_active;
DROP INDEX IF EXISTS idx_jobs_section_location_active;
DROP INDEX IF EXISTS idx_jobs_section_years_of_experience_active;
DROP INDEX IF EXISTS idx_jobs_section_skills_match_active;
DROP INDEX IF EXISTS idx_jobs_total_score_active;
DROP INDEX IF EXISTS idx_jobs_verdict_active;

ALTER TABLE jobs
DROP CONSTRAINT IF EXISTS jobs_total_score_range;

ALTER TABLE jobs
DROP COLUMN IF EXISTS flags,
DROP COLUMN IF EXISTS extracted,
DROP COLUMN IF EXISTS section_scores,
DROP COLUMN IF EXISTS reject_reason,
DROP COLUMN IF EXISTS total_score,
DROP COLUMN IF EXISTS verdict;