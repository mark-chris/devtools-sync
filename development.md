# Development Guide

This guide covers everything you need to set up a local development environment for DevTools Sync.

## Prerequisites

Ensure you have the following installed:

| Tool | Version | Installation |
|------|---------|--------------|
| Go | 1.22+ | [go.dev/dl](https://go.dev/dl/) |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) or use `nvm` |
| Docker | Latest | [docker.com](https://www.docker.com/get-started/) |
| Git | Latest | [git-scm.com](https://git-scm.com/) |

Optional but recommended:

| Tool | Purpose |
|------|---------|
| Make | Build automation |
| golangci-lint | Go linting ([install](https://golangci-lint.run/usage/install/)) |
| VS Code | Recommended editor |

## Quick Start

The fastest way to get everything running:

```bash
# Clone the repository
git clone https://github.com/youruser/devtools-sync.git
cd devtools-sync

# Start PostgreSQL
docker compose -f docker-compose.dev.yml up -d postgres

# In separate terminals, start server and dashboard:

# Terminal 1: Server
cd server
go run ./cmd serve

# Terminal 2: Dashboard
cd dashboard
npm install
npm run dev
```

The dashboard will be available at `http://localhost:5173` and the API at `http://localhost:8080`.

## Environment Setup

### Option A: Using .env File (Recommended)

```bash
# Copy the example environment file
cp .env.example .env

# Edit with your preferred values (or keep defaults for local dev)
# The default CHANGEME values should be replaced
```

Key variables to set:

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret for signing tokens (generate with `openssl rand -base64 32`) |
| `POSTGRES_PASSWORD` | Database password (must match DATABASE_URL) |

### Option B: Using Docker Compose Only

If you just want to run the database without configuring anything:

```bash
# Start PostgreSQL with pre-configured credentials
docker compose -f docker-compose.dev.yml up -d postgres

# The connection string for this setup is:
# postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync
```

## Component Setup

### Agent

The agent is a Go binary that runs on developer workstations.

```bash
cd agent

# Download dependencies
go mod download

# Run tests
go test -v ./...

# Build the binary
go build -o bin/devtools-sync-agent ./cmd

# Test the CLI
./bin/devtools-sync-agent --help

# Run with a local config file
./bin/devtools-sync-agent sync --config ~/.devtools-sync/config.yaml
```

Cross-compile for other platforms:

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o bin/devtools-sync-agent.exe ./cmd

# macOS (Intel)
GOOS=darwin GOARCH=amd64 go build -o bin/devtools-sync-agent-darwin ./cmd

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/devtools-sync-agent-darwin-arm64 ./cmd

# Linux
GOOS=linux GOARCH=amd64 go build -o bin/devtools-sync-agent-linux ./cmd
```

### Server

The management server provides the REST API.

```bash
cd server

# Download dependencies
go mod download

# Ensure PostgreSQL is running
docker compose -f ../docker-compose.dev.yml up -d postgres

# Run database migrations
go run ./cmd migrate up

# Start the server
go run ./cmd serve

# Or with environment variables
DATABASE_URL="postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync?sslmode=disable" \
JWT_SECRET="dev-secret" \
go run ./cmd serve
```

The server will start on `http://localhost:8080`. Test it:

```bash
curl http://localhost:8080/health
```

### Dashboard

The dashboard is a React SPA built with Vite.

```bash
cd dashboard

# Install dependencies
npm install

# Start development server (with hot reload)
npm run dev

# Run linter
npm run lint

# Run tests
npm test

# Build for production
npm run build
```

The development server runs on `http://localhost:5173` with hot module replacement.

## Database Management

### Starting PostgreSQL

```bash
# Start the database
docker compose -f docker-compose.dev.yml up -d postgres

# View logs
docker logs devtools-sync-postgres

# Stop the database (preserves data)
docker compose -f docker-compose.dev.yml stop postgres

# Stop and remove all data
docker compose -f docker-compose.dev.yml down -v
```

### Running Migrations

```bash
cd server

# Apply all pending migrations
go run ./cmd migrate up

# Rollback last migration
go run ./cmd migrate down

# Check migration status
go run ./cmd migrate status
```

### Connecting Directly

```bash
# Using Docker
docker exec -it devtools-sync-postgres psql -U devtools -d devtools_sync

# Using psql locally
psql "postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync"
```

### Using SQLite (Alternative)

For quick testing without PostgreSQL:

```bash
cd server
DATABASE_URL="sqlite://./devtools_sync.db" go run ./cmd serve
```

Note: SQLite is suitable for development and small deployments only.

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

## Running Tests

### All Tests

```bash
# Using Make (if Makefile exists)
make test

# Or manually
cd agent && go test -v ./...
cd ../server && go test -v ./...
cd ../dashboard && npm test
```

### With Coverage

```bash
# Agent
cd agent
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # Opens browser

# Server
cd server
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Dashboard
cd dashboard
npm test -- --coverage
```

### Integration Tests

```bash
cd server

# Ensure test database exists
docker exec devtools-sync-postgres psql -U devtools -c "CREATE DATABASE devtools_sync_test;" || true

# Run integration tests
DATABASE_URL="postgres://devtools:devtools-local-dev@localhost:5432/devtools_sync_test?sslmode=disable" \
go test -v -tags=integration ./...
```

## Code Quality

### Linting

```bash
# Go (requires golangci-lint)
cd agent && golangci-lint run
cd ../server && golangci-lint run

# Dashboard
cd dashboard && npm run lint

# Auto-fix where possible
cd dashboard && npm run lint:fix
```

### Formatting

```bash
# Go
cd agent && go fmt ./...
cd ../server && go fmt ./...

# Dashboard (with Prettier)
cd dashboard && npm run format
```

## Using Make

If you prefer Make, common targets are available in the root Makefile:

```bash
make help          # Show all targets
make all           # Build everything
make test          # Run all tests
make lint          # Lint all code
make dev-server    # Run server in dev mode
make dev-dashboard # Run dashboard in dev mode
make clean         # Remove build artifacts
```

## IDE Setup

### VS Code

Recommended extensions are listed in `.vscode/extensions.json`. VS Code should prompt you to install them when you open the project.

Manual install:

```bash
code --install-extension golang.go
code --install-extension dbaeumer.vscode-eslint
code --install-extension bradlc.vscode-tailwindcss
code --install-extension esbenp.prettier-vscode
```

Recommended settings (`.vscode/settings.json`):

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "[go]": {
    "editor.defaultFormatter": "golang.go"
  }
}
```

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080

# Kill it or change SERVER_PORT in your .env
```

### Database Connection Refused

```bash
# Check if PostgreSQL is running
docker ps | grep devtools-sync-postgres

# Check logs
docker logs devtools-sync-postgres

# Restart if needed
docker compose -f docker-compose.dev.yml restart postgres
```

### Go Module Errors

```bash
cd agent  # or server
go mod tidy
go mod download
```

### Node Module Errors

```bash
cd dashboard
rm -rf node_modules package-lock.json
npm install
```

### Permission Denied (Docker on Linux)

```bash
# Add your user to the docker group
sudo usermod -aG docker $USER

# Log out and back in, or run:
newgrp docker
```

## Project Structure

```
devtools-sync/
├── agent/                 # Go agent (runs on workstations)
│   ├── cmd/               # CLI entry point
│   ├── internal/          # Private packages
│   │   ├── api/           # API client
│   │   ├── config/        # Configuration handling
│   │   └── vscode/        # VS Code integration
│   └── go.mod
├── server/                # Go management service
│   ├── cmd/               # Server entry point
│   ├── internal/          # Private packages
│   │   ├── api/           # HTTP handlers
│   │   ├── database/      # Database layer
│   │   └── middleware/    # HTTP middleware
│   ├── migrations/        # SQL migrations
│   └── go.mod
├── dashboard/             # React dashboard
│   ├── src/
│   │   ├── components/    # Reusable UI components
│   │   ├── pages/         # Route pages
│   │   └── services/      # API client
│   └── package.json
├── docs/                  # Documentation
├── examples/              # Example configurations
├── .env.example           # Environment template
├── docker-compose.dev.yml # Local development services
└── Makefile               # Build automation
```

## Next Steps

- Read [CONTRIBUTING.md](../CONTRIBUTING.md) before submitting PRs
- Check [docs/architecture.md](./architecture.md) for design decisions
- See [docs/api-reference.md](./api-reference.md) for API documentation
