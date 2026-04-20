package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

var ErrMissingJWTSecret = errors.New("JWT_SECRET environment variable is required in production")

type Config struct {
	Port        string
	Debug       bool
	DatabaseURL string
	JWT         JWTConfig
	OpenAI      OpenAIConfig
}

type OpenAIConfig struct {
	APIKey string
	Model  string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

const defaultJWTSecret = "change-me-in-production-min-32-chars"

func Load() (*Config, error) {
	debug := getEnv("DEBUG", "false") == "true"
	jwtSecret := getEnv("JWT_SECRET", defaultJWTSecret)

	// In production, require a proper JWT secret
	if !debug && (jwtSecret == "" || jwtSecret == defaultJWTSecret) {
		return nil, ErrMissingJWTSecret
	}

	return &Config{
		Port:        getEnv("PORT", "8080"),
		Debug:       debug,
		DatabaseURL: getEnv("DATABASE_URL", "postgres://habitflow:habitflow@localhost:5432/habitflow?sslmode=disable"),
		JWT: JWTConfig{
			Secret:          jwtSecret,
			AccessTokenTTL:  time.Duration(getEnvAsInt("JWT_ACCESS_TTL_DAYS", 30)) * 24 * time.Hour,
			RefreshTokenTTL: time.Duration(getEnvAsInt("JWT_REFRESH_TTL_DAYS", 365)) * 24 * time.Hour,
		},
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
			Model:  getEnv("OPENAI_MODEL", "gpt-4o-mini"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
