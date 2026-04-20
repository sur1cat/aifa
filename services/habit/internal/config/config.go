package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port        string
	Debug       bool
	DatabaseURL string

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	NATSURL string

	JWTSecret string
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
		Port:          env("PORT", "8080"),
		Debug:         debug,
		DatabaseURL:   env("DATABASE_URL", "postgres://aifa:aifa@localhost:5432/aifa?sslmode=disable&search_path=habits,public"),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		NATSURL:       env("NATS_URL", "nats://localhost:4222"),
		JWTSecret:     secret,
	}, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
