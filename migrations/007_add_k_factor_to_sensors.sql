-- Migration: Add k_factor calibration column to m_sensors table
-- k_factor is the calibration factor in pulse/L for flow sensors
-- Volume (L) = pulse / k_factor
-- Default 0 means calibration not configured (skip calculation)

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'm_sensors' AND column_name = 'k_factor'
    ) THEN
        ALTER TABLE m_sensors ADD COLUMN k_factor DECIMAL(15,6) DEFAULT 0;
        RAISE NOTICE 'Added k_factor column to m_sensors';
    ELSE
        RAISE NOTICE 'Column k_factor already exists in m_sensors, skipping...';
    END IF;
END $$;
