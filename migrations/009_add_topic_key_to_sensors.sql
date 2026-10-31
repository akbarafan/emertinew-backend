-- Migration: Add normalized topic_key to m_sensors for cross-format duplicate detection
-- Prevents: hcml/vs001/bd/gps/1 vs hcml/flt001/sqd001/vs001/brd/gps/1 being treated as different sensors
-- TopicKey format (5-part): company:vessel_code:location:sensor_group:sensor_number
-- TopicKey format (7-part): company:vessel_code:location:sensor_group:sensor_number (flight/squad stripped)
-- Date: 2026-06-16

-- Add topic_key column (nullable initially for backfill)
ALTER TABLE m_sensors
    ADD COLUMN IF NOT EXISTS topic_key VARCHAR(255);

-- Backfill existing sensors: topic_key = topic with / replaced by :
UPDATE m_sensors SET topic_key = REPLACE(topic, '/', ':') WHERE topic_key IS NULL;

-- Make topic_key NOT NULL after backfill
ALTER TABLE m_sensors
    ALTER COLUMN topic_key SET NOT NULL;

-- Add unique index for topic_key
CREATE UNIQUE INDEX IF NOT EXISTS idx_m_sensors_topic_key ON m_sensors(topic_key);
