-- 016_simplify_geofence_zones.sql
-- Add zone types and operational rules to m_geofences
ALTER TABLE m_geofences 
    ADD COLUMN IF NOT EXISTS zone_type VARCHAR(50) NOT NULL DEFAULT 'work_zone',
    ADD COLUMN IF NOT EXISTS max_speed_knots DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS alert_on_exit BOOLEAN NOT NULL DEFAULT true,
    ADD COLUMN IF NOT EXISTS alert_on_entry BOOLEAN NOT NULL DEFAULT false;

-- Enhance geofence_logs for event-driven zone tracking
ALTER TABLE geofence_logs 
    ADD COLUMN IF NOT EXISTS event_type VARCHAR(50) NOT NULL DEFAULT 'outside',
    ADD COLUMN IF NOT EXISTS speed DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS message TEXT;

-- Index for fast log lookup
CREATE INDEX IF NOT EXISTS idx_geofence_logs_vessel_created ON geofence_logs(vessel_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_geofence_logs_geofence_created ON geofence_logs(geofence_id, created_at DESC);
