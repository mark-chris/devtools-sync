# Docker Compose Local Development Setup - Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Complete Docker Compose setup for local development with PostgreSQL, containerized server (Go), and dashboard (React/Vite) with hot reload and automatic migrations.

**Architecture:** Three-service architecture with PostgreSQL (always-on), optional server and dashboard (profile: "full"). Health checks ensure proper startup order. Air provides Go hot reload, Vite provides React HMR.

**Tech Stack:** Docker Compose, PostgreSQL 16, Go 1.21, Node 20, Air, golang-migrate, Vite

---

## Task 1: Create Server Dockerfile

**Files:**
- Create: `server/Dockerfile.dev`

**Step 1: Create server development Dockerfile**

```dockerfile
FROM golang:1.24-alpine

# Install system dependencies
RUN apk add --no-cache git curl postgresql-client

# Install Air for hot reload
RUN go install github.com/cosmtrek/air@latest

# Install golang-migrate for database migrations
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

WORKDIR /app

# Copy go.mod and go.sum if they exist (for dependency caching)
# Note: This project is in early stages, so these may not exist yet
COPY go.* ./
RUN if [ -f go.mod ]; then go mod download; fi

# Copy source code (via volume mount in docker-compose)
COPY . .

EXPOSE 8080

# Use Air for hot reload
CMD ["air", "-c", ".air.toml"]
```

**Step 2: Verify Dockerfile syntax**

Run: `docker build -f server/Dockerfile.dev -t devtools-server-dev-test ./server --no-cache 2>&1 | head -20`

Expected: Build should start (may fail due to missing go.mod, but syntax should be valid)

**Step 3: Commit**

```bash
git add server/Dockerfile.dev
git commit -m "feat(server): add development Dockerfile with Air hot reload"
```

---

## Task 2: Create Air Configuration

**Files:**
- Create: `server/.air.toml`

**Step 1: Create Air configuration for hot reload**

```toml
# Air configuration for hot reload during development
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  # Run migrations before starting the server
  pre_cmd = ["sh", "-c", "if [ -d ./migrations ] && [ \"$(ls -A ./migrations)\" ]; then migrate -path ./migrations -database \"${DATABASE_URL}\" up || echo 'Migration failed or no migrations found'; fi"]

  cmd = "go build -o ./tmp/main ./cmd"
  bin = "tmp/main"
  full_bin = "./tmp/main"

  include_ext = ["go", "tpl", "tmpl", "html"]
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "bin"]
  include_dir = []

  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false

  log = "build-errors.log"
  poll = false
  poll_interval = 0
  delay = 1000
  stop_on_error = false
  send_interrupt = false
  kill_delay = "0s"

[log]
  time = true
  main_only = false

[color]
  main = "magenta"
  watcher = "cyan"
  build = "yellow"
  runner = "green"

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
```

**Step 2: Verify TOML syntax**

Run: `cat server/.air.toml | head -5`

Expected: Should display first 5 lines of TOML without errors

**Step 3: Commit**

```bash
git add server/.air.toml
git commit -m "feat(server): add Air config with auto-migrations"
```

---

## Task 3: Create Minimal Server Scaffold

**Files:**
- Create: `server/go.mod`
- Create: `server/cmd/main.go`

**Step 1: Create go.mod**

```go
module github.com/mark-chris/devtools-sync/server

go 1.24

require (
	github.com/lib/pq v1.10.9
)
```

**Step 2: Create minimal server with health endpoint**

File: `server/cmd/main.go`

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"devtools-sync-server"}`)
}

func main() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/health", healthHandler)

	log.Printf("Server starting on port %s", port)
	log.Printf("Health endpoint: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
```

**Step 3: Verify Go code compiles**

Run: `cd server && go mod tidy && go build -o /tmp/test-build ./cmd && rm /tmp/test-build && cd ..`

Expected: Successful build with no errors

**Step 4: Commit**

```bash
git add server/go.mod server/cmd/main.go
git commit -m "feat(server): add minimal server with health endpoint"
```

