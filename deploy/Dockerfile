# syntax=docker/dockerfile:1

# ─── Build stage ─────────────────────────────────────────────────────────────
# go.mod declares `go 1.25.0`, so the toolchain has to match exactly or the
# module graph is rejected.
FROM golang:1.25-bookworm AS builder

WORKDIR /src

# Dependencies are resolved from the module files alone so this layer survives
# source edits.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO off keeps the binary fully static, which is what lets the runtime stage be
# a bare debian-slim with no Go runtime at all.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/emerti .

# ─── Runtime stage ───────────────────────────────────────────────────────────
FROM debian:bookworm-slim AS runtime

# ca-certificates — outbound TLS for the MQTT brokers.
# tzdata           — main.go calls time.LoadLocation("Asia/Jakarta"); without the
#                    zone database it silently falls back to UTC and every
#                    timestamp the API returns is off by seven hours.
# curl             — used by the container healthcheck.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata curl \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Jakarta \
    APP_PORT=6969

WORKDIR /app

COPY --from=builder /out/emerti /app/emerti

# `main -migrate` reads these from disk relative to the working directory.
COPY migrations /app/migrations

# Drop root. The app only ever writes to stdout and needs no filesystem access
# beyond the migrations it reads. A plain (non `--system`) account is used
# because uid 10001 exceeds SYS_UID_MAX and would only emit a warning.
RUN useradd --uid 10001 --create-home --shell /usr/sbin/nologin emerti
USER emerti

EXPOSE 6969

# The compose stack overrides this to run `-migrate` and `-seed` in throwaway
# containers from the same image.
ENTRYPOINT ["/app/emerti"]
CMD []