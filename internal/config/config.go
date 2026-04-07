package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration loaded from environment variables.
//
// Environment variables (set in .env or system environment):
//
//   DATABASE_URL          PostgreSQL connection string
//   PORT                  HTTP port (default 8080)
//   SYNC_INTERVAL_MINUTES How often to sync weather data (default 60)
//   ALLOWED_ORIGINS       Comma-separated CORS origins (default http://localhost:3000)
//   DEFAULT_CITIES        Comma-separated city names to track
//
//   OWM_API_KEY           OpenWeatherMap API key — NEVER expose to frontend
//                         Get yours free at: https://home.openweathermap.org/api_keys
//
//   INTERNAL_API_KEY      Secret shared with our own frontend (weather-ui).
//                         Protects air-quality endpoints from public access.
//                         Generate with: openssl rand -hex 32
type Config struct {
	DatabaseURL      string
	Port             string
	SyncIntervalMins int
	AllowedOrigins   []string
	DefaultCities    []string

	// OWMAPIKey is the OpenWeatherMap API key.
	// Used exclusively inside the Go backend to call OWM — never sent to browser.
	OWMAPIKey string

	// InternalAPIKey protects our own API routes (e.g. air quality).
	// The frontend sends this key; random internet traffic is rejected.
	InternalAPIKey string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	syncMins, _ := strconv.Atoi(getEnv("SYNC_INTERVAL_MINUTES", "60"))

	cities := strings.Split(
		getEnv("DEFAULT_CITIES", "Moscow,London,New York,Tokyo,Sydney"),
		",",
	)
	for i, c := range cities {
		cities[i] = strings.TrimSpace(c)
	}

	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://weather_user:weather_pass@localhost:5432/weather_db"),
		Port:             getEnv("PORT", "8080"),
		SyncIntervalMins: syncMins,
		AllowedOrigins:   strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		DefaultCities:    cities,
		OWMAPIKey:        getEnv("OWM_API_KEY", ""),        // empty = OWM disabled
		InternalAPIKey:   getEnv("INTERNAL_API_KEY", ""),   // empty = auth skipped in dev
	}
}

// SyncInterval returns the sync interval as a time.Duration.
func (c *Config) SyncInterval() time.Duration {
	return time.Duration(c.SyncIntervalMins) * time.Minute
}

// OWMEnabled reports whether the OpenWeatherMap key is configured.
func (c *Config) OWMEnabled() bool {
	return c.OWMAPIKey != ""
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