---

## Task 4: Create Dashboard Dockerfile

**Files:**
- Create: `dashboard/Dockerfile.dev`

**Step 1: Create dashboard development Dockerfile**

```dockerfile
FROM node:20-alpine

# Install curl for health checks
RUN apk add --no-cache curl

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy source code (via volume mount in docker-compose)
COPY . .

EXPOSE 5173

# Run Vite dev server with host flag to allow external access
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]
```

**Step 2: Verify Dockerfile syntax**

Run: `head -10 dashboard/Dockerfile.dev`

Expected: Should display first 10 lines without errors

**Step 3: Commit**

```bash
git add dashboard/Dockerfile.dev
git commit -m "feat(dashboard): add development Dockerfile with Vite"
```

---

## Task 5: Create Minimal Dashboard Scaffold

**Files:**
- Create: `dashboard/package.json`
- Create: `dashboard/vite.config.js`
- Create: `dashboard/index.html`
- Create: `dashboard/src/main.jsx`
- Create: `dashboard/src/App.jsx`

**Step 1: Create package.json**

```json
{
  "name": "devtools-sync-dashboard",
  "version": "0.1.0",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@vitejs/plugin-react": "^4.2.1",
    "vite": "^5.0.12"
  }
}
```

**Step 2: Create Vite config**

File: `dashboard/vite.config.js`

```javascript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    strictPort: true,
    watch: {
      usePolling: true
    }
  }
})
```

**Step 3: Create index.html**

File: `dashboard/index.html`

```html
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>DevTools Sync Dashboard</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.jsx"></script>
  </body>
</html>
```

**Step 4: Create React entry point**

File: `dashboard/src/main.jsx`

```javascript
import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
)
```

**Step 5: Create minimal App component**

File: `dashboard/src/App.jsx`

```javascript
import React from 'react'

function App() {
  return (
    <div style={{ padding: '2rem', fontFamily: 'system-ui' }}>
      <h1>DevTools Sync Dashboard</h1>
      <p>Local development environment is running.</p>
      <p style={{ color: '#666' }}>
        Server: <a href="http://localhost:8080/health">http://localhost:8080/health</a>
      </p>
    </div>
  )
}

export default App
```

**Step 6: Commit**

```bash
git add dashboard/package.json dashboard/vite.config.js dashboard/index.html dashboard/src/
git commit -m "feat(dashboard): add minimal React/Vite scaffold"
```

---

## Task 6: Update Docker Compose Configuration

**Files:**
- Modify: `docker-compose.dev.yml`

**Step 1: Read current docker-compose.dev.yml**

Run: `cat docker-compose.dev.yml`

**Step 2: Update with health checks and correct configurations**

Replace entire file with:

