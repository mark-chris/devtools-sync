# Production Dockerfile for dashboard
# Multi-stage build for minimal, secure final image

# Build stage
FROM node:20-alpine AS builder

WORKDIR /build

# Copy package files
COPY package*.json ./

# Install dependencies (including devDependencies for build)
RUN npm ci

# Copy source code
COPY . .

# Build production bundle
RUN npm run build

# Production dependencies only
RUN npm ci --omit=dev

# Final stage - distroless Node.js (debian13 for CVE-2026-0861 fix)
FROM gcr.io/distroless/nodejs20-debian13:nonroot

# Set working directory
WORKDIR /app

# Copy built application from builder
COPY --from=builder /build/dist /app/dist
COPY --from=builder /build/node_modules /app/node_modules
COPY --from=builder /build/package.json /app/package.json

# Use non-root user (provided by distroless)
USER nonroot:nonroot

EXPOSE 3000

# Serve the built application
CMD ["dist/index.js"]
