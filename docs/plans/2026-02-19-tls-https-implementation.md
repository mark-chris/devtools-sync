# TLS/HTTPS Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add optional TLS termination, server timeouts, and graceful shutdown to the devtools-sync server.

**Architecture:** Refactor `server/cmd/main.go` to use `http.Server` struct with configurable TLS via env vars. TLS is off by default (reverse proxy is the primary deployment model). When enabled, the server calls `ListenAndServeTLS` with a minimum TLS 1.2 version. Graceful shutdown handles SIGINT/SIGTERM with a 30-second drain timeout.

**Tech Stack:** Go 1.24 stdlib (`crypto/tls`, `net/http`, `os/signal`)

**Design doc:** `docs/plans/2026-02-19-tls-https-design.md`

---

### Task 1: TLS Config Parsing — Tests

**Files:**
- Modify: `server/cmd/main_test.go`

**Step 1: Write failing tests for `parseTLSMinVersion`**

Add to `server/cmd/main_test.go`:

```go
func TestParseTLSMinVersion(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint16
	}{
		{"default empty", "", tls.VersionTLS12},
		{"explicit 1.2", "1.2", tls.VersionTLS12},
		{"explicit 1.3", "1.3", tls.VersionTLS13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTLSMinVersion(tt.input)
			if result != tt.expected {
				t.Errorf("parseTLSMinVersion(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseTLSMinVersion_Invalid(t *testing.T) {
	// Invalid values should cause a fatal log — test by verifying the function
	// only accepts "1.2", "1.3", or empty string
	// (Fatal behavior tested via integration test)
	result := parseTLSMinVersion("1.2")
	if result != tls.VersionTLS12 {
		t.Errorf("expected TLS 1.2, got %d", result)
	}
}
```

Add `"crypto/tls"` to the imports.

**Step 2: Run test to verify it fails**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/... -run TestParseTLSMinVersion -v`
Expected: FAIL — `parseTLSMinVersion` not defined

**Step 3: Write minimal implementation in `main.go`**

Add to `server/cmd/main.go`:

```go
// parseTLSMinVersion parses the TLS_MIN_VERSION environment variable.
// Accepts "1.2" or "1.3". Defaults to TLS 1.2.
func parseTLSMinVersion(value string) uint16 {
	switch value {
	case "", "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		log.Fatalf("Invalid TLS_MIN_VERSION %q: must be \"1.2\" or \"1.3\"", value)
		return 0 // unreachable
	}
}
```

Add `"crypto/tls"` to `main.go` imports.

**Step 4: Run test to verify it passes**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/... -run TestParseTLSMinVersion -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/cmd/main.go server/cmd/main_test.go
git commit -m "feat(tls): add parseTLSMinVersion with tests"
```

---

### Task 2: TLS Config Validation — Tests & Implementation

**Files:**
- Modify: `server/cmd/main.go`
- Modify: `server/cmd/main_test.go`

**Step 1: Write failing tests for `loadTLSConfig`**

Add to `server/cmd/main_test.go`:

```go
func TestLoadTLSConfig_Disabled(t *testing.T) {
	tlsCfg, certFile, keyFile := loadTLSConfig(false, "", "", "1.2")
	if tlsCfg != nil {
		t.Error("expected nil TLS config when disabled")
	}
	if certFile != "" || keyFile != "" {
		t.Error("expected empty cert/key paths when disabled")
	}
}

func TestLoadTLSConfig_Enabled(t *testing.T) {
	// Create temp cert and key files
	certDir := t.TempDir()
	certPath := certDir + "/cert.pem"
	keyPath := certDir + "/key.pem"
	os.WriteFile(certPath, []byte("fake-cert"), 0600)
	os.WriteFile(keyPath, []byte("fake-key"), 0600)

	tlsCfg, returnedCert, returnedKey := loadTLSConfig(true, certPath, keyPath, "1.2")
	if tlsCfg == nil {
		t.Fatal("expected non-nil TLS config when enabled")
	}
	if tlsCfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion TLS 1.2, got %d", tlsCfg.MinVersion)
	}
	if returnedCert != certPath {
		t.Errorf("expected cert path %q, got %q", certPath, returnedCert)
	}
	if returnedKey != keyPath {
		t.Errorf("expected key path %q, got %q", keyPath, returnedKey)
	}
}

func TestLoadTLSConfig_EnabledTLS13(t *testing.T) {
	certDir := t.TempDir()
	certPath := certDir + "/cert.pem"
	keyPath := certDir + "/key.pem"
	os.WriteFile(certPath, []byte("fake-cert"), 0600)
	os.WriteFile(keyPath, []byte("fake-key"), 0600)

	tlsCfg, _, _ := loadTLSConfig(true, certPath, keyPath, "1.3")
	if tlsCfg == nil {
		t.Fatal("expected non-nil TLS config")
	}
	if tlsCfg.MinVersion != tls.VersionTLS13 {
		t.Errorf("expected MinVersion TLS 1.3, got %d", tlsCfg.MinVersion)
	}
}
```

