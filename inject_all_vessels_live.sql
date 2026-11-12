-- =============================================================================
-- INJECT LIVE TELEMETRY UNTUK 6 KAPAL (FLEET HCML / SELAT MADURA)
-- Kapal: 321, 319, 320, 322, 323, 324
-- Timestamp: NOW() - INTERVAL '2 minute' (agar langsung terbaca LIVE oleh backend)
--
-- PEMETAAN SENSOR DATABASE (m_sensors):
--   Port Main Engine      : flow1 (IN/Supply),  flow2 (OUT/Return), rpm1 & rpm2 (Tachometer)
--   Starboard Main Engine : flow3 (IN/Supply),  flow4 (OUT/Return), rpm3 & rpm4 (Tachometer)
--   Rumus Flow Netto      : (flow_in - flow_out) * 60 = L/jam (L/h)
--
-- Cara Eksekusi di Terminal Docker:
--   docker exec -i emerti-cc-db psql -U emerti -d emerti_db < backend/inject_all_vessels_live.sql
-- =============================================================================

BEGIN;

INSERT INTO public.trackings (
  "time", vessel_id, lat, lon, speed, heading,
  rpm1, rpm2, rpm3, rpm4,
  flow1, flow2, flow3, flow4,
  top_ac_volt, top_ac_amp, top_ac_freq, top_ac_power,
  top_dc_volt, top_dc_amp, top_bat_volt, top_bat_amp,
  bot_ac_volt, bot_ac_amp, bot_ac_freq, bot_ac_power,
  bot_dc_volt, bot_dc_amp, bot_bat_volt, bot_bat_amp,
  top_door1, top_door2, bot_door1, bot_door2,
  top_temp, bot_temp, accel_z
) VALUES
-- -----------------------------------------------------------------------------
-- 1. [ID 321] MV Tectonic Forge (AHT. LIMIN) - Anchor Handling Tug
-- Skenario: Full Speed Cruising transit Selat Madura menuju MAC Field
-- Kecepatan: 11.7 knots | Heading: 122.0° | Mode: Full Speed
-- Target Bench per ME: 70.0 L/h (140 L/h total)
-- Port ME : RPM 1425 | Flow: (2.14 - 1.00)*60 = 68.4 L/h (Deviasi: -2.3% NORMAL)
-- Stbd ME : RPM 1418 | Flow: (2.19 - 1.00)*60 = 71.4 L/h (Deviasi: +2.0% NORMAL)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 321, -7.2620, 113.8420, 11.7, 122.0,
  1425, 1425, 1418, 1418,
  2.14, 1.00, 2.19, 1.00,
  221.5, 12.8, 50.0, 2.83,
  24.3, 5.3, 24.5, 2.0,
  382.0, 29.4, 50.0, 11.20,
  24.2, 8.6, 24.4, 3.0,
  false, false, false, false,
  25.5, 39.5, 1.00
),

-- -----------------------------------------------------------------------------
-- 2. [ID 319] MV Abyssal Mechanica (CB Marine Hawk) - Crew Boat
-- Skenario: Standby / Berlabuh Jangkar di Area Anchorage Tanjung Perak Surabaya
-- Kecepatan: 0.2 knots | Heading: 045.0° | Mode: Standby / Idle
-- Target Bench per ME: 12.0 L/h
-- Port ME : RPM 550 | Flow: (0.59 - 0.40)*60 = 11.4 L/h (Deviasi: -5.0% NORMAL)
-- Stbd ME : RPM 540 | Flow: (0.61 - 0.40)*60 = 12.6 L/h (Deviasi: +5.0% NORMAL)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 319, -7.1950, 112.7310, 0.2, 45.0,
  550, 550, 540, 540,
  0.59, 0.40, 0.61, 0.40,
  220.2, 10.5, 50.0, 2.31,
  24.2, 4.8, 24.4, 1.8,
  380.0, 22.1, 50.0, 8.40,
  24.1, 7.5, 24.3, 2.5,
  false, false, false, false,
  26.5, 35.0, 1.00
),

