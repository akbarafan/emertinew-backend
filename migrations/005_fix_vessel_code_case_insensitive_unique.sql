-- Fix vessel code unique index to be case-insensitive
-- Ensures VS005 and vs005 are treated as the same code
-- Also normalizes existing codes to uppercase

-- Step 1: Normalize existing codes to uppercase
UPDATE m_vessels SET code = UPPER(code) WHERE code != UPPER(code);

-- Step 2: Drop the old case-sensitive unique index
DROP INDEX IF EXISTS idx_m_vessel_code;

-- Step 3: Create case-insensitive unique index (only for non-deleted records)
CREATE UNIQUE INDEX idx_m_vessel_code ON m_vessels(UPPER(code)) WHERE deleted_at IS NULL;
