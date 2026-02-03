# CLI Agent Foundation Design

**Date:** 2026-02-03
**Status:** Approved
**Issue:** #10
**Scope:** MVP - Core sync functionality

## Overview

Implement the CLI agent foundation using Cobra framework to enable basic profile management and synchronization workflows. This design focuses on delivering core value: save extensions, sync to server, load on another machine.

## Command Structure

```
devtools-sync
├── version          (print version info)
├── init             (create config file and directories)
├── config
│   ├── show         (display current config)
│   └── set          (update config values)
├── profile
│   ├── save <name>  (capture current extensions)
│   ├── load <name>  (install extensions from profile)
│   └── list         (show local profiles)
└── sync
    ├── push         (upload profiles to server)
    └── pull         (download profiles from server)
```

## Architecture

### Directory Structure

```
agent/
├── cmd/
│   ├── main.go           (entry point, root command)
│   ├── version.go        (version command)
│   ├── init.go           (init command)
│   ├── config.go         (config command + subcommands)
│   ├── profile.go        (profile command + subcommands)
│   └── sync.go           (sync command + subcommands)
├── internal/
│   ├── config/           (config loading, existing)
│   ├── vscode/           (VS Code interaction, existing)
│   ├── api/              (server client, existing)
│   └── profile/          (NEW: profile storage/management)
└── go.mod
```

### Dependencies

- `github.com/spf13/cobra` - CLI framework
- `gopkg.in/yaml.v3` - YAML config parsing

## Configuration System

### Config File Format

**Location:** `~/.devtools-sync/config.yaml`

```yaml
server:
  url: http://localhost:8080

profiles:
  directory: ~/.devtools-sync/profiles

logging:
  level: info
```

### Loading Strategy

1. Check if `~/.devtools-sync/config.yaml` exists
2. Parse YAML into `Config` struct
3. Override with environment variables:
   - `DEVTOOLS_SYNC_SERVER_URL` → `server.url`
   - `DEVTOOLS_SYNC_LOG_LEVEL` → `logging.level`
4. Apply defaults for missing values
5. Validate the final config

### Config Struct

```go
type Config struct {
    Server struct {
        URL string `yaml:"url"`
    } `yaml:"server"`
    Profiles struct {
        Directory string `yaml:"directory"`
    } `yaml:"profiles"`
    Logging struct {
        Level string `yaml:"level"`
    } `yaml:"logging"`
}
```

### Commands

#### `init`

```bash
devtools-sync init
# Creates:
# - ~/.devtools-sync/config.yaml (default config)
# - ~/.devtools-sync/profiles/ (empty directory)
# Output: "Configuration initialized at ~/.devtools-sync"
```

#### `config show`

```bash
devtools-sync config show
# Displays current effective config (after env var overrides)
```

#### `config set`

```bash
devtools-sync config set server.url https://api.example.com
# Updates config.yaml, validates, and confirms
```

## Profile Management

### Profile File Format

**Location:** `~/.devtools-sync/profiles/<name>.json`

```json
{
  "name": "work-setup",
  "created_at": "2026-02-03T12:34:56Z",
  "updated_at": "2026-02-03T12:34:56Z",
  "extensions": [
    {
      "id": "ms-python.python",
      "version": "2024.0.1",
      "enabled": true
    },
    {
      "id": "golang.go",
      "version": "0.41.0",
      "enabled": true
    }
  ]
}
```

### Profile Package

**File:** `internal/profile/profile.go`

```go
type Profile struct {
    Name       string      `json:"name"`
    CreatedAt  time.Time   `json:"created_at"`
    UpdatedAt  time.Time   `json:"updated_at"`
    Extensions []Extension `json:"extensions"`
}

type Extension struct {
    ID      string `json:"id"`
    Version string `json:"version"`
    Enabled bool   `json:"enabled"`
}

// Save captures current VS Code extensions to a profile
func Save(name string, profilesDir string) error

// Load installs extensions from a profile
func Load(name string, profilesDir string) error

// List returns all local profiles
func List(profilesDir string) ([]Profile, error)
```

### Commands

#### `profile save`

```bash
devtools-sync profile save work-setup
# 1. Call vscode.ListExtensions() to get current extensions
# 2. Create Profile struct with current timestamp
# 3. Write JSON to ~/.devtools-sync/profiles/work-setup.json
# Output: "Saved 47 extensions to profile 'work-setup'"
```

#### `profile load`

```bash
devtools-sync profile load work-setup
# 1. Read ~/.devtools-sync/profiles/work-setup.json
# 2. For each extension, call vscode.InstallExtension(id)
# Output: "Installing 47 extensions from profile 'work-setup'..."
```

#### `profile list`

```bash
devtools-sync profile list
# 1. Read all .json files from profiles directory
# 2. Parse and display: name, extension count, last updated
# Output: Table format showing all profiles
```

### VS Code Integration

The `vscode` package already has `ListExtensions()` and `InstallExtension()` stubs. Implementation will:
- Execute `code --list-extensions --show-versions` to list extensions
- Execute `code --install-extension <id>` to install extensions
- Parse output and handle errors

## Sync Implementation

### API Endpoints (Server-side context)

```
POST   /api/v1/profiles         (upload profile)
GET    /api/v1/profiles         (list all profiles)
GET    /api/v1/profiles/:name   (download specific profile)
DELETE /api/v1/profiles/:name   (delete profile)
```