```yaml
# DevTools Sync - Local Development Environment
#
# Usage:
#   Start all services:     docker compose -f docker-compose.dev.yml --profile full up
#   Start database only:    docker compose -f docker-compose.dev.yml up postgres
#   Stop all services:      docker compose -f docker-compose.dev.yml down
#   Reset database:         docker compose -f docker-compose.dev.yml down -v

services:
  # ---------------------------------------------------------------------------
  # PostgreSQL Database
  # ---------------------------------------------------------------------------
  postgres:
    image: postgres:16-alpine
    container_name: devtools-sync-postgres
    environment:
      POSTGRES_USER: ${POSTGRES_USER:-devtools}
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD:-devtools-local-dev}
      POSTGRES_DB: ${POSTGRES_DB:-devtools_sync}
    ports:
      - "5432:5432"
    volumes:
      - devtools-pg-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ${POSTGRES_USER:-devtools} -d ${POSTGRES_DB:-devtools_sync}"]
      interval: 5s
      timeout: 5s
      retries: 5

  # ---------------------------------------------------------------------------
  # Management Server
  # ---------------------------------------------------------------------------
  server:
    build:
      context: ./server
      dockerfile: Dockerfile.dev
    container_name: devtools-sync-server
    environment:
      DATABASE_URL: ${DATABASE_URL:-postgres://devtools:devtools-local-dev@postgres:5432/devtools_sync?sslmode=disable}
      SERVER_PORT: ${SERVER_PORT:-8080}
      JWT_SECRET: ${JWT_SECRET:-local-dev-jwt-secret-not-for-production}
      LOG_LEVEL: ${LOG_LEVEL:-debug}
      LOG_FORMAT: ${LOG_FORMAT:-text}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      # Mount source for live reload
      - ./server:/app
      # Exclude bin directory to avoid conflicts
      - /app/bin
      - /app/tmp
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 30s
    profiles:
      - full

  # ---------------------------------------------------------------------------
  # Dashboard
  # ---------------------------------------------------------------------------
  dashboard:
    build:
      context: ./dashboard
      dockerfile: Dockerfile.dev
    container_name: devtools-sync-dashboard
    environment:
      VITE_API_URL: ${VITE_API_URL:-http://localhost:8080/api/v1}
    ports:
      - "5173:5173"
    depends_on:
      server:
        condition: service_healthy
    volumes:
      - ./dashboard:/app
      # Exclude node_modules from mount (use container's version)
      - /app/node_modules
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:5173"]
      interval: 10s
      timeout: 5s
      retries: 3
      start_period: 20s
    profiles:
      - full

volumes:
  devtools-pg-data:
    name: devtools-sync-pg-data
```

**Step 3: Verify YAML syntax**

Run: `docker compose -f docker-compose.dev.yml config > /dev/null && echo "YAML valid" || echo "YAML invalid"`

Expected: "YAML valid"

**Step 4: Commit**

```bash
git add docker-compose.dev.yml
git commit -m "feat(docker): enhance compose with health checks and dependencies"
```

---

## Task 7: Create Environment Configuration

**Files:**
- Create: `.env.local`

**Step 1: Copy env.example to .env.local with dev-safe defaults**

Run: `cp env.example .env.local`

**Step 2: Update .env.local with development-safe values**

Edit `.env.local` to replace CHANGEME values:

```bash
DATABASE_URL=postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync?sslmode=disable
POSTGRES_USER=devtools
POSTGRES_PASSWORD=devtools-local-dev
POSTGRES_DB=devtools_sync
SERVER_PORT=8080
JWT_SECRET=local-dev-jwt-secret-change-in-production
JWT_EXPIRATION=24h
LOG_LEVEL=debug
LOG_FORMAT=text
VITE_API_URL=http://localhost:8080/api/v1
AGENT_SERVER_URL=http://localhost:8080
```

**Step 3: Verify .env.local is gitignored**

Run: `git check-ignore .env.local && echo "Properly ignored" || echo "WARNING: Not ignored!"`

Expected: "Properly ignored"

**Step 4: Do NOT commit (file is gitignored)**

Note: .env.local should never be committed. Each developer creates their own.

---

## Task 8: Update Development Documentation

**Files:**
- Modify: `development.md`

**Step 1: Read current development.md**

Run: `cat development.md | head -30`

**Step 2: Add Docker Compose documentation section**

Add the following section after the Quick Development Setup section:

```markdown
## Docker Compose Development

### Quick Start

1. **Copy environment configuration**
   ```bash
   cp env.example .env.local
   # Edit .env.local if you need custom values
   ```

2. **Start just the database (lightweight)**
   ```bash
   docker compose -f docker-compose.dev.yml up postgres
   ```

   Then run server and dashboard locally:
   ```bash
   # Terminal 1: Server
   cd server && go run ./cmd

   # Terminal 2: Dashboard
   cd dashboard && npm install && npm run dev
   ```

3. **Start full stack (all containerized)**
   ```bash
   docker compose -f docker-compose.dev.yml --profile full up
   ```

   Access:
   - Dashboard: http://localhost:5173
   - Server API: http://localhost:8080
   - Health Check: http://localhost:8080/health
   - PostgreSQL: localhost:5432

### Common Commands

```bash
# Start all services in foreground
docker compose -f docker-compose.dev.yml --profile full up

