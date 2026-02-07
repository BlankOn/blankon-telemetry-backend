-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- Create events table
CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL,
    event_name VARCHAR(255) NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (id, timestamp)
);

-- Convert to hypertable (auto-partitions by time)
SELECT create_hypertable('events', 'timestamp', if_not_exists => TRUE);

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_events_name ON events(event_name, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at DESC);

-- Enable compression for data older than 7 days
ALTER TABLE events SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'event_name'
);

SELECT add_compression_policy('events', INTERVAL '7 days', if_not_exists => TRUE);

-- Continuous aggregate: hourly event counts (for dashboards)
CREATE MATERIALIZED VIEW IF NOT EXISTS events_hourly
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 hour', timestamp) AS bucket,
    event_name,
    COUNT(*) as event_count,
    COUNT(DISTINCT payload->>'user_id') as unique_users
FROM events
GROUP BY bucket, event_name
WITH NO DATA;

-- Refresh policy: update hourly aggregate every 30 minutes
SELECT add_continuous_aggregate_policy('events_hourly',
    start_offset => INTERVAL '3 hours',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '30 minutes',
    if_not_exists => TRUE
);

-- Continuous aggregate: daily event counts
CREATE MATERIALIZED VIEW IF NOT EXISTS events_daily
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 day', timestamp) AS bucket,
    event_name,
    COUNT(*) as event_count,
    COUNT(DISTINCT payload->>'user_id') as unique_users
FROM events
GROUP BY bucket, event_name
WITH NO DATA;

SELECT add_continuous_aggregate_policy('events_daily',
    start_offset => INTERVAL '3 days',
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

-- Optional: retention policy (delete data older than 1 year)
-- SELECT add_retention_policy('events', INTERVAL '1 year', if_not_exists => TRUE);
