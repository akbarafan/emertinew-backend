# Example Dockerfile for a Go backend application
# Stage 1: build
FROM golang:1.24.5 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: runtime
FROM debian:bookworm-slim

WORKDIR /app

# Copy hanya binary dari builder
COPY --from=builder /app/main .

EXPOSE 1234

CMD ["./main"]
