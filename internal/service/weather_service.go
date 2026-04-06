package service

import (
	"database/sql"
	"fmt"

	"github.com/13axiom/weather-api/internal/database"
	"github.com/13axiom/weather-api/internal/models"
)

// WeatherService reads weather data from the database.
type WeatherService struct {
	db *database.DB
}

// NewWeatherService creates a WeatherService.
func NewWeatherService(db *database.DB) *WeatherService {
	return &WeatherService{db: db}
}

// GetCities returns all tracked cities sorted by name.
func (s *WeatherService) GetCities() ([]models.City, error) {
	rows, err := s.db.Query(
		`SELECT id, name, latitude, longitude, created_at FROM cities ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("query cities: %w", err)
	}
	defer rows.Close()

	var cities []models.City
	for rows.Next() {
		var c models.City
		if err = rows.Scan(&c.ID, &c.Name, &c.Latitude, &c.Longitude, &c.CreatedAt); err != nil {
			return nil, err
		}
		cities = append(cities, c)
	}
	return cities, rows.Err()
}

// GetWeather returns the current snapshot and recent history for a city.
func (s *WeatherService) GetWeather(cityName string, limit int) (*models.WeatherResponse, error) {
	city := &models.City{}
	err := s.db.QueryRow(
		`SELECT id, name, latitude, longitude FROM cities WHERE LOWER(name) = LOWER($1)`,
		cityName,
	).Scan(&city.ID, &city.Name, &city.Latitude, &city.Longitude)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("city not found: %s", cityName)
		}
		return nil, fmt.Errorf("query city: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT id, city_id, temperature, windspeed, precipitation,
		       weather_code, recorded_at, synced_at
		FROM weather_snapshots
		WHERE city_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2`,
		city.ID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []models.WeatherSnapshot
	for rows.Next() {
		var snap models.WeatherSnapshot
		if err = rows.Scan(
			&snap.ID, &snap.CityID, &snap.Temperature, &snap.Windspeed,
			&snap.Precipitation, &snap.WeatherCode, &snap.RecordedAt, &snap.SyncedAt,
		); err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snap)
	}

	var current *models.WeatherSnapshot
	if len(snapshots) > 0 {
		current = &snapshots[0]
	}

	return &models.WeatherResponse{
		City:    *city,
		Current: current,
		History: snapshots,
	}, rows.Err()
}
