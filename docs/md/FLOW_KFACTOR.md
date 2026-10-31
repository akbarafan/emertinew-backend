# Flow Sensor & K-Factor Calibration

## Overview

Setiap kapal bisa memiliki hingga **4 flow sensor** (flow/1 – flow/4), masing-masing mengukur konsumsi bahan bakar. Data flow diterima via MQTT dalam bentuk **pulse count**, lalu dikonversi ke **liter** di backend menggunakan **K-Factor** kalibrasi per sensor.

## Formula

$$
\text{Volume (L)} = \frac{\text{pulse}}{\text{K-Factor}}
$$

| Parameter  | Satuan  | Keterangan                                          |
| ---------- | ------- | --------------------------------------------------- |
| `pulse`    | pulse   | Jumlah pulse terbaca dari flowmeter dalam interval  |
| `k_factor` | pulse/L | Faktor kalibrasi flowmeter (berapa pulse = 1 liter) |
| Volume     | liter   | Liter yang lewat meter pada **satu interval 5 detik** |

### Contoh Kalkulasi

| K-Factor (pulse/L) | Pulse Terbaca | Hasil (Liter) |
| ------------------ | ------------- | ------------- |
| 200                | 1000          | 5.0           |
| 450                | 900           | 2.0           |
| 0.1                | 5             | 50.0          |

## Satuan yang Disimpan: liter per menit

`pulse` di atas adalah **volume 5 detik** (`INTERVAL = 5` di `hardware/pi_atas_new/flow_njk.py`, dan `count` direset setiap baca). Satu baris `trackings` mewakili **satu menit**, jadi scheduler menjumlahkan seluruh sampel menit tersebut:

| Lapis | Nilai | Satuan |
| ----- | ----- | ------ |
| Sampel hardware | `pulse / k_factor` | liter per 5 detik |
| `trackings.flow1..4` | jumlah 12 sampel dalam menit itu | **liter per menit** |
| `L/h` (tampilan) | `flow × 60` | liter per jam |
| Total harian (`/fuel`) | `SUM(flow1..4)` | liter per hari |

Kalau satu baris hanya menyimpan sampel terakhir, nilainya jadi 1/12 dari volume sebenarnya — dan setiap total harian terlihat masuk akal padahal terlalu kecil. Itu sebabnya scheduler memakai `extractor.FlowAccumulate`.

## Arah Flow: `in`, `out`, dan "Single"

Kolom `m_sensors.flow_direction` diisi dari **MCU Management** (`/master/mcu`, section *Flow Config*), per sensor:

| Nilai | Arti |
| ----- | ---- |
| `in`  | Fuel yang **kembali ke tangki** (return line) |
| `out` | Fuel yang **keluar dari tangki** menuju mesin |
| `NULL` | **Single**: satu meter tanpa arah; angkanya langsung dipakai sebagai konsumsi |

Konsumsi sebuah engine:

| Channel yang dimiliki | Konsumsi |
| --------------------- | -------- |
| `out` + `in` | `max(0, out - in)` |
| `out` saja | `out` |
| Single saja | nilainya |
| `in` saja | tidak terukur (`-1`) |

> Urutannya penting: `out - in`, **bukan** `in - out`. `in - out` adalah saldo stok tangki, dan kalau dijepit di nol selalu jadi nol untuk kapal yang sedang berlabuh. `net_total` lama di API sudah diisi ulang dengan nilai `consumed`; field baru `consumed`, `fuel_returned`, dan `stock_change` disediakan di `internal/fuel/dto.go`.

## Konfigurasi K-Factor

K-Factor disimpan **per sensor** di tabel `m_sensors`, kolom `k_factor` (`DECIMAL(15,6)`, default `0`).

- **`k_factor = 0`** → Kalkulasi pulse dilewati (skip). Berguna saat sensor belum dikalibrasi.
- **`k_factor > 0`** → Backend akan menghitung `pulse / k_factor` dan menyimpan hasilnya dalam liter.

### Set K-Factor via API

**Create sensor:**

```bash
POST /master/sensor
{
  "mcu_id": 1,
  "topic": "hcml/vs001/er/flow/1",
  "name": "Flow Sensor Port Engine",
  "k_factor": 200
}
```

**Update K-Factor sensor yang sudah ada:**

```bash
PATCH /master/sensor/:id
{
  "k_factor": 450
}
```

**Response:**

