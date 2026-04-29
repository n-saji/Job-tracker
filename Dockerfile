# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal ./internal
COPY migrations ./migrations

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/job-tracker-api ./main.go

FROM alpine:3.22
WORKDIR /app

# Runtime dependencies for HTTPS/cert validation.
RUN apk add --no-cache ca-certificates

# Match app relative paths used by config and migrations.
COPY --from=builder /out/job-tracker-api /app/job-tracker-api
COPY --from=builder /src/migrations /app/migrations

RUN touch /app/.env

WORKDIR /app
EXPOSE 8000

CMD ["./job-tracker-api"]
