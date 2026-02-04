# Agent-to-Server API Client Design (Issue #13)

## Overview

Enhance the existing API client with authentication, retry logic, and secure token storage.

## Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Authentication | JWT Bearer Token | Server already supports it, industry standard |
| Token Storage | System keychain | Secure, encrypted by OS |
| Retry Logic | Fixed defaults | Simple, works for most cases |
| Auth Failure | Auto re-login | Seamless UX if credentials cached |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    CLI Commands                      │
│         (login, sync push, sync pull, etc.)         │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│               AuthenticatedClient                    │
│  - Manages JWT tokens                                │
│  - Auto-refreshes on 401                            │
│  - Delegates to Client for HTTP                     │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                  Client (existing)                   │
│  - HTTP operations                                   │
│  - Retry with exponential backoff                   │
│  - Response size limits                             │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│               Keychain Storage                       │
│  - Store/retrieve access tokens                     │
│  - Store/retrieve credentials (for auto re-login)  │
└─────────────────────────────────────────────────────┘
```

### New Files

- `agent/internal/api/auth.go` - AuthenticatedClient and auth logic
- `agent/internal/api/auth_test.go` - Auth tests
- `agent/internal/keychain/keychain.go` - Cross-platform secure storage
- `agent/internal/keychain/keychain_test.go` - Keychain tests
- `agent/cmd/login.go` - Login CLI command
- `agent/cmd/logout.go` - Logout CLI command

## Authentication Flow

### Login

```
User runs: devtools-sync login

1. Prompt for email and password (or read from flags)
2. POST /auth/login with credentials
3. Server returns: { access_token, token_type, expires_in }
4. Store access_token in system keychain (key: "devtools-sync-token")
5. Store credentials in keychain (key: "devtools-sync-credentials") for auto re-login
6. Print success message with expiration time
```

### Authenticated Request Flow

```
1. CLI calls AuthenticatedClient.SyncPush()
2. AuthenticatedClient retrieves token from keychain
3. If no token → return error "Please run 'devtools-sync login' first"
4. Add "Authorization: Bearer <token>" header
5. Make request via Client (with retry logic)
6. If 401 response:
   a. Retrieve stored credentials from keychain
   b. If credentials exist → auto re-login, retry original request
   c. If no credentials → return "Session expired, please login again"
7. Return result to CLI
```

### Logout

```
User runs: devtools-sync logout

1. Delete token from keychain
2. Delete stored credentials from keychain
3. Print confirmation
```

## Retry Logic

### Configuration (Fixed Defaults)

```go
const (
    MaxRetries     = 3           // Total attempts = 4 (1 initial + 3 retries)
    InitialDelay   = 1 * time.Second
    MaxDelay       = 30 * time.Second
    BackoffFactor  = 2.0         // Delay doubles each retry
    JitterFactor   = 0.1         // ±10% randomization
)
```

### Retry Timeline Example

```
Attempt 1: Immediate
  ↓ fail (network error)
  Wait: ~1.0s (±0.1s jitter)
Attempt 2:
  ↓ fail (timeout)
  Wait: ~2.0s (±0.2s jitter)
Attempt 3:
  ↓ fail (503 Service Unavailable)
  Wait: ~4.0s (±0.4s jitter)
Attempt 4:
  ↓ fail → return error to caller
```

### Retryable Conditions

- Network errors (connection refused, DNS failure, timeout)
- HTTP 429 (Too Many Requests) - respect Retry-After header if present
- HTTP 502, 503, 504 (server/gateway errors)

### Non-Retryable (Fail Immediately)

- HTTP 400 (Bad Request)
- HTTP 401 (Unauthorized) - handled by auth layer
- HTTP 403 (Forbidden)
- HTTP 404 (Not Found)
- HTTP 422 (Validation error)

## Keychain Storage

### Interface

```go
type Keychain interface {
    Set(key, value string) error
    Get(key string) (string, error)
    Delete(key string) error
}

const (
    KeyAccessToken  = "devtools-sync-token"
    KeyCredentials  = "devtools-sync-credentials"
)
```

### Implementation

Using `github.com/zalando/go-keyring` for cross-platform support:

| Platform | Backend |
|----------|---------|
| Linux | libsecret (GNOME Keyring, KWallet) |
| macOS | macOS Keychain |
| Windows | Windows Credential Manager |

### Fallback

If keychain unavailable (headless server, CI), return error suggesting `DEVTOOLS_SYNC_TOKEN` environment variable.

## New Endpoints

### Authentication

```go
func (c *AuthenticatedClient) Login(email, password string) error
func (c *AuthenticatedClient) Logout() error
```

### Profile Update (Missing)

```go
func (c *Client) UpdateProfile(name string, profile *Profile) error  // PUT /api/v1/profiles/:name
```

### Sync

```go
func (c *Client) Sync(req *SyncRequest) (*SyncResponse, error)  // POST /api/v1/sync

type SyncRequest struct {
    ProfileName string      `json:"profile_name"`
    Extensions  []Extension `json:"extensions"`
    LastSync    time.Time   `json:"last_sync,omitempty"`
}

type SyncResponse struct {
    Status    string      `json:"status"`
    Merged    []Extension `json:"merged,omitempty"`
    Conflicts []Conflict  `json:"conflicts,omitempty"`
}
```

## Testing Strategy

| Test Type | Approach |
|-----------|----------|
| Unit tests | Mock HTTP server (httptest), mock keychain interface |
| Retry tests | Mock server returning errors then success |
| Auth tests | Mock 401 responses, verify re-login flow |
| Integration | Optional test against real server (skipped in CI) |

**Coverage target:** 80%+

## Security Considerations

1. **Credentials in keychain** - Encrypted by OS, not readable by other processes
2. **Token expiration** - Server sets expiration, client respects it
3. **Auto re-login** - Only if user explicitly logged in (credentials stored)
4. **No plaintext storage** - Never write tokens/passwords to config files
5. **CI/Headless fallback** - Environment variable for automation scenarios

## Implementation Order

1. Keychain package (foundation for storage)
2. Retry logic in existing Client
3. AuthenticatedClient wrapper
4. Login/Logout CLI commands
5. UpdateProfile endpoint
6. Sync endpoint
7. Integration tests
