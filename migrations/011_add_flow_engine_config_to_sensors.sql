-- Add flow_engine and flow_direction columns to m_sensors
-- flow_engine: arbitrary engine name (e.g. "Port Engine", "Starboard Engine")
-- flow_direction: "in", "out", or NULL (for single/undirectional)
-- Only meaningful for sensors with sensor_group = "flow"

ALTER TABLE m_sensors
ADD COLUMN flow_engine VARCHAR(100) DEFAULT NULL;

ALTER TABLE m_sensors
ADD COLUMN flow_direction VARCHAR(10) DEFAULT NULL;