```json
{
  "id": 5,
  "mcu_id": 1,
  "topic": "hcml/vs001/er/flow/1",
  "name": "Flow Sensor Port Engine",
  "k_factor": 450
}
```

## MQTT Payload

**Topic format:** `<company>/<vessel>/<location>/flow/<number>`

Contoh: `hcml/vs001/er/flow/1`

**Payload:**

```json
{
  "timestamp": "2026-04-18 14:30:00",
  "pulse": 1000
}
```

> `pulse = 0` → hasil = 0 liter (valid, artinya tidak ada aliran dalam interval tersebut).

## Logika Extractor

```
1. Tidak ada field "pulse"   → Skip                        ⏭️
2. k_factor <= 0             → Skip (belum dikalibrasi)    ⏭️
3. k_factor > 0              → Volume = pulse / k_factor   ✅
```

Hasil kalkulasi disimpan ke field `flow1`–`flow4` di tabel `trackings` berdasarkan `sensorNumber` dari topic MQTT. Pada scheduler, seluruh sampel menit itu dijumlahkan — lihat `extractor.FlowAccumulate`.

## Pipeline Data

```
Hardware (flowmeter)
  │ pulse count (direset tiap 5 detik)
  ▼
MQTT Broker
  │ {"pulse": 1000}
  ▼
mqtt.go ─── GetSensorByTopic() ─── ambil k_factor dari DB
  │
  ▼
┌─────────────────────────────────┐
│ Realtime Path                   │
│ service.CreateTracking()        │
│   → extractor.Flow(kFactor)    │  (sampel terakhir, liter per 5 detik)
│   → WebSocket broadcast        │
├─────────────────────────────────┤
│ Scheduler Path (per menit)      │
│ scheduler.createTrackingForMinute() │
│   → load sensors + kFactor map  │
│   → extractor.FlowAccumulate() │  (jumlah seluruh sampel menit itu)
│   → INSERT/UPDATE trackings    │
│   → WebSocket broadcast        │
└─────────────────────────────────┘
  │
  ▼
trackings.flow1–flow4 (liter per menit)
```

## File Terkait

| File                                         | Keterangan                                     |
| -------------------------------------------- | ---------------------------------------------- |
| `model/master.go`                            | Field `KFactor` di struct `MSensor`            |
| `internal/master/sensor/dto.go`              | Request/Response DTO dengan `k_factor`         |
| `internal/master/sensor/service.go`          | CRUD sensor termasuk `k_factor`                |
| `internal/master/sensor/mqtt.go`             | Ambil `k_factor` dari sensor, pass ke tracking |
| `internal/tracking/extractor/flow.go`        | Logika kalkulasi `pulse / k_factor` + akumulasi per menit |
| `internal/tracking/service.go`               | Pipeline realtime dengan `kFactor` param       |
| `internal/tracking/scheduler.go`             | Pipeline scheduler dengan `sensorKFactorMap`   |
| `internal/fuel/service.go`                   | Perhitungan `consumed = out - in` per engine   |
| `internal/master/sensor/validation.go`       | Validasi `flow_direction` & `k_factor`          |
| `migrations/007_add_k_factor_to_sensors.sql` | Migration kolom `k_factor`                     |
| `migrations/011_add_flow_engine_config_to_sensors.sql` | Migration kolom `flow_engine` & `flow_direction` |

## Catatan

- K-Factor berbeda tiap flowmeter — sesuaikan dengan spesifikasi/kalibrasi masing-masing unit. Nilai kapal yang tercatat: 36.34, 33.79, 24.97, 25.58 pulse/L (`FLOW_A_BIRU`, `FLOW_B_BIRU`, `FLOW_C_ORANGE`, `FLOW_D_ORANGE`).
- Hardware **wajib** mengirim field `pulse` dalam payload. Payload tanpa `pulse` akan di-skip.
- `k_factor` **wajib** di-set > 0 agar data flow dihitung. Sensor dengan `k_factor = 0` akan di-skip, dan API sekarang menolak membuat/memperbarui flow sensor tanpa kalibrasi.
- `pulse = 0` adalah valid — artinya tidak ada aliran, hasil = 0 liter.
- Perubahan `k_factor` via API langsung berlaku untuk data berikutnya (tidak retroaktif ke data lama).
- `flow_engine` dan `flow_direction` hanya boleh diisi pada sensor flow. "Single" berarti satu meter tanpa arah.
