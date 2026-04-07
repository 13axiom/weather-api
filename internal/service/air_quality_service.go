// Package service — AirQualityService syncs and reads air quality data.
//
// Data flow:
//   OpenWeatherMap API → OWMClient → AirQualityService → PostgreSQL → handler → frontend
//
// The OWM API key never leaves the server. The frontend only sees
// our own API responses, protected by the internal API key middleware.
package service

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/13axiom/weather-api/internal/client"
	"github.com/13axiom/weather-api/internal/database"
	"github.com/13axiom/weather-api/internal/models"
)

// AirQualityService syncs air quality data and reads it from the database.
type AirQualityService struct {
	db       *database.DB
	owm      *client.OWMClient
	meteo    *client.OpenMeteoClient // reused for geocoding
	interval time.Duration
	cities   []string
}

// NewAirQualityService creates an AirQualityService.
func NewAirQualityService(
	db *database.DB,
	owm *client.OWMClient,
	meteo *client.OpenMeteoClient,
	interval time.Duration,
	cities []string,
) *AirQualityService {
	return &AirQualityService{db: db, owm: owm, meteo: meteo, interval: interval, cities: cities}
}

// Start launches the background sync goroutine for air quality data.
// Syncs immediately on startup, then repeats every interval.
func (s *AirQualityService) Start() {
	go func() {
		log.Printf("[aq-sync] Starting — interval: %v, cities: %v", s.interval, s.cities)
		s.SyncAll()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for range ticker.C {
			s.SyncAll()
		}
	}()
}

// SyncAll fetches and stores air quality data for all configured cities.
// Called automatically on the interval, or on-demand via POST /api/v1/air/sync.
func (s *AirQualityService) SyncAll() []models.AirQualitySyncResult {
	results := make([]models.AirQualitySyncResult, 0, len(s.cities))
	for _, city := range s.cities {
		res := s.syncCity(city)
		results = append(results, res)
		if res.Error != "" {
			log.Printf("[aq-sync] ERROR %s: %s", city, res.Error)
		} else {
			log.Printf("[aq-sync] %s: new=%d skipped=%d", city, res.New, res.Skipped)
		}
	}
	return results
}

func (s *AirQualityService) syncCity(cityName string) models.AirQualitySyncResult {
	res := models.AirQualitySyncResult{City: cityName}

	// 1. Ensure city exists (reuses the cities table, same as weather)
	city, err := s.ensureCity(cityName)
	if err != nil {
		res.Error = fmt.Sprintf("ensureCity: %v", err)
		return res
	}

	// 2. Fetch air quality from OpenWeatherMap
	data, err := s.owm.GetAirQuality(city.Latitude, city.Longitude)
	if err != nil {
		res.Error = fmt.Sprintf("getAirQuality: %v", err)
		return res
	}

	// 3. Deduplication hash (same reading at same minute = skip)
	hash := computeAQHash(data.AQI, data.PM25, data.PM10, data.Time.Format("2006-01-02T15:04"))

	// 4. Insert only if not seen before for this city
	result, err := s.db.Exec(`
		INSERT INTO air_quality_snapshots
			(city_id, aqi, co, no2, o3, pm2_5, pm10, so2, data_hash, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (city_id, data_hash) DO NOTHING`,
		city.ID, data.AQI, data.CO, data.NO2, data.O3,
		data.PM25, data.PM10, data.SO2, hash, data.Time,
	)
	if err != nil {
		res.Error = fmt.Sprintf("insert: %v", err)
		return res
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		res.New = 1
	} else {
		res.Skipped = 1
	}
	return res
}

// GetAirQuality returns the current snapshot and history for a city.
func (s *AirQualityService) GetAirQuality(cityName string, limit int) (*models.AirQualityResponse, error) {
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
		SELECT id, city_id, aqi, co, no2, o3, pm2_5, pm10, so2, recorded_at, synced_at
		FROM air_quality_snapshots
		WHERE city_id = $1
		ORDER BY recorded_at DESC
		LIMIT $2`,
		city.ID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query aq snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []models.AirQualitySnapshot
	for rows.Next() {
		var snap models.AirQualitySnapshot
		if err = rows.Scan(
			&snap.ID, &snap.CityID, &snap.AQI,
			&snap.CO, &snap.NO2, &snap.O3, &snap.PM25, &snap.PM10, &snap.SO2,
			&snap.RecordedAt, &snap.SyncedAt,
		); err != nil {
			return nil, err
		}
		// Enrich with label and colour (not stored in DB, computed on read)
		if level, ok := models.AQILevel[snap.AQI]; ok {
			snap.AQILabel = level.Label
			snap.AQIColor = level.Color
		}
		snapshots = append(snapshots, snap)
	}

	var current *models.AirQualitySnapshot
	if len(snapshots) > 0 {
		current = &snapshots[0]
	}

	return &models.AirQualityResponse{
		City:    *city,
		Current: current,
		History: snapshots,
	}, rows.Err()
}

// GetAllCitiesAirQuality returns the latest AQ snapshot for every tracked city.
func (s *AirQualityService) GetAllCitiesAirQuality() ([]models.AirQualitySnapshot, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT ON (aq.city_id)
			aq.id, aq.city_id, c.name, aq.aqi,
			aq.co, aq.no2, aq.o3, aq.pm2_5, aq.pm10, aq.so2,
			aq.recorded_at, aq.synced_at
		FROM air_quality_snapshots aq
		JOIN cities c ON c.id = aq.city_id
		ORDER BY aq.city_id, aq.recorded_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query all aq: %w", err)
	}
	defer rows.Close()

	var result []models.AirQualitySnapshot
	for rows.Next() {
		var snap models.AirQualitySnapshot
		if err = rows.Scan(
			&snap.ID, &snap.CityID, &snap.CityName, &snap.AQI,
			&snap.CO, &snap.NO2, &snap.O3, &snap.PM25, &snap.PM10, &snap.SO2,
			&snap.RecordedAt, &snap.SyncedAt,
		); err != nil {
			return nil, err
		}
		if level, ok := models.AQILevel[snap.AQI]; ok {
			snap.AQILabel = level.Label
			snap.AQIColor = level.Color
		}
		result = append(result, snap)
	}
	return result, rows.Err()
}

// ensureCity reuses the existing cities table (same geocoding as weather sync).
func (s *AirQualityService) ensureCity(name string) (*models.City, error) {
	city := &models.City{}
	err := s.db.QueryRow(
		`SELECT id, name, latitude, longitude FROM cities WHERE name = $1`, name,
	).Scan(&city.ID, &city.Name, &city.Latitude, &city.Longitude)
	if err == nil {
		return city, nil
	}

	geo, err := s.meteo.GetCoordinates(name)
	if err != nil {
		return nil, fmt.Errorf("geocoding %q: %w", name, err)
	}

	err = s.db.QueryRow(`
		INSERT INTO cities (name, latitude, longitude)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id, name, latitude, longitude`,
		geo.Name, geo.Latitude, geo.Longitude,
	).Scan(&city.ID, &city.Name, &city.Latitude, &city.Longitude)
	return city, err
}

func computeAQHash(aqi int, pm25, pm10 float64, timeStr string) string {
	raw := fmt.Sprintf("%d|%.2f|%.2f|%s", aqi, pm25, pm10, timeStr)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
}