### API Client Methods

**File:** `internal/api/client.go` (extend existing)

```go
// UploadProfile sends a profile to the server
func (c *Client) UploadProfile(profile *profile.Profile) error

// ListProfiles retrieves all profile names from server
func (c *Client) ListProfiles() ([]string, error)

// DownloadProfile retrieves a specific profile
func (c *Client) DownloadProfile(name string) (*profile.Profile, error)
```

### Commands

#### `sync push`

```bash
devtools-sync sync push
# 1. Read all profiles from ~/.devtools-sync/profiles/
# 2. For each profile, POST to /api/v1/profiles
# 3. Report success/failure for each
# Output: "Pushed 3 profiles: work-setup, personal, minimal"
```

#### `sync pull`

```bash
devtools-sync sync pull
# 1. GET /api/v1/profiles (list server profiles)
# 2. For each server profile, GET /api/v1/profiles/:name
# 3. Write to local profiles directory
# 4. Skip if local version is newer (compare updated_at)
# Output: "Pulled 2 profiles, skipped 1 (local is newer)"
```

### Authentication

For MVP, authentication is deferred. The server's health endpoint currently works without auth. Future implementation will add `Authorization: Bearer <token>` headers when server auth is available.

## Error Handling

### Principles

**User-Facing Errors** must be:
- Clear and specific
- Actionable (what to do next)
- Contextual (include relevant details)

### Examples

```go
// Bad
return fmt.Errorf("failed: %w", err)

// Good
return fmt.Errorf("failed to save profile 'work-setup': VS Code not found. Install VS Code or check PATH")
```

### Common Error Scenarios

| Scenario | Error Message | Guidance |
|----------|--------------|----------|
| VS Code not installed | `VS Code not found in PATH` | Install VS Code or add to PATH |
| Config missing | `Config file not found` | Run `devtools-sync init` |
| Server unreachable | `Failed to connect to server at <URL>` | Check server URL in config |
| Profile not found | `Profile 'foo' not found` | Show available profiles |
| Network timeout | `Request timeout after 10s` | Check network connection |

## Testing Strategy

### Unit Tests

**Config Package** (`internal/config/config_test.go`)
- `TestLoadConfigFile` - YAML parsing
- `TestEnvVarOverride` - Environment variable precedence
- `TestConfigDefaults` - Default value application
- `TestConfigValidation` - Invalid config detection

**Profile Package** (`internal/profile/profile_test.go`)
- `TestSaveProfile` - Save with mock vscode data
- `TestLoadProfile` - Verify install calls
- `TestListProfiles` - Empty dir, multiple profiles
- `TestProfileValidation` - Invalid profile data

**API Client** (`internal/api/client_test.go`)
- `TestUploadProfile` - Mock HTTP POST
- `TestListProfiles` - Mock HTTP GET
- `TestDownloadProfile` - Mock HTTP GET
- `TestNetworkErrors` - Connection failures

**Command Tests** (`cmd/*_test.go`)
- `TestVersionCommand` - Output format
- `TestInitCommand` - Creates files/directories
- `TestConfigShowCommand` - Displays values
- `TestConfigSetCommand` - Updates config
- `TestProfileSaveCommand` - Happy path
- `TestProfileLoadCommand` - Happy path
- `TestProfileListCommand` - Output formatting
- `TestSyncPushCommand` - Upload flow
- `TestSyncPullCommand` - Download flow

### Integration Tests

**Profile Workflow** (`test/integration/profile_workflow_test.go`)
- Save → List → Load complete workflow
- Save → Push → Pull → Load (requires test server)

**VS Code Integration**
- Tests require VS Code installed (skip in CI)
- Mock exec commands for unit tests
- Real integration tests run manually

### Coverage Goals

- Core logic (profile, config): 80%+
- Commands (happy path): 100%
- Error paths: Key scenarios covered

## Implementation Plan

### Phase 1: Foundation
1. Add Cobra and YAML dependencies
2. Implement root command and version
3. Update config package for YAML + env vars
4. Implement `init` command

### Phase 2: Configuration
1. Implement `config show` command
2. Implement `config set` command
3. Add tests for config commands

### Phase 3: Profile Management
1. Create profile package with Save/Load/List
2. Implement VS Code CLI integration
3. Implement `profile save/load/list` commands
4. Add tests for profile package and commands

### Phase 4: Sync
1. Extend API client with profile methods
2. Implement `sync push` command
3. Implement `sync pull` command
4. Add tests for sync commands

### Phase 5: Polish
1. Add shell completion support
2. Improve error messages
3. Add integration tests
4. Update documentation

## Success Criteria

- [ ] All commands are accessible via Cobra
- [ ] Help text is clear and useful
- [ ] Config file is created and loaded correctly
- [ ] Profiles can be saved, loaded, and listed
- [ ] Sync push/pull work with server
- [ ] Unit tests pass with 80%+ coverage
- [ ] Error messages are actionable
- [ ] Documentation is complete

## Future Enhancements (Out of Scope)

- Authentication/authorization
- Auto-sync mode
- Conflict resolution (merge strategies)
- Profile templates
- Extension search/discovery
- Team profile sharing
- Bulk extension operations

## References

- Issue #10: Implement CLI agent foundation
- Cobra documentation: https://github.com/spf13/cobra
- VS Code CLI: https://code.visualstudio.com/docs/editor/command-line
