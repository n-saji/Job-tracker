# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/job-tracker-api ./cmd/api

FROM alpine:3.22
WORKDIR /app

# Runtime dependencies for HTTPS/cert validation.
RUN apk add --no-cache ca-certificates

# Match app relative paths used by config and migrations.
COPY --from=builder /out/job-tracker-api /app/cmd/api/job-tracker-api
COPY --from=builder /src/migrations /app/migrations

# App expects ../../.env relative to /app/cmd/api.
RUN mkdir -p /app/cmd/api && touch /app/.env

WORKDIR /app/cmd/api
EXPOSE 8000

CMD ["./job-tracker-api"]