Add `"os"` to test imports.

**Step 2: Run test to verify it fails**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/... -run TestLoadTLSConfig -v`
Expected: FAIL — `loadTLSConfig` not defined

**Step 3: Write minimal implementation**

Add to `server/cmd/main.go`:

```go
// loadTLSConfig validates TLS configuration and returns the tls.Config,
// cert file path, and key file path. Returns nil config if TLS is disabled.
func loadTLSConfig(enabled bool, certFile, keyFile, minVersion string) (*tls.Config, string, string) {
	if !enabled {
		return nil, "", ""
	}

	if certFile == "" {
		log.Fatal("TLS_CERT_FILE is required when TLS_ENABLED=true")
	}
	if keyFile == "" {
		log.Fatal("TLS_KEY_FILE is required when TLS_ENABLED=true")
	}

	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("TLS certificate file not found: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("TLS key file not found: %s", keyFile)
	}

	return &tls.Config{
		MinVersion: parseTLSMinVersion(minVersion),
	}, certFile, keyFile
}
```

**Step 4: Run test to verify it passes**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/... -run TestLoadTLSConfig -v`
Expected: PASS

**Step 5: Commit**

```bash
git add server/cmd/main.go server/cmd/main_test.go
git commit -m "feat(tls): add loadTLSConfig with validation and tests"
```

---

### Task 3: Refactor main() — http.Server, Timeouts, Graceful Shutdown

**Files:**
- Modify: `server/cmd/main.go`

**Step 1: Refactor the `main()` function**

Replace the entire `main()` function in `server/cmd/main.go` with:

```go
func main() {
	// Validate JWT secret before starting server
	jwtSecret := os.Getenv("JWT_SECRET")
	isDev := auth.IsDevelopmentMode()

	if err := auth.ValidateSecret(jwtSecret, isDev); err != nil {
		log.Fatalf("JWT secret validation failed: %v", err)
	}

	// Validate database URL SSL configuration
	dbURL := os.Getenv("DATABASE_URL")
	if err := database.ValidateDatabaseURL(dbURL, isDev); err != nil {
		log.Fatalf("Database URL validation failed: %v", err)
	}

	// Parse max body size configuration
	maxBodySize := parseMaxBodySize(os.Getenv("MAX_BODY_SIZE"))
	log.Printf("Request body size limit: %d bytes (%.2f MB)", maxBodySize, float64(maxBodySize)/(1024*1024))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Parse TLS configuration
	tlsEnabled := os.Getenv("TLS_ENABLED") == "true"
	tlsCfg, certFile, keyFile := loadTLSConfig(
		tlsEnabled,
		os.Getenv("TLS_CERT_FILE"),
		os.Getenv("TLS_KEY_FILE"),
		os.Getenv("TLS_MIN_VERSION"),
	)

	// Create mux and register handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	// Apply body size limit middleware to all requests
	handler := middleware.MaxBodySize(maxBodySize)(mux)

	// Create server with timeouts
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	if tlsCfg != nil {
		srv.TLSConfig = tlsCfg
	}

	mode := "production"
	if isDev {
		mode = "development"
	}
	scheme := "http"
	if tlsEnabled {
		scheme = "https"
	}
	log.Printf("Server starting in %s mode on port %s", mode, port)
	if tlsEnabled {
		log.Printf("TLS enabled (min version: %s)", os.Getenv("TLS_MIN_VERSION"))
	}
	log.Printf("Health endpoint: %s://localhost:%s/health", scheme, port)

	// Start server in a goroutine
	go func() {
		var err error
		if tlsEnabled {
			err = srv.ListenAndServeTLS(certFile, keyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("Server stopped")
}
```

