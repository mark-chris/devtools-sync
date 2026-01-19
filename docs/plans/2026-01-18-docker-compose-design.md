# Docker Compose Local Development Setup - Design Document

**Date:** 2026-01-18
**Issue:** #7 - Set up Docker Compose for local development

## Overview

Complete Docker Compose setup for DevTools Sync local development with PostgreSQL database, optional containerized server and dashboard services, hot reload support, and automatic database migrations.

## Architecture

### Services

1. **PostgreSQL** (always runs)
   - Base image: `postgres:16-alpine`
   - Port: 5432
   - Named volume for data persistence
   - Health checks ensure readiness before dependent services start

2. **Server** (optional, profile: "full")
   - Go API backend
   - Hot reload via Air
   - Auto-runs migrations on startup
   - Depends on healthy PostgreSQL

3. **Dashboard** (optional, profile: "full")
   - React/Vite frontend
   - Built-in HMR (Hot Module Replacement)
   - Depends on healthy Server

### Startup Modes

```bash
# Lightweight: Just PostgreSQL
docker compose -f docker-compose.dev.yml up

# Full stack: All services
docker compose -f docker-compose.dev.yml --profile full up
```

## Implementation Details

### Dockerfiles

#### server/Dockerfile.dev

- Base: `golang:1.21-alpine`
- Tools installed:
  - Air for hot reload
  - golang-migrate for database migrations
- Working directory: `/app`
- Volume strategy:
  - Mount `./server:/app` for source code
  - Exclude `/app/bin` to keep compiled binaries in container
- Exposed port: 8080
- Entrypoint: Air with pre-migration command

#### dashboard/Dockerfile.dev

- Base: `node:20-alpine`
- Working directory: `/app`
- Build-time: Copy package.json and run npm install
- Volume strategy:
  - Mount `./dashboard:/app` for source code
  - Exclude `/app/node_modules` to keep container's modules
- Exposed port: 5173
- Command: `npm run dev -- --host` (allows host access)

### Health Checks

**PostgreSQL:**
```yaml
test: ["CMD-SHELL", "pg_isready -U devtools -d devtools_sync"]
interval: 5s
timeout: 5s
retries: 5
```

**Server:**
```yaml
test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
interval: 10s
timeout: 5s
retries: 3
```
*Requires /health endpoint in Go server*

**Dashboard:**
```yaml
test: ["CMD", "curl", "-f", "http://localhost:5173"]
interval: 10s
timeout: 5s
```

### Dependency Chain

```
postgres (healthy) → server (healthy) → dashboard (running)
```

### Database Migrations

**Strategy:** Auto-run on server startup

**Implementation:**
- Air config includes `pre_cmd` that runs migrations before starting server
- Uses golang-migrate CLI: `migrate -path ./migrations -database $DATABASE_URL up`
- Fail-fast: If migrations fail, server won't start
- Migration output logged for debugging

### Environment Configuration

- Development secrets in `.env.local` (gitignored)
- Template provided in `env.example`
- Docker Compose reads `.env.local` automatically
- Safe development defaults (non-production secrets)

### Volume Mounts

**Benefits:**
- Instant code updates without rebuilding containers
- Build artifacts remain in container for performance
- Database data persists across container restarts

**Configuration:**
- Source code: Bidirectional mounts for hot reload
- Dependencies: Excluded from mounts (node_modules, Go binaries)
- Data: Named volume for PostgreSQL

## Files to Create/Modify

### New Files
- [ ] `server/Dockerfile.dev`
- [ ] `server/.air.toml` (Air configuration)
- [ ] `dashboard/Dockerfile.dev`
- [ ] `.env.local` (from env.example)

### Modified Files
- [ ] `docker-compose.dev.yml` (add health checks, update service configs)
- [ ] `development.md` (document Docker Compose usage)
- [ ] `server/cmd/main.go` or equivalent (add /health endpoint if missing)

## Testing Plan

1. **Database only:** `docker compose -f docker-compose.dev.yml up postgres`
   - Verify PostgreSQL starts and is healthy
   - Verify can connect from host

2. **Full stack:** `docker compose -f docker-compose.dev.yml --profile full up`
   - Verify services start in correct order
   - Verify migrations run automatically
   - Verify server /health endpoint responds
   - Verify dashboard accessible at http://localhost:5173

3. **Hot reload:**
   - Modify Go file in server → verify Air rebuilds and restarts
   - Modify React file in dashboard → verify Vite HMR updates browser

4. **Data persistence:**
   - Create test data in database
   - Stop and restart containers
   - Verify data persists

## Success Criteria

- ✅ `docker-compose.dev.yml` works out of the box
- ✅ Database migrations run automatically
- ✅ Services start in correct order (via health checks)
- ✅ Hot reload works for server (Air) and dashboard (Vite HMR)
- ✅ Documentation updated in development.md
- ✅ No Redis service (deferred to future if needed)

## Future Enhancements

- Add Redis service when caching requirements are defined
- Production docker-compose.yml with optimized builds
- Multi-stage Dockerfiles for smaller production images
- Docker Compose profiles for different development scenarios
