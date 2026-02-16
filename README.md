# DevTools Sync

[![Agent CI](https://github.com/mark-chris/devtools-sync/actions/workflows/agent.yml/badge.svg)](https://github.com/mark-chris/devtools-sync/actions/workflows/agent.yml)
[![Server CI](https://github.com/mark-chris/devtools-sync/actions/workflows/server.yml/badge.svg)](https://github.com/mark-chris/devtools-sync/actions/workflows/server.yml)
[![Dashboard CI](https://github.com/mark-chris/devtools-sync/actions/workflows/dashboard.yml/badge.svg)](https://github.com/mark-chris/devtools-sync/actions/workflows/dashboard.yml)
[![Lint](https://github.com/mark-chris/devtools-sync/actions/workflows/lint.yml/badge.svg)](https://github.com/mark-chris/devtools-sync/actions/workflows/lint.yml)
[![codecov](https://codecov.io/gh/mark-chris/devtools-sync/branch/main/graph/badge.svg)](https://codecov.io/gh/mark-chris/devtools-sync)

A comprehensive platform for managing and synchronizing VS Code extensions across multiple machines and teams.

## Project Status

This project is under active development. See our [GitHub Projects board](https://github.com/users/mark-chris/projects/1) for current progress and roadmap.

## Features

- 🔄 **Profile Management**: Create and manage multiple extension profiles
- 🌐 **Cross-Machine Sync**: Keep your extensions synchronized across all your development machines
- 👥 **Team Collaboration**: Share extension profiles with your team
- 📊 **Extension Analytics**: Track extension usage and optimize your setup
- 🔍 **Smart Recommendations**: Get intelligent extension suggestions based on your workspace
- 🎯 **Bulk Operations**: Install, uninstall, and manage extensions in batches
- 📈 **Conflict Detection**: Identify and resolve extension conflicts automatically
- 🔒 **Secure Sync**: End-to-end encrypted synchronization of your extension data

## Architecture

DevTools Sync consists of three main components:

- **Agent** (Go): CLI tool that runs on developer workstations to manage local VS Code extensions
- **Server** (Go): Backend API that handles authentication, profile storage, and synchronization
- **Dashboard** (React): Web interface for managing profiles, teams, and viewing analytics

## Quick Start

### Using the CLI Agent

```bash
# Install the agent
curl -fsSL https://raw.githubusercontent.com/mark-chris/devtools-sync/main/install.sh | sh

# Initialize and login
devtools-sync init
devtools-sync login

# Save your current extensions to a profile
devtools-sync profile save my-setup

# Sync to the server
devtools-sync sync push

# On another machine, pull your profile
devtools-sync sync pull
devtools-sync profile load my-setup
```

### Using the Dashboard

Visit the dashboard at `https://devtools-sync.example.com` to:
- Browse and manage your extension profiles
- Create and manage teams
- View extension usage analytics
- Search and discover new extensions

## Installation

### CLI Agent

**macOS/Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/mark-chris/devtools-sync/main/install.sh | sh
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/mark-chris/devtools-sync/main/install.ps1 | iex
```

**From Source:**
```bash
git clone https://github.com/mark-chris/devtools-sync.git
cd devtools-sync/agent
go build -o devtools-sync-agent ./cmd
```

### Self-Hosting

See the [deployment guide](docs/deployment.md) for instructions on hosting your own DevTools Sync server.

## Usage

### Profile Management

```bash
# Save current extensions to a new profile
devtools-sync profile save work-setup

# List all profiles
devtools-sync profile list

# Compare a profile with currently installed extensions
devtools-sync profile diff work-setup

# Load a profile
devtools-sync profile load work-setup

# Show profile details
devtools-sync profile show work-setup

# Delete a profile
devtools-sync profile delete old-setup
```

#### Profile Features

- **Validation**: Profiles are validated to ensure extension IDs follow the correct format (publisher.name)
- **Conflict Detection**: Automatically detects which extensions are already installed vs. need to be installed
- **Diff Command**: Preview what would change before loading a profile
- **Idempotent Loading**: Loading a profile multiple times won't reinstall already installed extensions
- **Save-Diff-Load Workflow**: Compare profiles before applying them to avoid unexpected changes

#### Example Workflow

```bash
# Save your current setup
devtools-sync profile save my-setup

# On another machine, check what would be installed
devtools-sync profile diff my-setup
# Output shows:
#   To Install (5): new extensions that will be installed
#   Already Installed (3): extensions you already have

# Load the profile (only installs missing extensions)
devtools-sync profile load my-setup
# Skips already installed extensions automatically
```

### Authentication

```bash
# Login (stores token securely in system keychain)
devtools-sync login
# Or with flags:
devtools-sync login --email user@example.com --password mypassword

# Logout (removes stored credentials)
devtools-sync logout
```

### Synchronization

Sync commands require authentication. Run `devtools-sync login` first.

```bash
# Push local profiles to server
devtools-sync sync push

# Pull profiles from server
devtools-sync sync pull

# Auto-sync (watches for changes)
devtools-sync sync auto
```

### Team Collaboration

```bash
# Share a profile with your team
devtools-sync team share my-profile team-name

# Load a team profile
devtools-sync team load team-profile

# List team profiles
devtools-sync team list
```

### Extension Operations

```bash
# Install extensions from a profile
devtools-sync install my-profile

# Search for extensions
devtools-sync search "python"

# Get recommendations
devtools-sync recommend
```

## Configuration

Configuration is stored in `~/.devtools-sync/config.yaml`:

```yaml
server:
  url: https://devtools-sync.example.com
  api_key: your-api-key

sync:
  auto_sync: false
  interval: 300  # seconds

profiles:
  default: work-setup

logging:
  level: info
  file: ~/.devtools-sync/logs/agent.log
```

## Security

### Token Storage

Authentication tokens are stored securely in your system's native keychain:
- **Linux**: libsecret (GNOME Keyring, KWallet)
- **macOS**: macOS Keychain
- **Windows**: Windows Credential Manager

### Retry Logic

The client automatically retries failed requests with exponential backoff:
- **Max retries**: 3 (4 total attempts)
- **Initial delay**: 1 second
- **Backoff factor**: 2x (1s, 2s, 4s)
- **Jitter**: +/-10% randomization
- **Retries on**: network errors, HTTP 429/502/503/504

## Development

See [development.md](development.md) for detailed development setup instructions.

### Quick Development Setup

```bash
# Clone the repository
git clone https://github.com/mark-chris/devtools-sync.git
cd devtools-sync

# Start PostgreSQL
docker compose -f docker-compose.dev.yml up -d postgres

# Terminal 1: Run the server
cd server
go run ./cmd serve

# Terminal 2: Run the dashboard
cd dashboard
npm install
npm run dev

# Terminal 3: Build the agent
cd agent
go build -o bin/devtools-sync-agent ./cmd
```

### Running Tests

```bash
# Run all tests
make test

# Run tests for a specific component
cd agent && go test ./...
cd server && go test ./...
cd dashboard && npm test
```

## Documentation

- [API Reference](docs/api-reference.md)
- [Architecture Overview](docs/architecture.md)
- [Development Guide](development.md)
- [Deployment Guide](docs/deployment.md)
- [Contributing Guidelines](CONTRIBUTING.md)
- [Security Policy](SECURITY.md)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on:

- Code of Conduct
- Development setup
- Submitting pull requests
- Coding standards
- Testing requirements

## License

MIT License - see [LICENSE.md](LICENSE.md) for details.

## Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/mark-chris/devtools-sync/issues)
- 💬 [Discussions](https://github.com/mark-chris/devtools-sync/discussions)
- 📧 Email: support@devtools-sync.example.com

## Acknowledgments

Built with ❤️ for the developer community.

---

**Note**: This tool is not affiliated with or endorsed by Microsoft or Visual Studio Code.