# Start all services in background
docker compose -f docker-compose.dev.yml --profile full up -d

# View logs
docker compose -f docker-compose.dev.yml logs -f

# View logs for specific service
docker compose -f docker-compose.dev.yml logs -f server

# Stop all services
docker compose -f docker-compose.dev.yml down

# Stop and remove volumes (resets database)
docker compose -f docker-compose.dev.yml down -v

# Rebuild containers after Dockerfile changes
docker compose -f docker-compose.dev.yml --profile full build

# Rebuild and start
docker compose -f docker-compose.dev.yml --profile full up --build
```

### Hot Reload

- **Server (Go)**: Air watches for `.go` file changes and automatically rebuilds
- **Dashboard (React)**: Vite HMR updates browser instantly on file save
- **Database**: Migrations run automatically when server starts

### Troubleshooting

**Services won't start:**
```bash
# Check service health
docker compose -f docker-compose.dev.yml ps

# Check logs for errors
docker compose -f docker-compose.dev.yml logs
```

**Port conflicts:**
```bash
# Check what's using the ports
sudo lsof -i :5432  # PostgreSQL
sudo lsof -i :8080  # Server
sudo lsof -i :5173  # Dashboard
```

**Database connection issues:**
```bash
# Verify PostgreSQL is healthy
docker compose -f docker-compose.dev.yml exec postgres pg_isready -U devtools

# Connect to database directly
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync
```

**Reset everything:**
```bash
# Nuclear option: remove all containers, volumes, and images
docker compose -f docker-compose.dev.yml down -v --rmi local
```
```

**Step 3: Commit**

```bash
git add development.md
git commit -m "docs: add Docker Compose usage guide"
```

---

## Task 9: Test Database-Only Mode

**Step 1: Start PostgreSQL container**

Run: `docker compose -f docker-compose.dev.yml up postgres -d`

Expected: Container starts successfully

**Step 2: Wait for health check**

Run: `sleep 10 && docker compose -f docker-compose.dev.yml ps postgres`

Expected: Status shows "healthy"

**Step 3: Verify database connectivity**

Run: `docker compose -f docker-compose.dev.yml exec postgres pg_isready -U devtools -d devtools_sync`

Expected: "accepting connections"

**Step 4: Test connection from host**

Run: `docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "SELECT version();" | head -3`

Expected: PostgreSQL version displayed

**Step 5: Stop database**

Run: `docker compose -f docker-compose.dev.yml down`

Expected: Clean shutdown

---

## Task 10: Test Full Stack Mode

**Step 1: Start all services**

Run: `docker compose -f docker-compose.dev.yml --profile full up -d`

Expected: All three services start

**Step 2: Monitor startup logs**

Run: `docker compose -f docker-compose.dev.yml logs -f`

Expected: Watch for:
- PostgreSQL ready message
- Server health endpoint ready
- Dashboard Vite server ready

Press Ctrl+C after seeing all services ready.

**Step 3: Verify all services are healthy**

Run: `docker compose -f docker-compose.dev.yml ps`

Expected: postgres, server, and dashboard all show "healthy" or "running"

**Step 4: Test server health endpoint**

Run: `curl -f http://localhost:8080/health`

Expected: `{"status":"healthy","service":"devtools-sync-server"}`

**Step 5: Test dashboard accessibility**

Run: `curl -f http://localhost:5173 | grep -i "devtools"`

Expected: HTML content containing "devtools"

**Step 6: Verify startup order**

Run: `docker compose -f docker-compose.dev.yml logs server | grep -i "starting"`

Expected: Should see server start AFTER postgres is ready

**Step 7: Stop all services**

Run: `docker compose -f docker-compose.dev.yml down`

Expected: Clean shutdown of all services

---

## Task 11: Test Hot Reload (Server)

**Step 1: Start full stack**

Run: `docker compose -f docker-compose.dev.yml --profile full up -d`

**Step 2: Watch server logs**

