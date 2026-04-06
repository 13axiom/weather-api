package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL      string
	Port             string
	SyncIntervalMins int
	AllowedOrigins   []string
	DefaultCities    []string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	syncMins, _ := strconv.Atoi(getEnv("SYNC_INTERVAL_MINUTES", "60"))
	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://weather_user:weather_pass@localhost:5432/weather_db"),
		Port:             getEnv("PORT", "8080"),
		SyncIntervalMins: syncMins,
		AllowedOrigins:   strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		DefaultCities:    []string{"Moscow", "London", "New York", "Tokyo", "Sydney"},
	}
}

// SyncInterval returns the sync interval as a time.Duration.
func (c *Config) SyncInterval() time.Duration {
	return time.Duration(c.SyncIntervalMins) * time.Minute
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