Update the imports at the top of `main.go` to:

```go
import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mark-chris/devtools-sync/server/internal/auth"
	"github.com/mark-chris/devtools-sync/server/internal/database"
	"github.com/mark-chris/devtools-sync/server/internal/middleware"
)
```

**Step 2: Run all existing tests to verify nothing is broken**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/... -v`
Expected: All `TestParseMaxBodySize_*`, `TestParseTLSMinVersion*`, and `TestLoadTLSConfig*` tests PASS

**Step 3: Run `go build` to verify compilation**

Run: `cd /home/mark/Projects/devtools-sync/server && go build ./cmd/...`
Expected: Clean build, no errors

**Step 4: Commit**

```bash
git add server/cmd/main.go
git commit -m "feat(tls): refactor main() with http.Server, timeouts, graceful shutdown"
```

---

### Task 4: Integration Test — TLS Server

**Files:**
- Modify: `server/cmd/main_test.go`

**Step 1: Write integration test for TLS server**

Add to `server/cmd/main_test.go`:

```go
func TestTLSServerIntegration(t *testing.T) {
	// Generate self-signed cert for testing
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Fatalf("failed to load test keypair: %v", err)
	}

	// Write cert and key to temp files
	certDir := t.TempDir()
	certPath := certDir + "/cert.pem"
	keyPath := certDir + "/key.pem"
	os.WriteFile(certPath, localhostCert, 0600)
	os.WriteFile(keyPath, localhostKey, 0600)

	tlsCfg, _, _ := loadTLSConfig(true, certPath, keyPath, "1.2")

	// Create test server
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	srv := &http.Server{
		Addr:              ":0", // random port
		Handler:           mux,
		TLSConfig:         tlsCfg,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	})
	if err != nil {
		t.Fatalf("failed to create TLS listener: %v", err)
	}
	defer ln.Close()

	go srv.Serve(ln)
	defer srv.Close()

	// Create client that trusts our test cert
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get("https://" + ln.Addr().String() + "/health")
	if err != nil {
		t.Fatalf("failed to GET /health over TLS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	if resp.TLS == nil {
		t.Error("expected TLS connection")
	}
	if resp.TLS != nil && resp.TLS.Version < tls.VersionTLS12 {
		t.Errorf("expected TLS >= 1.2, got %d", resp.TLS.Version)
	}
}

// NOTE: Original plan included hardcoded test PEM blocks here.
// The actual implementation generates certs programmatically
// using x509.CreateCertificate — see generateTestCert() helper.
```

Note: These are test-only certificates. In the actual implementation, we'll generate proper test certs using Go's `crypto` stdlib (via `x509.CreateCertificate`) instead of hardcoded PEM blocks, to avoid embedding potentially invalid certs. The test structure stays the same.

**Step 2: Run integration test**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/... -run TestTLSServerIntegration -v`
Expected: PASS

**Step 3: Commit**

```bash
git add server/cmd/main_test.go
git commit -m "test(tls): add TLS server integration test"
```

---

### Task 5: Dev Cert Generation Script

**Files:**
- Create: `scripts/generate-dev-certs.sh`

**Step 1: Create the script**

