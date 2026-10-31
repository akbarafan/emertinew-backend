-- Migration: Update sensor unique index to include mcu_id
-- This allows the same sensor_id and topic to exist on different MCUs (different vessels)

-- Drop old unique index
DROP INDEX IF EXISTS idx_sensor_id_topic;

-- Create new unique index that includes mcu_id
CREATE UNIQUE INDEX IF NOT EXISTS idx_mcu_sensor_topic ON m_sensors(mcu_id, id, topic);
