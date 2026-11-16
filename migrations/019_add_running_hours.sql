-- Migration 019: Add vessel engine running hours and trip reset logs

CREATE TABLE IF NOT EXISTS vessel_engine_running_hours (
    id SERIAL PRIMARY KEY,
    vessel_id INTEGER NOT NULL,
    sensor_id INTEGER,
    engine_slot VARCHAR(20) NOT NULL,
    total_running_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    trip_running_hours DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_active_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_vessel_engine_slot UNIQUE (vessel_id, engine_slot)
);

CREATE INDEX IF NOT EXISTS idx_vessel_engine_slot ON vessel_engine_running_hours(vessel_id, engine_slot);
CREATE INDEX IF NOT EXISTS idx_vessel_running_sensor ON vessel_engine_running_hours(sensor_id);

CREATE TABLE IF NOT EXISTS trip_running_hours_reset_logs (
    id SERIAL PRIMARY KEY,
    vessel_id INTEGER NOT NULL,
    sensor_id INTEGER,
    engine_slot VARCHAR(20) NOT NULL,
    hours_before_reset DOUBLE PRECISION NOT NULL,
    trip_name VARCHAR(255),
    reset_by_user_id INTEGER,
    reset_by_name VARCHAR(100) NOT NULL,
    reset_by_role VARCHAR(50) NOT NULL,
    reset_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_trip_resets_vessel ON trip_running_hours_reset_logs(vessel_id);
CREATE INDEX IF NOT EXISTS idx_trip_resets_sensor ON trip_running_hours_reset_logs(sensor_id);
CREATE INDEX IF NOT EXISTS idx_trip_resets_at ON trip_running_hours_reset_logs(reset_at);

-- High-performance composite index for 100+ vessels sub-millisecond running hours queries
CREATE INDEX IF NOT EXISTS idx_trackings_vessel_time ON trackings(vessel_id, time DESC);
