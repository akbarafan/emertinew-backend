-- 017_add_geofence_routes.sql
-- Create tables for maritime route corridor geofencing

CREATE TABLE IF NOT EXISTS m_geofence_routes (
    id SERIAL PRIMARY KEY,
    company_id INT REFERENCES m_companies(id) ON UPDATE CASCADE ON DELETE SET NULL,
    voyage_id INT REFERENCES t_voyages(id) ON UPDATE CASCADE ON DELETE SET NULL,
    vessel_id INT REFERENCES m_vessels(id) ON UPDATE CASCADE ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    total_distance_nm DOUBLE PRECISION DEFAULT 0,
    estimated_duration_hours DOUBLE PRECISION,
    default_corridor_width_m DOUBLE PRECISION DEFAULT 926,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    is_active BOOLEAN DEFAULT true,
    alert_on_deviation BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by INT,
    updated_by INT,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_m_geofence_routes_company_id ON m_geofence_routes(company_id);
CREATE INDEX IF NOT EXISTS idx_m_geofence_routes_voyage_id ON m_geofence_routes(voyage_id);
CREATE INDEX IF NOT EXISTS idx_m_geofence_routes_vessel_id ON m_geofence_routes(vessel_id);
CREATE INDEX IF NOT EXISTS idx_m_geofence_routes_deleted_at ON m_geofence_routes(deleted_at);

CREATE TABLE IF NOT EXISTS m_geofence_route_waypoints (
    id SERIAL PRIMARY KEY,
    route_id INT NOT NULL REFERENCES m_geofence_routes(id) ON UPDATE CASCADE ON DELETE CASCADE,
    sequence INT NOT NULL,
    name VARCHAR(255),
    latitude DECIMAL(10,6) NOT NULL,
    longitude DECIMAL(10,6) NOT NULL,
    waypoint_type VARCHAR(50) NOT NULL DEFAULT 'waypoint',
    corridor_width_m DOUBLE PRECISION DEFAULT 926,
    radius_m DOUBLE PRECISION DEFAULT 185,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by INT,
    updated_by INT,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_m_geofence_route_waypoints_route_id ON m_geofence_route_waypoints(route_id);
CREATE INDEX IF NOT EXISTS idx_m_geofence_route_waypoints_sequence ON m_geofence_route_waypoints(route_id, sequence);
