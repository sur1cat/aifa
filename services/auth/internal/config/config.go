package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port        string
	Debug       bool
	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	NATSURL string

	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration

	GoogleClientID string
	AppleClientID  string
}

const defaultJWTSecret = "change-me-in-production-min-32-chars-long-secret"

func Load() (*Config, error) {
	debug := env("DEBUG", "false") == "true"

	secret := env("JWT_SECRET", defaultJWTSecret)
	if !debug && (secret == "" || secret == defaultJWTSecret) {
		return nil, errors.New("JWT_SECRET must be set in production")
	}

	redisDB, err := strconv.Atoi(env("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("REDIS_DB: %w", err)
	}

	return &Config{
		Port:           env("PORT", "8080"),
		Debug:          debug,
		DatabaseURL:    env("DATABASE_URL", "postgres://aifa:aifa@localhost:5432/aifa?sslmode=disable&search_path=auth,public"),
		RedisAddr:      env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  env("REDIS_PASSWORD", ""),
		RedisDB:        redisDB,
		NATSURL:        env("NATS_URL", "nats://localhost:4222"),
		JWTSecret:      secret,
		AccessTTL:      envDuration("JWT_ACCESS_TTL_HOURS", 720*time.Hour),
		RefreshTTL:     envDuration("JWT_REFRESH_TTL_HOURS", 8760*time.Hour),
		GoogleClientID: env("GOOGLE_CLIENT_ID", ""),
		AppleClientID:  env("APPLE_CLIENT_ID", "app.aifa.ios"),
	}, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	hours, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return time.Duration(hours) * time.Hour
}
