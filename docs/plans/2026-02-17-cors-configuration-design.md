# CORS Configuration Design

**Issue:** #62 — Implement CORS configuration for dashboard API access
**Date:** 2026-02-17
**Severity:** Low (Threat Model Finding L-1)

## Context

The server has no CORS configuration. The React dashboard needs to make API calls to the server from a different origin during development and potentially in production. Deployment topology is undecided, so origins must be configurable via environment variable.

## Approach

Hand-rolled CORS middleware matching existing `func(http.Handler) http.Handler` pattern. No new dependencies.

## Design

### New file: `server/internal/middleware/cors.go`

```go
func CORS(allowedOrigins []string) func(http.Handler) http.Handler
```

- Builds `map[string]bool` from origins list for O(1) lookup
- Empty list → no-op passthrough (secure by default)
- On every request:
  1. Sets `Vary: Origin` (required for correct HTTP caching)
  2. Checks `Origin` request header against allowlist
  3. If matched: sets `Access-Control-Allow-Origin: <matched origin>`, `Access-Control-Allow-Credentials: true`
  4. If `OPTIONS` preflight and matched: additionally sets `Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`, `Allow-Headers: Authorization, Content-Type`, `Max-Age: 86400`, returns `204 No Content`
  5. Otherwise: calls next handler

### Integration in `server/cmd/main.go`

CORS wraps outermost so preflight OPTIONS requests are answered before body size checks or auth:

```go
corsOrigins := parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
handler := middleware.CORS(corsOrigins)(middleware.MaxBodySize(maxBodySize)(mux))
```

### Configuration

`CORS_ALLOWED_ORIGINS` environment variable, comma-separated. No default origins — if unset, no CORS headers are sent.

Example: `CORS_ALLOWED_ORIGINS=http://localhost:5173,https://dashboard.example.com`

### Security properties

- No wildcard `*` origin (incompatible with credentials)
- `Allow-Credentials: true` only sent to matched origins
- Unmatched origins receive no CORS headers (browser blocks the request)
- `Vary: Origin` always set for correct caching

### Tests (`server/internal/middleware/cors_test.go`)

- Allowed origin gets CORS headers
- Disallowed origin gets no CORS headers
- Preflight OPTIONS returns 204 with correct headers
- No Origin header → no CORS headers
- Empty config → passthrough
- Vary: Origin always set
- Credentials header set correctly
