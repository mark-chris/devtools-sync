# Rate Limiter Integration with Auth Endpoints Design

**Issue:** #58 — Integrate rate limiter with authentication endpoints
**Date:** 2026-02-16
**Status:** Approved

## Problem

The rate limiter component exists and is hardened (Issue #59), but is not integrated with authentication handlers. Login, refresh, logout, and accept-invite endpoints are vulnerable to brute force attacks.

## Scope

Wire the existing `RateLimiter` into auth endpoints via middleware. IP-based rate limiting only (account-based deferred).

## Design

### Approach: Middleware per endpoint

A `RateLimit` middleware function wraps individual handlers, matching the existing middleware pattern (`MaxBodySize`, `RequireAuth`).

### New files

- `server/internal/middleware/rate_limit.go` — middleware + `getClientIP` helper
- `server/internal/middleware/rate_limit_test.go` — tests

### Modified files

- `server/internal/api/auth_handlers.go` — `NewLoginHandler` gains `*auth.RateLimiter` param, calls `ResetLimit` on successful login

### Middleware signature

```go
func RateLimit(rl *auth.RateLimiter, maxAttempts int, window time.Duration) func(http.Handler) http.Handler
```

On limit exceeded: responds 429 with `Retry-After` header, logs the event.

### getClientIP

Checks `X-Forwarded-For` (first IP), then `X-Real-IP`, then `r.RemoteAddr` (stripping port).

### Rate limits

| Endpoint | Max Attempts | Window |
|---|---|---|
| `/auth/login` | 5 | 15 minutes |
| `/auth/refresh` | 10 | 1 minute |
| `/auth/logout` | 20 | 1 minute |
| `/accept-invite` | 5 | 1 hour |

### Login reset

Successful login calls `rateLimiter.ResetLimit(clientIP)` to clear the counter.

### Tests

- Middleware allows requests under limit
- Middleware blocks with 429 + Retry-After over limit
- getClientIP extracts from X-Forwarded-For, X-Real-IP, RemoteAddr
- Login handler resets rate limit on success