-- -----------------------------------------------------------------------------
-- 3. [ID 320] MV Cybernetic Triton (NMS ACHIEVE) - Fast Utility / Crew Supply
-- Skenario: Cruising transit Selat Madura menuju BD Field
-- Kecepatan: 8.8 knots | Heading: 098.0° | Mode: Full Speed / Cruising
-- Target Bench per ME: 70.0 L/h
-- Port ME : RPM 1120 | Flow: (2.14 - 1.00)*60 = 68.4 L/h (Deviasi: -2.3% NORMAL)
-- Stbd ME : RPM 1125 | Flow: (2.19 - 1.00)*60 = 71.4 L/h (Deviasi: +2.0% NORMAL)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 320, -7.3450, 113.0500, 8.8, 98.0,
  1120, 1120, 1125, 1125,
  2.14, 1.00, 2.19, 1.00,
  221.0, 11.6, 50.0, 2.56,
  24.4, 5.1, 24.6, 1.9,
  381.5, 26.2, 50.0, 9.98,
  24.3, 8.1, 24.5, 2.8,
  false, false, false, false,
  25.8, 37.2, 1.00
),

-- -----------------------------------------------------------------------------
-- 4. [ID 322] MV Subaquatic Vector (CB ALDWIN) - Crew Boat
-- Skenario: Maneuvering (Olah Gerak merapat ke Platform BD WHP / FPSO)
-- Kecepatan: 2.2 knots | Heading: 185.0° | Mode: Maneuvering
-- Target Bench per ME: 15.0 L/h (30 L/h total)
-- Port ME : RPM 750 | Flow: (0.64 - 0.40)*60 = 14.4 L/h (Deviasi: -4.0% NORMAL)
-- Stbd ME : RPM 740 | Flow: (0.66 - 0.40)*60 = 15.6 L/h (Deviasi: +4.0% NORMAL)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 322, -7.3680, 113.2720, 2.2, 185.0,
  750, 750, 740, 740,
  0.64, 0.40, 0.66, 0.40,
  220.8, 11.2, 50.0, 2.47,
  24.1, 5.0, 24.3, 1.9,
  380.8, 24.5, 50.0, 9.33,
  24.0, 7.9, 24.2, 2.7,
  false, false, false, false,
  26.0, 36.5, 1.00
),

-- -----------------------------------------------------------------------------
-- 5. [ID 323] MV Hydro-Kinetic Prime (OPS ALHENA) - OSV / Platform Supply
-- Skenario: Towing / Heavy Support di sekitar FPU Trunojoyo 01 (MBH Field)
-- Kecepatan: 5.2 knots (beban tarik) | Heading: 265.0° | Mode: Towing
-- Target Bench per ME: 72.5 L/h (145 L/h total)
-- Port ME : RPM 1380 | Flow: (2.18 - 1.00)*60 = 70.8 L/h (Deviasi: -2.3% NORMAL)
-- Stbd ME : RPM 1375 | Flow: (2.24 - 1.00)*60 = 74.4 L/h (Deviasi: +2.6% NORMAL)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 323, -7.3150, 114.3020, 5.2, 265.0,
  1380, 1380, 1375, 1375,
  2.18, 1.00, 2.24, 1.00,
  221.8, 13.5, 50.0, 2.99,
  24.5, 5.6, 24.7, 2.1,
  383.0, 31.0, 50.0, 11.87,
  24.4, 9.0, 24.6, 3.2,
  false, false, false, false,
  25.2, 40.1, 1.00
),

-- -----------------------------------------------------------------------------
-- 6. [ID 324] MV Hydroterra Vanguard (AHT KYLIE S) - AHTS
-- Skenario: Safety Standby Guard Vessel di MAC Field Safety Zone
-- Kecepatan: 0.1 knots | Heading: 310.0° | Mode: ME Stop (Off / Standby)
-- Target Bench per ME: 0 L/h (ME Stop / Moored)
-- Port ME : RPM 0 | Flow: 0.00 L/h (Status: ME Tidak Aktif / NORMAL)
-- Stbd ME : RPM 0 | Flow: 0.00 L/h (Status: ME Tidak Aktif / NORMAL)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 324, -7.3820, 113.8400, 0.1, 310.0,
  0, 0, 0, 0,
  0.00, 0.00, 0.00, 0.00,
  220.5, 9.8, 50.0, 2.16,
  24.2, 4.5, 24.4, 1.7,
  380.0, 20.5, 50.0, 7.79,
  24.1, 7.2, 24.3, 2.4,
  false, false, false, false,
  27.0, 34.0, 1.00
);

COMMIT;
