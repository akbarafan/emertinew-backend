-- Fuel types and the vessel's fuel type.
--
-- The emissions figures on the command center are only as good as the factor
-- behind them, so the factor lives in data, not in code, and every column
-- keeps the reference it came from.
--
-- Two scopes are supported, because they answer different questions:
--
--   tank-to-wake (TtW)  CO2 emitted between the tank and the propeller.
--                       This is the regulatory figure (IMO DCS / MARPOL
--                       Annex VI) and the one the engine meters can support.
--   well-to-wake (WtW)  the full life cycle, including the fuel chain
--                       upstream (MEPC.391(81), 2024 LCA Guidelines).
--
-- The stored parameters are the raw ones from the IMO documents; the per-litre
-- factors are derived in the application so they can never drift from their
-- inputs:
--
--   TtW CO2    [kg/L]  = cf_co2                                   x density
--   TtW CO2e   [kg/L]  = (cf_co2 + 28*cf_ch4 + 265*cf_n2o + slip) x density
--   WtW CO2e   [kg/L]  = (wtt_intensity x lhv + TtW CO2e)          x density
--
-- GWP100 values (CH4 = 28, N2O = 265) are the ones the 2024 LCA Guidelines
-- state in paragraph 2.4.
--
-- A NULL density or a NULL wtt intensity means the figure is not derivable
-- from an IMO reference, and the application then reports it as unavailable
-- rather than guessing. The density can always be filled in from the bunker
-- delivery note of the fuel actually supplied.

