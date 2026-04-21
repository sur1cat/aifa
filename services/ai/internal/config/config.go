package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Port  string
	Debug bool

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	JWTSecret string

	OpenAIAPIKey string
	OpenAIModel  string
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
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		RedisPassword: env("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,
		JWTSecret:     secret,
		OpenAIAPIKey:  env("OPENAI_API_KEY", ""),
		OpenAIModel:   env("OPENAI_MODEL", "gpt-4o-mini"),
	}, nil
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
