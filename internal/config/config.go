package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort        string
	DatabaseURL    string
	DBMaxConns     int32
	RequestTimeout time.Duration
	N8NWebhookURL  string
}

func Load() (Config, error) {
	err := godotenv.Load("../../.env") // Load .env file if it exists
	if err != nil {
		return Config{}, fmt.Errorf("failed to load .env file: %w", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	port := getEnvOrDefault("APP_PORT", "8080")
	maxConns := int32(getEnvIntOrDefault("DB_MAX_CONNS", 10))
	timeoutSeconds := getEnvIntOrDefault("REQUEST_TIMEOUT_SECONDS", 5)
	n8nWebhookURL := getEnvOrDefault("N8N_WEBHOOK_URL", "http://localhost:5678/webhook/a260eeb6-7c50-4599-933a-ef3eeb58cafe")

	return Config{
		AppPort:        port,
		DatabaseURL:    dbURL,
		DBMaxConns:     maxConns,
		RequestTimeout: time.Duration(timeoutSeconds) * time.Second,
		N8NWebhookURL:  n8nWebhookURL,
	}, nil
}

func getEnvOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
