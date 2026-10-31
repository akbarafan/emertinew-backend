-- Fix unique index untuk vessel_dimensions dan vessel_capacity
-- Unique constraint sekarang hanya apply pada record yang belum di-soft delete
-- Ini memungkinkan re-create dimensions/capacity setelah delete tanpa konflik

-- Drop old unique indexes (if exists)
DROP INDEX IF EXISTS idx_vessel_dimensions_vessel_id;
DROP INDEX IF EXISTS idx_vessel_capacity_vessel_id;

-- Create partial unique indexes for vessel_dimensions (only if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vessel_dimensions') THEN
        DROP INDEX IF EXISTS idx_vessel_dimensions_vessel_id;
        CREATE UNIQUE INDEX idx_vessel_dimensions_vessel_id ON vessel_dimensions(vessel_id) WHERE deleted_at IS NULL;
        RAISE NOTICE 'Created index idx_vessel_dimensions_vessel_id';
    ELSE
        RAISE NOTICE 'Table vessel_dimensions does not exist, skipping...';
    END IF;
END $$;

-- Create partial unique indexes for vessel_capacity (only if table exists)
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vessel_capacity') THEN
        DROP INDEX IF EXISTS idx_vessel_capacity_vessel_id;
        CREATE UNIQUE INDEX idx_vessel_capacity_vessel_id ON vessel_capacity(vessel_id) WHERE deleted_at IS NULL;
        RAISE NOTICE 'Created index idx_vessel_capacity_vessel_id';
    ELSE
        RAISE NOTICE 'Table vessel_capacity does not exist, skipping...';
    END IF;
END $$;
