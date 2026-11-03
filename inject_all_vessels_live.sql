-- =============================================================================
-- INJECT LIVE TELEMETRY UNTUK 6 KAPAL (FLEET HCML / SELAT MADURA)
-- Kapal: 321, 319, 320, 322, 323, 324
-- Timestamp: NOW() - INTERVAL '2 minute' (agar langsung terbaca LIVE oleh backend)
--
-- Cara Eksekusi di Terminal Docker:
-- docker exec -i emerti-db psql -U user -d emerti_db < backend/inject_all_vessels_live.sql
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
-- Kecepatan: 11.7 knots | RPM: 1425 / 1418 | Flow ME: ~69 L/h (1.15 L/min)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 321, -7.2620, 113.8420, 11.7, 122.0,
  1425, 1418, 1500, NULL,
  1.15, 1.14, 0.20, NULL,
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
-- Kecepatan: 0.2 knots | RPM: 550 / 540 (Idle) | Flow ME: ~6.0 L/h (0.10 L/min)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 319, -7.1950, 112.7310, 0.2, 45.0,
  550, 540, 1500, NULL,
  0.10, 0.10, 0.20, NULL,
  220.2, 10.5, 50.0, 2.31,
  24.2, 4.8, 24.4, 1.8,
  380.0, 22.1, 50.0, 8.40,
  24.1, 7.5, 24.3, 2.5,
  false, false, false, false,
  26.5, 35.0, 1.00
),

-- -----------------------------------------------------------------------------
-- 3. [ID 320] MV Cybernetic Triton (NMS ACHIEVE) - Fast Utility / Crew Supply
-- Skenario: Economical Speed (Eco Sailing) membawa suplai logistik ke BD Field
-- Kecepatan: 8.8 knots | RPM: 1120 / 1125 | Flow ME: ~42.3 L/h (0.705 L/min)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 320, -7.3450, 113.0500, 8.8, 98.0,
  1120, 1125, 1500, NULL,
  0.70, 0.71, 0.20, NULL,
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
-- Kecepatan: 3.5 knots | RPM: 780 / 760 | Flow ME: ~14.7 L/h (0.245 L/min)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 322, -7.3680, 113.2720, 3.5, 185.0,
  780, 760, 1500, NULL,
  0.25, 0.24, 0.20, NULL,
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
-- Kecepatan: 5.2 knots | RPM: 1380 / 1375 | Flow ME: ~62.7 L/h (1.045 L/min)
-- (RPM & Beban Mesin tinggi untuk menarik/menahan beban, kecepatan rendah)
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 323, -7.3150, 114.3020, 5.2, 265.0,
  1380, 1375, 1500, NULL,
  1.05, 1.04, 0.25, NULL,
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
-- Kecepatan: 0.1 knots | RPM: 0 / 0 (ME Stop, Genset AE On) | Flow ME: 0 L/h
-- -----------------------------------------------------------------------------
(
  NOW() - INTERVAL '2 minute', 324, -7.3820, 113.8400, 0.1, 310.0,
  0, 0, 1500, NULL,
  0.00, 0.00, 0.22, NULL,
  220.5, 9.8, 50.0, 2.16,
  24.2, 4.5, 24.4, 1.7,
  380.0, 20.5, 50.0, 7.79,
  24.1, 7.2, 24.3, 2.4,
  false, false, false, false,
  27.0, 34.0, 1.00
);

COMMIT;
