# ---- Build stage ----
FROM golang:latest AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# modernc.org/sqlite is pure Go (no CGO needed), so we can build a
# static-ish binary without a C toolchain in the runtime image.
RUN CGO_ENABLED=0 GOOS=linux go build -o /pulseguard-collector .

# ---- Runtime stage ----
FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates wget \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /pulseguard-collector /usr/local/bin/pulseguard-collector

# The collector writes its SQLite DB as "./pulseguard.db" (relative
# path, see internal/config/config.go). Running with WORKDIR=/data and
# mounting a volume there means the DB survives container restarts
# without needing to touch the Go code.
WORKDIR /data

EXPOSE 8080

# PULSEGUARD_SECRET is expected to come from a real environment
# variable (see docker-compose.yml) - not a .env file baked into the
# image. Never hardcode the secret into the image itself.
CMD ["pulseguard-collector"]
