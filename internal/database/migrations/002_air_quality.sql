-- Air Quality feature — migration 002
-- Provider: OpenWeatherMap Air Pollution API (free, requires OWM_API_KEY)
-- Safe to run on every startup — uses CREATE TABLE IF NOT EXISTS

CREATE TABLE IF NOT EXISTS air_quality_snapshots (
    id                  SERIAL PRIMARY KEY,
    city_id             INTEGER REFERENCES cities(id) ON DELETE CASCADE,

    -- AQI scores (1=Good, 2=Fair, 3=Moderate, 4=Poor, 5=Very Poor — OWM scale)
    aqi                 INTEGER,

    -- Pollutant concentrations in µg/m³ (except co which is in µg/m³ too)
    co                  DOUBLE PRECISION,   -- Carbon monoxide
    no2                 DOUBLE PRECISION,   -- Nitrogen dioxide
    o3                  DOUBLE PRECISION,   -- Ozone
    pm2_5               DOUBLE PRECISION,   -- Fine particles (≤2.5 µm)
    pm10                DOUBLE PRECISION,   -- Coarse particles (≤10 µm)
    so2                 DOUBLE PRECISION,   -- Sulphur dioxide

    -- Deduplication: same hash = same reading, don't store twice
    data_hash           VARCHAR(64) NOT NULL,
    recorded_at         TIMESTAMPTZ NOT NULL,
    synced_at           TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(city_id, data_hash)
);

-- Fast lookup: latest N records for a city
CREATE INDEX IF NOT EXISTS idx_aq_city_recorded
    ON air_quality_snapshots(city_id, recorded_at DESC);
