# EMERTI Command Center — backend stack

Self-contained Docker stack for the backend that backs the Command Center
frontend. **Completely separate from the legacy `docker-compose.yml` in the
backend root** — different Compose project, container names, named volumes,
network and host ports, so both stacks can run side by side on one machine.

```
| stack        | project      | containers                                  | host ports                    |
|--------------|--------------|---------------------------------------------|-------------------------------|
| legacy       | emerti-fms   | emerti-fms-db, emerti-fms-redis             | 6666, 6379                    |
| this one     | emerti-cc    | emerti-cc-db, emerti-cc-redis, emerti-cc-api | 6667, 6380, 6969               |
```

## Files

| File | Purpose |
|---|---|
| `Dockerfile` | Multi-stage build. `golang:1.25-bookworm` → static `debian:bookworm-slim`, non-root |
| `docker-compose.yml` | Five services: `db`, `redis`, `migrate`, `seed`, `api` |
| `.env.example` | Documented template — copy to `.env` (gitignored) to override |
| `.env` | Working local values (gitignored, never committed) |
| `../.dockerignore` | Build-context filters (must live at the context root, i.e. `backend/`) |

## Services

| Service | Image | Role | Lifetime |
|---|---|---|---|
| `db` | `timescale/timescaledb:2.17.2-pg17` | TimescaleDB / PostgreSQL 17 | long-running |
| `redis` | `redis:8-alpine` | Session store, password-protected | long-running |
| `migrate` | `emerti-cc-api:latest` | runs `main -migrate` over `migrations/*.sql` | one-shot, exits 0 |
| `seed` | `emerti-cc-api:latest` | runs `main -seed` (master data) | one-shot, exits 0 |
| `api` | `emerti-cc-api:latest` | Gin HTTP server + MQTT + tracking scheduler | long-running |

Startup order is gated by real health checks, never by sleeps:

```
db healthy ─┬─> migrate (exits 0) ─> seed (exits 0) ─> api
redis healthy┘
```

`migrate` needs only Postgres — `main.go` returns from the `-migrate` branch
*before* `InitRedis`. `seed` runs *after* Redis initialisation, where a failure
is fatal, so it depends on both.

## Quick start

`deploy/.env` is optional. Every value in `docker-compose.yml` carries a
`${VAR:-default}`, so the stack boots with sensible defaults out of the box; the
file exists only to override them. It is matched by the backend `.gitignore` rule
for `.env`, so real credentials never reach git — `.env.example` is the
committed template.

```powershell
cd D:\bar\emertinew\backend

docker compose -f deploy/docker-compose.yml up -d --build
```

Watch it come up:

```powershell
docker compose --env-file deploy/.env -f deploy/docker-compose.yml logs -f api
```

Verify:

```powershell
Invoke-RestMethod http://localhost:6969/            # health
Invoke-RestMethod http://localhost:6969/api/auth/login -Method Post `
  -ContentType 'application/json' `
  -Body '{"username":"superadmin","password":"rahasia123"}'
```

`docker compose -f deploy/docker-compose.yml <cmd>` is enough on its own — Compose
reads `deploy/.env` automatically because it sits next to the compose file. Add
`--env-file` only when pointing at a different file, e.g. a production override.

## Everyday commands

```powershell
# status
docker compose -f deploy/docker-compose.yml ps -a

# logs
docker compose -f deploy/docker-compose.yml logs -f api
docker logs emerti-cc-seed        # one-shot jobs

# re-run the seed on its own (safe to repeat)
docker compose -f deploy/docker-compose.yml run --rm seed

# re-run SQL migrations on their own
docker compose -f deploy/docker-compose.yml run --rm migrate

# restart just the API
docker compose -f deploy/docker-compose.yml restart api

# stop, keep data
docker compose -f deploy/docker-compose.yml stop

# stop and remove containers + network, KEEP volumes
docker compose -f deploy/docker-compose.yml down

# stop and DESTROY all data
docker compose -f deploy/docker-compose.yml down -v
```

`make cc-up` / `make cc-down` / `make cc-logs` wrap the same commands.

## Frontend wiring

The API is published on **6969**, which is exactly what the frontend already
expects:

```
NEXT_PUBLIC_API_BASE_URL=http://localhost:6969/api
```

No frontend configuration change is needed. If you remap `API_HOST_PORT`, change
that variable too.

## Seeded accounts

| Username | Password | Role |
|---|---|---|
| `superadmin` | `rahasia123` | superadmin |
| `admin` | `pr4m1st1` | superadmin |
| `logistic` | `logistic` | admin |
| `hcml_guest` | `hcml_guest` | guest |

