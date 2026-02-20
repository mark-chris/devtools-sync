# TLS/HTTPS Support for Server — Design

Issue: #38

## Context

The server currently uses bare `http.ListenAndServe` on port 8080 with no TLS, no server timeouts, and no graceful shutdown. The primary deployment model is reverse proxy (nginx/Caddy/Traefik) for TLS termination. Native TLS is a secondary option for simpler setups and local development.

## Approach

Refactor `main.go` to use an `http.Server` struct with optional TLS, proper timeouts, and graceful shutdown. All configuration via environment variables following the existing pattern.

## Configuration

New environment variables:

| Variable | Default | Description |
|---|---|---|
| `TLS_ENABLED` | `false` | Set `true` to enable native TLS |
| `TLS_CERT_FILE` | (required if TLS enabled) | Path to PEM certificate file |
| `TLS_KEY_FILE` | (required if TLS enabled) | Path to PEM private key file |
| `TLS_MIN_VERSION` | `1.2` | Minimum TLS version (`1.2` or `1.3`) |

Validation:
- If `TLS_ENABLED=true`, both `TLS_CERT_FILE` and `TLS_KEY_FILE` must be set and files must exist
- `TLS_MIN_VERSION` only accepts `1.2` or `1.3`
- TLS is never required in development mode

## Server Struct & Timeouts

Replace `http.ListenAndServe` with:

```go
srv := &http.Server{
    Addr:              ":" + port,
    Handler:           handler,
    ReadHeaderTimeout: 10 * time.Second,
    ReadTimeout:       30 * time.Second,
    WriteTimeout:      60 * time.Second,
    IdleTimeout:       120 * time.Second,
}
```

When TLS is enabled:

```go
srv.TLSConfig = &tls.Config{
    MinVersion: tlsMinVersion, // tls.VersionTLS12 or tls.VersionTLS13
}
```

Cipher suites are left to Go 1.24 defaults (secure, well-ordered, no maintenance burden).

## Graceful Shutdown

Start the server in a goroutine. The main goroutine waits for SIGINT or SIGTERM, then calls `srv.Shutdown` with a 30-second context deadline to drain in-flight requests.

## Dev Cert Script

`scripts/generate-dev-certs.sh`:
- Generates a self-signed CA + server certificate
- Outputs to `certs/` directory (gitignored)
- Server cert includes `localhost` and `127.0.0.1` SANs
- 2048-bit RSA, 365-day validity
- Prints trust instructions

## Documentation

- `env.example` updated with TLS section (commented out by default)
- Comments clarify that reverse proxy is primary, native TLS is optional
- `certs/` and `*.pem`/`*.key` added to `.gitignore`

## Testing

Unit tests for:
- `parseTLSConfig()` — valid/invalid env var combinations (disabled, enabled with certs, missing files, invalid min version)
- `parseTLSMinVersion()` — `"1.2"`, `"1.3"`, invalid values

Integration tests:
- Start server with TLS using test certs, verify HTTPS connection
- Verify TLS 1.1 rejected when min version is 1.2

## Scope

Changes are localized to:
- `server/cmd/main.go` (server setup refactor)
- `server/cmd/main_test.go` (new tests for TLS config parsing)
- `scripts/generate-dev-certs.sh` (new)
- `env.example` (TLS section)
- `.gitignore` (certs exclusion)
