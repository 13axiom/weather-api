# weather-api

Go backend for the Weather Dashboard project. Syncs weather data from [Open-Meteo](https://open-meteo.com) (free, no API key required) into PostgreSQL and exposes a REST API.

## Stack

- **Language:** Go 1.21+
- **Router:** [chi](https://github.com/go-chi/chi)
- **Database:** PostgreSQL 16 (via Docker)
- **API docs:** Swagger UI (via swaggo/swag)
- **Data source:** [Open-Meteo](https://open-meteo.com) — completely free, no registration

## Quick Start

```bash
# 1. Start PostgreSQL
docker compose up -d

# 2. Copy env file
cp .env.example .env

# 3. Generate Swagger docs
swag init -g cmd/server/main.go --output docs

# 4. Run the server
go run ./cmd/server/...
```

Server starts on `http://localhost:8080`

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| GET | `/swagger/index.html` | Swagger UI |
| GET | `/api/v1/cities` | List tracked cities |
| GET | `/api/v1/weather/{city}` | Get weather + history |
| POST | `/api/v1/sync` | Manually trigger sync |

## Configuration

All settings via environment variables (see `.env.example`):

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://...` | PostgreSQL connection string |
| `PORT` | `8080` | HTTP server port |
| `SYNC_INTERVAL_MINUTES` | `60` | How often to sync from Open-Meteo |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | CORS allowed origins |

## Running Tests

```bash
go test ./... -v
```

## Project Structure

```
cmd/server/        — main.go entry point
internal/
  config/          — environment configuration
  database/        — DB connection + migrations
  client/          — Open-Meteo HTTP client
  models/          — shared data types
  service/         — business logic (sync, weather queries)
  handler/         — HTTP handlers
docs/              — generated Swagger (run swag init)
```
