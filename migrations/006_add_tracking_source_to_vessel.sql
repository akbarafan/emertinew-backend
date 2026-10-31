-- Migration: Add tracking_source column to m_vessel
-- Date: 2026-03-04
-- Allowed values enforced at application level: 'alugara', 'hexa'

ALTER TABLE m_vessels
    ADD COLUMN IF NOT EXISTS tracking_source VARCHAR(10) DEFAULT NULL;

-- Add CHECK constraint for allowed values
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_vessel_tracking_source'
    ) THEN
        ALTER TABLE m_vessels
            ADD CONSTRAINT chk_vessel_tracking_source
            CHECK (tracking_source IN ('alugara', 'hexa'));
    END IF;
END
$$;
