# weather-api

Go backend for the Weather Dashboard project.

Provides two data sources:

- **Weather** — [Open-Meteo](https://open-meteo.com) (completely free, no API key required)
- **Air Quality** — [OpenWeatherMap Air Pollution API](https://openweathermap.org/api/air-pollution) (free tier, requires a free API key)

## Stack

- **Language:** Go 1.21+
- **Router:** [chi](https://github.com/go-chi/chi)
- **Database:** PostgreSQL 16 (via Docker)
- **API docs:** Swagger UI (swaggo/swag)

## Quick Start

```bash
# 1. Start PostgreSQL
docker compose up -d

# 2. Copy and edit env file
cp .env.example .env
# → Fill in OWM_API_KEY and INTERNAL_API_KEY (see Configuration below)

# 3. Generate Swagger docs
swag init -g cmd/server/main.go --output docs

# 4. Run the server
go run ./cmd/server/...
```

Server starts on `http://localhost:8080`

---

## API Endpoints

### Weather (public — no key required)

| Method | Path                              | Description                            |
| ------ | --------------------------------- | -------------------------------------- |
| GET    | `/health`                         | Health check (used by uptime monitors) |
| GET    | `/swagger/index.html`             | Interactive API documentation          |
| GET    | `/api/v1/cities`                  | List all tracked cities                |
| GET    | `/api/v1/weather/{city}?limit=24` | Current weather + history              |
| POST   | `/api/v1/sync`                    | Manually trigger weather sync          |

### Air Quality (protected — requires `X-Internal-Key` header)

| Method | Path                          | Description                        |
| ------ | ----------------------------- | ---------------------------------- |
| GET    | `/api/v1/air`                 | Latest AQI snapshot for all cities |
| GET    | `/api/v1/air/{city}?limit=24` | Current AQI + history for one city |
| POST   | `/api/v1/air/sync`            | Manually trigger air quality sync  |

---

## Configuration

All settings via environment variables. Copy `.env.example` → `.env`.

| Variable                | Default                 | Description                                         |
| ----------------------- | ----------------------- | --------------------------------------------------- |
| `DATABASE_URL`          | `postgres://...`        | PostgreSQL connection string                        |
| `PORT`                  | `8080`                  | HTTP server port                                    |
| `SYNC_INTERVAL_MINUTES` | `60`                    | How often to sync data (weather + AQ)               |
| `DEFAULT_CITIES`        | `Moscow,London,...`     | Cities to track (comma-separated)                   |
| `ALLOWED_ORIGINS`       | `http://localhost:3000` | CORS allowed origins                                |
| `OWM_API_KEY`           | _(empty)_               | OpenWeatherMap key — get free at openweathermap.org |
| `INTERNAL_API_KEY`      | _(empty)_               | Shared secret between backend and frontend          |

### Setting up the API keys

#### 1. OpenWeatherMap key (for Air Quality)

The OWM key is used **only inside the Go server** to call OpenWeatherMap.
It is never sent to the browser. This is the key security property.

```
1. Go to https://home.openweathermap.org/api_keys
2. Create a free account and generate a key
3. Add to weather-api/.env:
   OWM_API_KEY=your_key_here
```

Free tier limits: 60 calls/minute — more than enough for hourly syncs.

#### 2. Internal API key (protects our own /air/\* endpoints)

This key is **not** an external provider key — it's a secret we generate ourselves.
The frontend sends it on every air quality request to prove it's "our" frontend,
not a random internet user eating our OWM quota.

```bash
# Generate the key (run once):
openssl rand -hex 32
# → e.g. a3f8c2d1...

# Add to weather-api/.env:
INTERNAL_API_KEY=a3f8c2d1...

# Add the SAME value to weather-ui/.env.local:
NEXT_PUBLIC_INTERNAL_KEY=a3f8c2d1...
```

If `INTERNAL_API_KEY` is empty, the auth check is skipped (convenient for local dev)

---

## How data flows (Air Quality)

```
Browser (weather-ui)
    │  sends X-Internal-Key header
    ▼
Go server (/api/v1/air/*)
    │  middleware checks X-Internal-Key
    │  if valid → calls OWM with OWM_API_KEY (server-side only)
    ▼
OpenWeatherMap Air Pollution API
    │  returns AQI + pollutant concentrations
    ▼
PostgreSQL (air_quality_snapshots table)
    │  stored, deduplicated by hash
    ▼
Go server returns clean JSON to frontend
```

---

## Running Tests

```bash
go test ./... -v
```

---

## Project Structure

```
cmd/server/          — main.go entry point (wires everything together)
internal/
  config/            — env var loading with documented defaults
  database/          — DB connection + running migrations
    migrations/
      001_init.sql   — cities + weather_snapshots tables
      002_air_quality.sql — air_quality_snapshots table
  client/
    openmeteo.go     — Open-Meteo weather + geocoding (no key)
    openweathermap.go — OWM Air Pollution API (requires OWM_API_KEY)
  models/
    weather.go       — City, WeatherSnapshot, WeatherResponse
    air_quality.go   — AirQualitySnapshot, AirQualityResponse, AQI levels
  service/
    weather_service.go      — reads weather from DB
    sync_service.go         — syncs weather from Open-Meteo
    air_quality_service.go  — syncs AQ from OWM + reads from DB
  handler/
    weather_handler.go      — GET /cities, GET /weather/{city}
    sync_handler.go         — POST /sync
    air_quality_handler.go  — GET /air, GET /air/{city}, POST /air/sync
    helpers.go              — respond / respondError helpers
  middleware/
    auth.go          — RequireInternalKey middleware
docs/                — generated Swagger (run: swag init -g cmd/server/main.go)
```
