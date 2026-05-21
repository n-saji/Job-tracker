-- +goose Up
ALTER TYPE discard_reason ADD VALUE IF NOT EXISTS 'sponsorship';
ALTER TYPE discard_reason ADD VALUE IF NOT EXISTS 'primary_stack_mismatch';

-- +goose Down

