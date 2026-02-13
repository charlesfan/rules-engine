package config

import (
	"os"
)

// Config holds all configuration for the application
type Config struct {
	DatabaseURL string
	ServerPort  string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/rules_engine?sslmode=disable"),
		ServerPort:  getEnv("SERVER_PORT", "8080"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
