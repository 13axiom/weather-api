// Package client provides an HTTP client for the Open-Meteo API.
// No API key required — Open-Meteo is completely free and open.
// Docs: https://open-meteo.com/en/docs
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	geocodingBaseURL = "https://geocoding-api.open-meteo.com/v1/search"
	weatherBaseURL   = "https://api.open-meteo.com/v1/forecast"
)

// OpenMeteoClient fetches weather data from api.open-meteo.com.
type OpenMeteoClient struct {
	http *http.Client
}

// New creates a new OpenMeteoClient with a 10-second timeout.
func New() *OpenMeteoClient {
	return &OpenMeteoClient{
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

// GeoResult holds coordinates for a city resolved by the geocoding API.
type GeoResult struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Country   string  `json:"country"`
}

// WeatherData is the parsed result from the weather API.
type WeatherData struct {
	Temperature   float64
	Windspeed     float64
	Precipitation float64
	WeatherCode   int
	Time          time.Time
}

// GetCoordinates resolves a city name to geographic coordinates.
func (c *OpenMeteoClient) GetCoordinates(cityName string) (*GeoResult, error) {
	u := fmt.Sprintf("%s?name=%s&count=1&language=en&format=json",
		geocodingBaseURL, url.QueryEscape(cityName))

	resp, err := c.http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("geocoding request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Results []GeoResult `json:"results"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode geocoding response: %w", err)
	}
	if len(result.Results) == 0 {
		return nil, fmt.Errorf("city not found: %s", cityName)
	}
	return &result.Results[0], nil
}

// GetWeather fetches current weather data for the given coordinates.
func (c *OpenMeteoClient) GetWeather(lat, lon float64) (*WeatherData, error) {
	u := fmt.Sprintf(
		"%s?latitude=%.4f&longitude=%.4f&current_weather=true&hourly=precipitation&forecast_days=1",
		weatherBaseURL, lat, lon,
	)

	resp, err := c.http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("weather request: %w", err)
	}
	defer resp.Body.Close()

	var raw struct {
		CurrentWeather struct {
			Temperature float64 `json:"temperature"`
			Windspeed   float64 `json:"windspeed"`
			WeatherCode int     `json:"weathercode"`
			Time        string  `json:"time"`
		} `json:"current_weather"`
		Hourly struct {
			Precipitation []float64 `json:"precipitation"`
		} `json:"hourly"`
	}

	if err = json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode weather response: %w", err)
	}

	t, _ := time.Parse("2006-01-02T15:04", raw.CurrentWeather.Time)

	precip := 0.0
	if len(raw.Hourly.Precipitation) > 0 {
		precip = raw.Hourly.Precipitation[0]
	}

	return &WeatherData{
		Temperature:   raw.CurrentWeather.Temperature,
		Windspeed:     raw.CurrentWeather.Windspeed,
		WeatherCode:   raw.CurrentWeather.WeatherCode,
		Precipitation: precip,
		Time:          t,
	}, nil
}
