package config

import (
	"fmt"
	"os"
)

// Config holds runtime configuration sourced from environment variables.
type Config struct {
	HTTPPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
}

// Load reads configuration from the environment, applying defaults.
func Load() Config {
	return Config{
		HTTPPort:   env("HTTP_PORT", "8080"),
		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "5432"),
		DBUser:     env("DB_USER", "postgres"),
		DBPassword: env("DB_PASS", "postgres"),
		DBName:     env("DB_NAME", "entaintest"),
		DBSSLMode:  env("DB_SSLMODE", "disable"),
	}
}

// DSN returns a libpq-style connection string for pgx.
func (c Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}
