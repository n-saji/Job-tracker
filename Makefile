APP_NAME=job-tracker-api
GOOSE=go run github.com/pressly/goose/v3/cmd/goose@latest

.PHONY: run build test migrate-up migrate-down migrate-status

run:
	go run ./cmd/api

build:
	go build ./...

test:
	go test ./...

migrate-up:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" down

migrate-status:
	$(GOOSE) -dir migrations postgres "$(DATABASE_URL)" status