Change these before any deployment that is not a local sandbox.

## What the seed creates

- 4 users, 1 company (`hcml`)
- 4 classifications, 5 vessel types, 3 operation statuses, 3 material hulls
- 6 test vessels (`TS001`–`TS006`) with bridge-deck and engine-room MCUs
- 6 production vessels (`VS001`–`VS006`) with MCUs and sensors
- 288 sensors total (MQTT topics under `hcml/<vessel-code>/bd|er/<group>/<n>`)
- 11 offshore fields, 2 ports, 1 geofence (`Jalur Pelayaran Madura`, 10 waypoints)

Re-running the seed is safe: every insert is guarded by a `WHERE name = ?`
existence check and logs `Skip … exists`. See "Known issues" for the one
exception.

## Known issues

Both are pre-existing backend bugs, not container problems. They do not prevent
the stack from starting or serving requests.

### 1. Six seeded vessels fail with a primary-key collision

`seed/vessel.go` defines a `testVessels` slice with explicit ids `1`–`6`, then a
second `vessels` slice whose first six entries **also** hardcode ids `1`–`6`
(codes `VS100`–`VS105`, names `Test Alugara 1`–`6`). The second batch therefore
always hits `duplicate key value violates unique constraint "m_vessels_pkey"`.

The existence check on line 134 queries by `name`, which does not exist yet, so
it proceeds to insert and collides. Result: 12 of 18 vessels are created.

Fix options, in order of least surprise:

1. Drop the hardcoded `ID` from those six entries and let the sequence assign
   one — GORM omits a zero-value primary key and writes the generated value back
   into the struct, so the `vessel.ID` used by `assignMCUsToVessel` stays correct.
2. Move the `Test Alugara` rows into their own slice with non-colliding ids.
3. Delete the rows if they are leftovers from an older schema.

They are all `IsActive: false`, so no MCUs or sensors are attached to them
either way — only the vessel rows are missing.

### 2. `trackings` never becomes a TimescaleDB hypertable

`config/database.go` logs `Failed to create hypertable for tracking`. TimescaleDB
refuses to convert a table that already carries a unique index not including the
partitioning column, and `trackings_pkey ON (id)` is exactly that.

`sensor_data` converts fine because its hypertable is created before the primary
key index is in place. Only `sensor_data` appears in
`timescaledb_information.hypertables`; `trackings` stays a plain table.

Impact is performance, not correctness: telemetry still stores and reads fine, it
just does not get partitioning, compression or retention policies.

### 3. Two SQL migrations report as failed on a fresh database

`010_add_is_visible_to_sensors.sql` and `011_add_flow_engine_config_to_sensors.sql`
report `column … already exists`, so `-migrate` prints `7/9 successful`. This is
expected: GORM's `AutoMigrate` already creates `is_visible` and `flow_engine`
from the `MSensor` struct before any SQL file runs, so those two files are
redundant on an empty database. No action needed.

### 4. `mqtt_status` in the health payload is hardcoded

`routes/routes.go` returns a literal `"mqtt_status": "connected"` and a stale
`192.168.18.88:1883` broker address, regardless of actual broker state. Check
`docker logs emerti-cc-api` for the real `[MQTT-OLD]` / `[MQTT-NEW]` lines
instead.

## MQTT

`MQTT_BROKER` / `MQTT2_BROKER` point at external brokers, not containers in this
stack. If a broker is unreachable the API still starts and retries the
connection in the background, so a bad value degrades rather than breaks the
service. At the time of writing, `MQTT-NEW` (103.150.194.254) connects and
subscribes 288 topics; `MQTT-OLD` (103.103.20.19) refuses the connection.

To run without MQTT at all, set `MQTT_BROKER=` and `MQTT2_BROKER=` to empty
strings in `deploy/.env`.

## Why the Go toolchain is pinned to 1.25

`go.mod` declares `go 1.25.0`. A lower base image (the legacy `Dockerfile` in the
backend root still uses `golang:1.24.5`) is rejected by the module loader, so
that file cannot build this tree as it stands.

## Timezone

`main.go` calls `time.LoadLocation("Asia/Jakarta")`. Without the tzdata package
in the runtime image that call fails and the process silently falls back to UTC,
which shifts every timestamp the API returns by seven hours. The runtime stage
installs `tzdata`, the database runs with `TZ=Asia/Jakarta` and
`-c timezone=Asia/Jakarta`, and the image sets `ENV TZ=Asia/Jakarta`.