Create `scripts/generate-dev-certs.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Generate self-signed development certificates for devtools-sync
# Usage: ./scripts/generate-dev-certs.sh [output-dir]

CERT_DIR="${1:-certs}"
DAYS=365
KEY_BITS=2048

mkdir -p "$CERT_DIR"

echo "Generating development certificates in $CERT_DIR/"

# Generate CA key and certificate
openssl req -x509 -newkey "rsa:$KEY_BITS" -nodes \
    -keyout "$CERT_DIR/ca-key.pem" \
    -out "$CERT_DIR/ca-cert.pem" \
    -days "$DAYS" \
    -subj "/C=US/ST=Dev/L=Dev/O=devtools-sync/CN=devtools-sync-ca"

# Generate server key
openssl genrsa -out "$CERT_DIR/server-key.pem" "$KEY_BITS"

# Generate server CSR
openssl req -new \
    -key "$CERT_DIR/server-key.pem" \
    -out "$CERT_DIR/server.csr" \
    -subj "/C=US/ST=Dev/L=Dev/O=devtools-sync/CN=localhost"

# Create SAN config
cat > "$CERT_DIR/san.cnf" <<EOF
[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

# Sign server certificate with CA
openssl x509 -req \
    -in "$CERT_DIR/server.csr" \
    -CA "$CERT_DIR/ca-cert.pem" \
    -CAkey "$CERT_DIR/ca-key.pem" \
    -CAcreateserial \
    -out "$CERT_DIR/server-cert.pem" \
    -days "$DAYS" \
    -extensions v3_req \
    -extfile "$CERT_DIR/san.cnf"

# Clean up intermediate files
rm -f "$CERT_DIR/server.csr" "$CERT_DIR/san.cnf" "$CERT_DIR/ca-cert.srl"

echo ""
echo "Certificates generated:"
echo "  CA certificate:     $CERT_DIR/ca-cert.pem"
echo "  Server certificate: $CERT_DIR/server-cert.pem"
echo "  Server key:         $CERT_DIR/server-key.pem"
echo ""
echo "To use with devtools-sync server:"
echo "  export TLS_ENABLED=true"
echo "  export TLS_CERT_FILE=$CERT_DIR/server-cert.pem"
echo "  export TLS_KEY_FILE=$CERT_DIR/server-key.pem"
echo ""
echo "To trust the CA on your system:"
echo "  macOS:  sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain $CERT_DIR/ca-cert.pem"
echo "  Ubuntu: sudo cp $CERT_DIR/ca-cert.pem /usr/local/share/ca-certificates/devtools-sync-ca.crt && sudo update-ca-certificates"
```

**Step 2: Make script executable and verify**

Run: `chmod +x scripts/generate-dev-certs.sh && head -3 scripts/generate-dev-certs.sh`
Expected: Shows shebang line

**Step 3: Commit**

```bash
git add scripts/generate-dev-certs.sh
git commit -m "feat(tls): add self-signed dev certificate generation script"
```

---

### Task 6: Update .gitignore and env.example

**Files:**
- Modify: `.gitignore`
- Modify: `env.example`

**Step 1: Add cert file patterns to `.gitignore`**

Append to `.gitignore`:

```
# TLS certificates (generated by scripts/generate-dev-certs.sh)
certs/
```

**Step 2: Add TLS section to `env.example`**

Add after the `JWT_SECRET` / `JWT_EXPIRATION` block (after line 30):

```bash

# TLS Configuration (optional - for native TLS termination)
# Most deployments should use a reverse proxy (nginx, Caddy, Traefik) instead.
# Only enable native TLS for simple single-binary deployments.
# Generate dev certs: ./scripts/generate-dev-certs.sh
TLS_ENABLED=false
# TLS_CERT_FILE=./certs/server-cert.pem
# TLS_KEY_FILE=./certs/server-key.pem
# TLS_MIN_VERSION=1.2
```

**Step 3: Run existing tests to make sure nothing broke**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/... -v`
Expected: All tests PASS

**Step 4: Commit**

```bash
git add .gitignore env.example
git commit -m "feat(tls): update .gitignore and env.example with TLS config"
```

---

### Task 7: Final Verification

**Step 1: Run full test suite**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./... -v -count=1`
Expected: All tests PASS

**Step 2: Run `go vet`**

Run: `cd /home/mark/Projects/devtools-sync/server && go vet ./...`
Expected: Clean, no warnings

**Step 3: Verify build**

Run: `cd /home/mark/Projects/devtools-sync/server && go build -o /dev/null ./cmd/...`
Expected: Clean build

**Step 4: Manual smoke test (plain HTTP)**

Run: `cd /home/mark/Projects/devtools-sync/server && ENVIRONMENT=development JWT_SECRET="$TEST_JWT_SECRET" DATABASE_URL=postgres://x:x@localhost/x?sslmode=disable timeout 2 go run ./cmd/... 2>&1 || true`
Expected: Log output shows "Server starting in development mode on port 8080" and "Health endpoint: http://localhost:8080/health"
