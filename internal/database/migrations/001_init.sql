-- Weather Dashboard initial schema
-- Run automatically on server startup via db.Migrate()

CREATE TABLE IF NOT EXISTS cities (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(100) UNIQUE NOT NULL,
    latitude   DOUBLE PRECISION    NOT NULL,
    longitude  DOUBLE PRECISION    NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS weather_snapshots (
    id            SERIAL PRIMARY KEY,
    city_id       INTEGER REFERENCES cities(id) ON DELETE CASCADE,
    temperature   DOUBLE PRECISION,
    windspeed     DOUBLE PRECISION,
    precipitation DOUBLE PRECISION,
    weather_code  INTEGER,
    -- data_hash prevents storing duplicate readings
    data_hash     VARCHAR(64) NOT NULL,
    recorded_at   TIMESTAMPTZ NOT NULL,
    synced_at     TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(city_id, data_hash)
);

-- Index for fast "latest N records for a city" queries
CREATE INDEX IF NOT EXISTS idx_snapshots_city_recorded
    ON weather_snapshots(city_id, recorded_at DESC);
