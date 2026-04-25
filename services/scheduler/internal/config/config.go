package config

import (
	"os"
)

type Config struct {
	Debug   bool
	NATSURL string

	// Cron schedules (standard 5-field cron, no seconds).
	// Empty value disables that particular tick.
	RecurringSchedule string
	ReminderSchedule  string

	// HealthAddr is where the worker exposes a tiny /health server so docker
	// healthchecks have something to hit.
	HealthAddr string
}

func Load() *Config {
	return &Config{
		Debug:             env("DEBUG", "false") == "true",
		NATSURL:           env("NATS_URL", "nats://localhost:4222"),
		RecurringSchedule: env("CRON_RECURRING", "*/5 * * * *"),
		ReminderSchedule:  env("CRON_REMINDERS", "0 9 * * *"),
		HealthAddr:        env("HEALTH_ADDR", ":8080"),
	}
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}