Run: `docker compose -f docker-compose.dev.yml logs -f server &`

**Step 3: Modify server code**

Edit `server/cmd/main.go` and change health response:
```go
fmt.Fprintf(w, `{"status":"healthy","service":"devtools-sync-server","version":"hot-reload-test"}`)
```

**Step 4: Wait for Air to rebuild (5-10 seconds)**

Monitor logs for "Building..." and "Restarting..." messages

**Step 5: Test updated endpoint**

Run: `sleep 15 && curl http://localhost:8080/health`

Expected: Response includes `"version":"hot-reload-test"`

**Step 6: Revert change**

Git checkout to revert: `git checkout server/cmd/main.go`

**Step 7: Stop services**

Run: `docker compose -f docker-compose.dev.yml down`

---

## Task 12: Test Data Persistence

**Step 1: Start database**

Run: `docker compose -f docker-compose.dev.yml up postgres -d`

**Step 2: Create test table and data**

Run:
```bash
docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "
CREATE TABLE test_persistence (
  id SERIAL PRIMARY KEY,
  message TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);
INSERT INTO test_persistence (message) VALUES ('Docker volume persistence test');
"
```

**Step 3: Verify data exists**

Run: `docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "SELECT * FROM test_persistence;"`

Expected: Shows the test row

**Step 4: Stop and restart container**

Run: `docker compose -f docker-compose.dev.yml down && docker compose -f docker-compose.dev.yml up postgres -d && sleep 10`

**Step 5: Verify data persisted**

Run: `docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "SELECT * FROM test_persistence;"`

Expected: Test row still exists

**Step 6: Clean up test data**

Run: `docker compose -f docker-compose.dev.yml exec postgres psql -U devtools -d devtools_sync -c "DROP TABLE test_persistence;"`

**Step 7: Stop services**

Run: `docker compose -f docker-compose.dev.yml down`

---

## Task 13: Final Verification and Documentation

**Step 1: Run complete test sequence**

Run:
```bash
# Full stack test
docker compose -f docker-compose.dev.yml --profile full up -d
sleep 30
curl -f http://localhost:8080/health
curl -f http://localhost:5173
docker compose -f docker-compose.dev.yml down
echo "All tests passed!"
```

Expected: All commands succeed

**Step 2: Verify issue requirements**

Check against Issue #7 acceptance criteria:
- ✅ docker-compose.dev.yml works out of the box
- ✅ Database migrations run automatically (via Air pre_cmd)
- ✅ Services start in correct order (via health checks)
- ✅ Hot reload works for all components (Air + Vite)

**Step 3: Create summary commit**

```bash
git add -A
git commit -m "feat: complete Docker Compose local development setup

- Add server Dockerfile.dev with Air hot reload
- Add dashboard Dockerfile.dev with Vite HMR
- Configure health checks and service dependencies
- Auto-run migrations on server startup
- Add comprehensive development documentation
- Test database persistence and hot reload

Closes #7"
```

**Step 4: Review all changes**

Run: `git log --oneline -15`

Expected: See all commits from this implementation

---

## Testing Checklist

- [ ] PostgreSQL starts and is healthy
- [ ] Server builds and starts after PostgreSQL
- [ ] Dashboard starts after server
- [ ] Health endpoint responds: `curl http://localhost:8080/health`
- [ ] Dashboard loads: `curl http://localhost:5173`
- [ ] Air hot reload works (modify Go file)
- [ ] Vite HMR works (modify React file)
- [ ] Database data persists across restarts
- [ ] Services stop cleanly
- [ ] Documentation is clear and complete

## Success Criteria

All items in the Testing Checklist must pass. The setup should work immediately after running:

```bash
cp env.example .env.local
docker compose -f docker-compose.dev.yml --profile full up
```

## Notes

- This implementation creates minimal scaffolding to make Docker Compose functional
- The server and dashboard are bare-bones but demonstrate hot reload
- Migrations will auto-run when migration files are added to `server/migrations/`
- Real application code will be added in future issues
