# Production Dockerfile for server
# Multi-stage build for minimal, secure final image
# Updated: 2026-02-03 - Rebuild with golang.org/x/crypto v0.47.0 security fixes

# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
# CGO_ENABLED=0 creates a fully static binary
# -ldflags="-w -s" strips debug info to reduce size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o devtools-sync-server \
    ./cmd

# Final stage - distroless
FROM gcr.io/distroless/static-debian12:nonroot

# Copy binary from builder
COPY --from=builder /build/devtools-sync-server /devtools-sync-server

# Copy migrations if they exist
COPY --from=builder /build/migrations /migrations

# Use non-root user (provided by distroless)
USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["/devtools-sync-server"]
