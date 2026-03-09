# Security Headers Design — Issue #42

## Overview

Add security headers to the React dashboard to protect against XSS, clickjacking, MIME sniffing, and other common web vulnerabilities. Headers are applied at two layers: Vite dev server config for local development, and Go `SecurityHeaders` middleware for production.

## Headers

| Header | Value |
|--------|-------|
| Content-Security-Policy | `default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'` |
| X-Frame-Options | `DENY` |
| X-Content-Type-Options | `nosniff` |
| Strict-Transport-Security | `max-age=31536000; includeSubDomains` |
| Referrer-Policy | `strict-origin-when-cross-origin` |
| Permissions-Policy | `camera=(), microphone=(), geolocation=()` |

- `style-src 'unsafe-inline'` required for React inline styles.
- HSTS only set in Go middleware (not dev), since dev may not use TLS.
- No nonce-based CSP needed — Vite bundles all scripts, no inline scripts.
- `connect-src 'self'` sufficient since dashboard and API share the same origin in production.

## Architecture

### Dev (Vite)

Static headers in `vite.config.js` `server.headers` block. All headers except HSTS.

### Production (Go middleware)

New `SecurityHeaders` middleware in `server/internal/middleware/security_headers.go`, following the same pattern as `cors.go`. Applied in the middleware chain in `server/cmd/main.go`:

```
CORS → MaxBodySize → SecurityHeaders → mux
```

## Files

| File | Action |
|------|--------|
| `dashboard/vite.config.js` | Add `server.headers` with all headers except HSTS |
| `server/internal/middleware/security_headers.go` | New middleware |
| `server/internal/middleware/security_headers_test.go` | Test all headers are set |
| `server/cmd/main.go` | Add `SecurityHeaders` to middleware chain |