CREATE TABLE IF NOT EXISTS m_fuel_types (
    id            BIGSERIAL PRIMARY KEY,
    -- The unique constraint carries GORM's name, because the auto-migration on
    -- startup drops and recreates it by that name.
    code          VARCHAR(50)  NOT NULL CONSTRAINT uni_m_fuel_types_code UNIQUE,
    name          VARCHAR(255) NOT NULL,

    -- Carbon factors, tank-to-wake. IMO DCS / MEPC.364(79) Annex 9.
    cf_co2_kg_per_kg  NUMERIC(12,6),
    cf_ch4_kg_per_kg  NUMERIC(12,8),
    cf_n2o_kg_per_kg  NUMERIC(12,8),
    -- Unburned methane escaping the engine, in mass % of the fuel. Only
    -- meaningful for gaseous fuels; MEPC.391(81) Appendix 2 lists it per
    -- energy converter (LNG Otto medium speed: 3.5 %).
    methane_slip_percent NUMERIC(6,3),

    -- Well-to-tank intensity in gCO2eq per MJ of fuel energy, and the lower
    -- heating value in MJ per kg, both from MEPC.391(81) Appendix 2. The
    -- product gives the well-to-tank part in gCO2eq per kg of fuel.
    wtt_intensity_gco2eq_per_mj NUMERIC(12,4),
    lhv_mj_per_kg               NUMERIC(10,3),

    -- Density in kg per litre. Fourth IMO GHG Study 2020, Table 10.
    density_kg_per_litre NUMERIC(8,4),

    -- Where the numbers above come from, so an auditor can trace a figure back
    -- to the standard it was taken from.
    reference VARCHAR(255),
    method    VARCHAR(100),

    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

COMMENT ON TABLE m_fuel_types IS 'Fuel types with the IMO emission parameters used for the fleet emissions figures';
COMMENT ON COLUMN m_fuel_types.cf_co2_kg_per_kg IS 'IMO DCS carbon factor, kg CO2 per kg fuel (MEPC.364(79))';
COMMENT ON COLUMN m_fuel_types.wtt_intensity_gco2eq_per_mj IS 'Well-to-tank GHG intensity, gCO2eq per MJ (MEPC.391(81) Appendix 2). NULL when the IMO has no default for this fuel';
COMMENT ON COLUMN m_fuel_types.density_kg_per_litre IS 'Density, kg per litre (Fourth IMO GHG Study 2020 Table 10). NULL when not published';
COMMENT ON COLUMN m_fuel_types.reference IS 'Source of the parameters in this row';

-- The fuel the vessel runs on. NULL means "not declared", and the vessel then
-- has no derivable emission figure.
ALTER TABLE m_vessels
    ADD COLUMN IF NOT EXISTS fuel_type_id BIGINT;

CREATE INDEX IF NOT EXISTS idx_m_vessels_fuel_type_id ON m_vessels (fuel_type_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_m_vessels_fuel_type'
    ) THEN
        ALTER TABLE m_vessels
            ADD CONSTRAINT fk_m_vessels_fuel_type
                FOREIGN KEY (fuel_type_id) REFERENCES m_fuel_types (id)
                ON UPDATE CASCADE ON DELETE SET NULL;
    END IF;
END $$;

-- Initial fuel list.
--
-- The carbon factors and heating values are the IMO DCS values; the
-- well-to-tank intensities are the ones MEPC.391(81) Appendix 2 publishes (HFO
-- and MDO only); the densities are the ones the Fourth IMO GHG Study used.
-- Everything left NULL is filled in from a bunker delivery note or from a life
-- cycle study, and stays NULL until then.
INSERT INTO m_fuel_types (
    code, name, cf_co2_kg_per_kg, cf_ch4_kg_per_kg, cf_n2o_kg_per_kg,
    methane_slip_percent, wtt_intensity_gco2eq_per_mj, lhv_mj_per_kg,
    density_kg_per_litre, reference, method, created_at, updated_at
) VALUES
    (
        'MGO', 'Marine Gas Oil (MDO/MGO)',
        3.206, 0.00005, 0.00018, NULL, 17.7, 42.700, 0.895,
        'MEPC.364(79) Annex 9; MEPC.391(81) Appendix 2; Fourth IMO GHG Study 2020 Table 10',
        'IMO DCS / IMO LCA 2024', now(), now()
    ),
    (
        'HFO', 'Heavy Fuel Oil (VLSFO)',
        3.114, 0.00005, 0.00018, NULL, 16.8, 40.200, 1.001,
        'MEPC.364(79) Annex 9; MEPC.391(81) Appendix 2 (VLSFO); Fourth IMO GHG Study 2020 Table 10',
        'IMO DCS / IMO LCA 2024', now(), now()
    ),
    (
        'LFO', 'Light Fuel Oil (ULSFO)',
        3.151, 0.00005, 0.00018, NULL, NULL, 41.200, 0.895,
        'MEPC.364(79) Annex 9; density as MDO (distillate) per Fourth IMO GHG Study 2020 Table 10',
        'IMO DCS', now(), now()
    ),
    (
        'LNG', 'Liquefied Natural Gas (Otto, medium speed)',
        2.750, 0.00000, 0.00011, 3.5, NULL, 48.000, 0.450,
        'MEPC.364(79) Annex 9; MEPC.391(81) Appendix 2 (methane slip 3.5 %); Fourth IMO GHG Study 2020 Table 10',
        'IMO DCS', now(), now()
    ),
    (
        'LPG_PROPANE', 'Liquefied Petroleum Gas (Propane)',
        3.000, 0.00005, 0.00018, NULL, NULL, 46.300, NULL,
        'MEPC.364(79) Annex 9; density to be filled from the bunker delivery note',
        'IMO DCS', now(), now()
    ),
    (
        'LPG_BUTANE', 'Liquefied Petroleum Gas (Butane)',
        3.030, 0.00005, 0.00018, NULL, NULL, 45.700, NULL,
        'MEPC.364(79) Annex 9; density to be filled from the bunker delivery note',
        'IMO DCS', now(), now()
    ),
    (
        'METHANOL', 'Methanol',
        1.375, NULL, NULL, NULL, NULL, 19.900, 0.790,
        'MEPC.364(79) Annex 9; Fourth IMO GHG Study 2020 Table 10',
        'IMO DCS', now(), now()
    ),
    (
        'ETHANOL', 'Ethanol',
        1.913, NULL, NULL, NULL, NULL, 26.800, NULL,
        'MEPC.364(79) Annex 9; density to be filled from the bunker delivery note',
        'IMO DCS', now(), now()
    )
ON CONFLICT (code) DO UPDATE SET
    name          = EXCLUDED.name,
    cf_co2_kg_per_kg  = EXCLUDED.cf_co2_kg_per_kg,
    cf_ch4_kg_per_kg  = EXCLUDED.cf_ch4_kg_per_kg,
    cf_n2o_kg_per_kg  = EXCLUDED.cf_n2o_kg_per_kg,
    methane_slip_percent = EXCLUDED.methane_slip_percent,
    wtt_intensity_gco2eq_per_mj = EXCLUDED.wtt_intensity_gco2eq_per_mj,
    lhv_mj_per_kg  = EXCLUDED.lhv_mj_per_kg,
    density_kg_per_litre = EXCLUDED.density_kg_per_litre,
    reference      = EXCLUDED.reference,
    method         = EXCLUDED.method,
    updated_at     = now();

-- The fleet runs on marine gas oil until the master data says otherwise. The
-- attribute can be corrected per vessel at any time; the figures follow.
UPDATE m_vessels v
SET fuel_type_id = (SELECT id FROM m_fuel_types WHERE code = 'MGO'),
    updated_at   = now()
WHERE v.fuel_type_id IS NULL;
