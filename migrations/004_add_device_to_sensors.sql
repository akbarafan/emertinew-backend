-- Migration: Add device column to m_sensors table
-- This allows differentiation between top and bot sensors on the same topic

-- Add device column (nullable) - GORM AutoMigrate will handle this
-- But we add it here for completeness
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'm_sensors' AND column_name = 'device'
    ) THEN
        ALTER TABLE m_sensors ADD COLUMN device VARCHAR(50);
        RAISE NOTICE 'Added device column to m_sensors';
    ELSE
        RAISE NOTICE 'Column device already exists in m_sensors, skipping...';
    END IF;
END $$;

-- Drop old unique index (if exists)
DROP INDEX IF EXISTS idx_mcu_sensor_topic;
DROP INDEX IF EXISTS idx_sensor_id_topic;

-- Create new unique index that includes device
-- Using COALESCE to handle NULL device values properly
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_indexes 
        WHERE indexname = 'idx_mcu_sensor_topic_device'
    ) THEN
        CREATE UNIQUE INDEX idx_mcu_sensor_topic_device ON m_sensors(mcu_id, id, topic, COALESCE(device, ''));
        RAISE NOTICE 'Created index idx_mcu_sensor_topic_device';
    ELSE
        RAISE NOTICE 'Index idx_mcu_sensor_topic_device already exists, skipping...';
    END IF;
END $$;
