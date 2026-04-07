// Package client — OpenWeatherMap Air Pollution API client.
//
// Provider: https://openweathermap.org/api/air-pollution
// Free tier: 60 calls/minute, no daily cap.
// Key required: yes — register at https://home.openweathermap.org/api_keys
// Set the key via the OWM_API_KEY environment variable (never in source code).
//
// Endpoint used:
//   GET https://api.openweathermap.org/data/2.5/air_pollution
//       ?lat={lat}&lon={lon}&appid={OWM_API_KEY}
//
// Response AQI scale (OWM):
//   1 = Good  2 = Fair  3 = Moderate  4 = Poor  5 = Very Poor
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const owmBaseURL = "https://api.openweathermap.org/data/2.5/air_pollution"

// OWMClient calls the OpenWeatherMap Air Pollution API.
type OWMClient struct {
	apiKey string
	http   *http.Client
}

// NewOWMClient creates a new OWMClient.
// apiKey must be a valid OpenWeatherMap API key (free tier is fine).
func NewOWMClient(apiKey string) *OWMClient {
	return &OWMClient{
		apiKey: apiKey,
		http:   &http.Client{Timeout: 10 * time.Second},
	}
}

// AirQualityData is the parsed result from one OWM Air Pollution call.
type AirQualityData struct {
	AQI  int
	CO   float64 // µg/m³
	NO2  float64 // µg/m³
	O3   float64 // µg/m³
	PM25 float64 // µg/m³
	PM10 float64 // µg/m³
	SO2  float64 // µg/m³
	Time time.Time
}

// GetAirQuality fetches current air quality for the given coordinates.
func (c *OWMClient) GetAirQuality(lat, lon float64) (*AirQualityData, error) {
	// API key is appended here — stays server-side, never reaches the browser.
	url := fmt.Sprintf("%s?lat=%.4f&lon=%.4f&appid=%s", owmBaseURL, lat, lon, c.apiKey)

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("owm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("owm returned HTTP %d — check OWM_API_KEY", resp.StatusCode)
	}

	// OWM response shape (simplified):
	// {"list":[{"main":{"aqi":2},"components":{...},"dt":1234567890}]}
	var raw struct {
		List []struct {
			Main struct {
				AQI int `json:"aqi"`
			} `json:"main"`
			Components struct {
				CO   float64 `json:"co"`
				NO2  float64 `json:"no2"`
				O3   float64 `json:"o3"`
				PM25 float64 `json:"pm2_5"`
				PM10 float64 `json:"pm10"`
				SO2  float64 `json:"so2"`
			} `json:"components"`
			DT int64 `json:"dt"` // Unix timestamp
		} `json:"list"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode owm response: %w", err)
	}
	if len(raw.List) == 0 {
		return nil, fmt.Errorf("owm returned empty list for lat=%.4f lon=%.4f", lat, lon)
	}

	item := raw.List[0]
	return &AirQualityData{
		AQI:  item.Main.AQI,
		CO:   item.Components.CO,
		NO2:  item.Components.NO2,
		O3:   item.Components.O3,
		PM25: item.Components.PM25,
		PM10: item.Components.PM10,
		SO2:  item.Components.SO2,
		Time: time.Unix(item.DT, 0).UTC(),
	}, nil
}
