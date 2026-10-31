-- Migration 015: Add company_id to m_offshore and m_geofences for multi-tenant isolation
ALTER TABLE m_offshore ADD COLUMN IF NOT EXISTS company_id bigint REFERENCES m_company(id) ON UPDATE CASCADE ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_m_offshore_company_id ON m_offshore(company_id);

ALTER TABLE m_geofences ADD COLUMN IF NOT EXISTS company_id bigint REFERENCES m_company(id) ON UPDATE CASCADE ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_m_geofences_company_id ON m_geofences(company_id);

-- Backfill existing HCML offshore assets and geofence corridor to HCML (company_id = 1)
UPDATE m_offshore SET company_id = 1 WHERE company_id IS NULL;
UPDATE m_geofences SET company_id = 1 WHERE company_id IS NULL;
