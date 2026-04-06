package service

import (
	"crypto/sha256"
	"fmt"
	"log"
	"time"

	"github.com/13axiom/weather-api/internal/client"
	"github.com/13axiom/weather-api/internal/database"
	"github.com/13axiom/weather-api/internal/models"
)

// SyncService periodically fetches weather data and stores it in the database.
// Only new data (determined by a content hash) is inserted.
type SyncService struct {
	db       *database.DB
	client   *client.OpenMeteoClient
	interval time.Duration
	cities   []string
}

// NewSyncService creates a SyncService.
// interval controls how often data is synced (set via SYNC_INTERVAL_MINUTES env var).
func NewSyncService(
	db *database.DB,
	c *client.OpenMeteoClient,
	interval time.Duration,
	cities []string,
) *SyncService {
	return &SyncService{db: db, client: c, interval: interval, cities: cities}
}

// Start launches the background sync goroutine.
// Syncs immediately on startup, then repeats every interval.
func (s *SyncService) Start() {
	go func() {
		log.Printf("[sync] Starting — interval: %v, cities: %v", s.interval, s.cities)
		s.SyncAll()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for range ticker.C {
			s.SyncAll()
		}
	}()
}

// SyncAll syncs all configured cities and returns results.
// Can also be called directly via the POST /api/v1/sync endpoint.
func (s *SyncService) SyncAll() []models.SyncResult {
	results := make([]models.SyncResult, 0, len(s.cities))
	for _, city := range s.cities {
		res := s.syncCity(city)
		results = append(results, res)
		if res.Error != "" {
			log.Printf("[sync] ERROR %s: %s", city, res.Error)
		} else {
			log.Printf("[sync] %s: new=%d skipped=%d", city, res.New, res.Skipped)
		}
	}
	return results
}

func (s *SyncService) syncCity(cityName string) models.SyncResult {
	res := models.SyncResult{City: cityName}

	// 1. Ensure city exists in DB (creates it if not found)
	city, err := s.ensureCity(cityName)
	if err != nil {
		res.Error = fmt.Sprintf("ensureCity: %v", err)
		return res
	}

	// 2. Fetch current weather from Open-Meteo
	data, err := s.client.GetWeather(city.Latitude, city.Longitude)
	if err != nil {
		res.Error = fmt.Sprintf("getWeather: %v", err)
		return res
	}

	// 3. Compute content hash for deduplication
	hash := ComputeHash(data.Temperature, data.Windspeed, data.Precipitation,
		data.WeatherCode, data.Time.Format("2006-01-02T15:04"))

	// 4. Insert only if this hash hasn't been stored for this city yet
	result, err := s.db.Exec(`
		INSERT INTO weather_snapshots
			(city_id, temperature, windspeed, precipitation, weather_code, data_hash, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (city_id, data_hash) DO NOTHING`,
		city.ID, data.Temperature, data.Windspeed, data.Precipitation,
		data.WeatherCode, hash, data.Time,
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

// ensureCity returns a city from DB or creates it by geocoding the name.
func (s *SyncService) ensureCity(name string) (*models.City, error) {
	city := &models.City{}
	err := s.db.QueryRow(
		`SELECT id, name, latitude, longitude FROM cities WHERE name = $1`, name,
	).Scan(&city.ID, &city.Name, &city.Latitude, &city.Longitude)
	if err == nil {
		return city, nil // found
	}

	// Not found — geocode and insert
	geo, err := s.client.GetCoordinates(name)
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

// ComputeHash computes a deterministic SHA-256 hash for a weather reading.
// Exported so it can be used in unit tests.
func ComputeHash(temp, wind, precip float64, code int, timeStr string) string {
	raw := fmt.Sprintf("%.2f|%.2f|%.2f|%d|%s", temp, wind, precip, code, timeStr)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
}
