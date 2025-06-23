package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	SentryDSN       string
	Environment     string
	DatabaseURL     string
	NotionAPIKey    string
	NotionDatabaseID string
}

func Load() *Config {
	// Build database URL from individual components
	dbUser := getEnv("POSTGRES_USER", "mainote")
	dbPassword := getEnv("POSTGRES_PASSWORD", "mainote_dev_password")
	dbHost := getEnv("DATABASE_HOST", "postgres")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbName := getEnv("POSTGRES_DB", "mainote")
	
	databaseURL := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	return &Config{
		Port:            getEnv("APP_PORT", "8081"),
		SentryDSN:       getEnv("SENTRY_DSN", ""),
		Environment:     getEnv("ENVIRONMENT", "development"),
		DatabaseURL:     databaseURL,
		NotionAPIKey:    getEnv("NOTION_API_KEY", ""),
		NotionDatabaseID: getEnv("NOTION_DATABASE_ID", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
