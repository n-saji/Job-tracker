# Job Tracker Backend (Go)

Layered Go backend for tracking job applications.

## Architecture

- controller: HTTP handlers and routing
- service: business logic and validation
- dao: database access
- dto: request/response payloads
- globals: shared constants and error codes
- db: pgxpool setup

## Requirements

- Go 1.23+
- PostgreSQL

## Setup

1. Copy `.env.example` to `.env` and update values.
2. Export env vars:
   - `export $(cat .env | xargs)`
3. Run migrations:
   - `make migrate-up`
4. Run app:
   - `make run`

## APIs

- `POST /jobs`
- `GET /jobs`
- `GET /jobs/{id}`
- `PUT /jobs/{id}`
- `DELETE /jobs/{id}` (soft delete)
- `GET /jobs/exists?apply_link=...`
- `GET /health`

## Notes

- `apply_link` is unique only for active rows (`deleted_at IS NULL`).
- `location` is required.
- Duplicate `apply_link` returns `409 Conflict`.
