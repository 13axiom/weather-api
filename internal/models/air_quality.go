package models

import "time"

// AQILevel maps a numeric AQI (1–5, OWM scale) to a human-readable label and hex colour.
// Used by both the API response and the frontend widget.
var AQILevel = map[int]struct {
	Label string `json:"label"`
	Color string `json:"color"`
}{
	1: {Label: "Good",      Color: "#22c55e"},
	2: {Label: "Fair",      Color: "#84cc16"},
	3: {Label: "Moderate",  Color: "#eab308"},
	4: {Label: "Poor",      Color: "#f97316"},
	5: {Label: "Very Poor", Color: "#ef4444"},
}

// AirQualitySnapshot is one air-quality measurement stored in the database.
type AirQualitySnapshot struct {
	ID         int       `json:"id"`
	CityID     int       `json:"city_id"`
	CityName   string    `json:"city_name,omitempty"`

	// AQI index: 1 (Good) … 5 (Very Poor) — OpenWeatherMap scale
	AQI  int     `json:"aqi"`

	// Individual pollutant concentrations, all in µg/m³
	CO   float64 `json:"co"`    // Carbon monoxide
	NO2  float64 `json:"no2"`   // Nitrogen dioxide
	O3   float64 `json:"o3"`    // Ozone
	PM25 float64 `json:"pm2_5"` // Fine particles
	PM10 float64 `json:"pm10"`  // Coarse particles
	SO2  float64 `json:"so2"`   // Sulphur dioxide

	// Computed label and colour for the frontend (not stored in DB)
	AQILabel string `json:"aqi_label,omitempty"`
	AQIColor string `json:"aqi_color,omitempty"`

	RecordedAt time.Time `json:"recorded_at"`
	SyncedAt   time.Time `json:"synced_at"`
}

// AirQualityResponse is what the API sends back for a city.
type AirQualityResponse struct {
	City    City                  `json:"city"`
	Current *AirQualitySnapshot   `json:"current"`
	History []AirQualitySnapshot  `json:"history"`
}

// AirQualitySyncResult describes the result of syncing one city's AQ data.
type AirQualitySyncResult struct {
	City    string `json:"city"`
	New     int    `json:"new_records"`
	Skipped int    `json:"skipped"`
	Error   string `json:"error,omitempty"`
}
