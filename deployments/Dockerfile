# =============================================================================
# Hotel Management System API — Multi-Stage Dockerfile
# Stage 1: Builder  (golang:1.23-alpine)
# Stage 2: Runner   (alpine)
# =============================================================================

# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.23-alpine AS builder

# Install build dependencies for CGO (needed by go-sqlite3 and mattn libs)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy dependency manifests first so Docker can cache the module download layer
COPY go.mod go.sum ./
RUN go mod download

# Copy the full source tree
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /app/hotel-api ./cmd/server/main.go

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
FROM alpine:3.20

# Install CA certificates (for TLS/HTTPS outbound calls) and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy compiled binary from builder
COPY --from=builder /app/hotel-api .

# Copy static assets needed at runtime
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/envs ./envs

# Create logs directory so the logger can write to it inside the container
RUN mkdir -p /app/logs

# Expose the default API port
EXPOSE 8080

# Run as a non-root user for security
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Pass the -env flag; override with DOCKER_ENV build-arg or CMD override
ENTRYPOINT ["./hotel-api"]
CMD ["-env", "dev"]
