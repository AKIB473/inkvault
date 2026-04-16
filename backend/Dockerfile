# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Install build deps
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

# Build with security flags: strip debug info, disable CGO
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" \
    -trimpath \
    -o bin/inkvault-server \
    ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -ldflags="-w -s" \
    -trimpath \
    -o bin/inkvault-migrate \
    ./cmd/migrate

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
FROM scratch

# Non-root user (security hardening — Ghost does this too)
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/bin/inkvault-server /inkvault-server
COPY --from=builder /app/bin/inkvault-migrate /inkvault-migrate

# Run as nobody (uid 65534)
USER 65534:65534

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/inkvault-server", "-health"]

ENTRYPOINT ["/inkvault-server"]
