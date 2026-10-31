-- Add is_visible column to m_sensors for daily report chart visibility
-- Default true so existing sensors remain visible

ALTER TABLE m_sensors
ADD COLUMN is_visible boolean NOT NULL DEFAULT true;
