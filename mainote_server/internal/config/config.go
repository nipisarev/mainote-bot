package config

import "os"

type Config struct {
	Port        string
	SentryDSN   string
	Environment string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("GO_PORT", "8081"),
		SentryDSN:   getEnv("SENTRY_DSN", ""),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
