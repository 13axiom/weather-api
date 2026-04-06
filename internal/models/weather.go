package models

import "time"

// City represents a tracked city with its geographic coordinates.
type City struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	CreatedAt time.Time `json:"created_at"`
}

// WeatherSnapshot is a single weather data point stored in the database.
type WeatherSnapshot struct {
	ID            int       `json:"id"`
	CityID        int       `json:"city_id"`
	CityName      string    `json:"city_name,omitempty"`
	Temperature   float64   `json:"temperature"`
	Windspeed     float64   `json:"windspeed"`
	Precipitation float64   `json:"precipitation"`
	WeatherCode   int       `json:"weather_code"`
	DataHash      string    `json:"-"` // internal deduplication — not exposed in API
	RecordedAt    time.Time `json:"recorded_at"`
	SyncedAt      time.Time `json:"synced_at"`
}

// WeatherResponse is the API response for weather requests.
type WeatherResponse struct {
	City    City              `json:"city"`
	Current *WeatherSnapshot  `json:"current"`
	History []WeatherSnapshot `json:"history"`
}

// SyncResult describes the result of syncing a single city.
type SyncResult struct {
	City    string `json:"city"`
	New     int    `json:"new_records"`
	Skipped int    `json:"skipped"`
	Error   string `json:"error,omitempty"`
}
