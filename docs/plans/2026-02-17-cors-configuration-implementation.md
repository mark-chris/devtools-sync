# CORS Configuration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add CORS middleware so the React dashboard can make cross-origin API calls with credentials.

**Architecture:** Hand-rolled CORS middleware using existing `func(http.Handler) http.Handler` pattern. Configured via `CORS_ALLOWED_ORIGINS` env var (comma-separated). Applied as outermost middleware so preflight OPTIONS responses bypass body size limits and auth. TDD throughout.

**Tech Stack:** Go stdlib (`net/http`, `net/http/httptest`, `strings`)

**Design doc:** `docs/plans/2026-02-17-cors-configuration-design.md`

---

### Task 1: CORS middleware — tests

**Files:**
- Create: `server/internal/middleware/cors_test.go`

**Step 1: Write all CORS middleware tests**

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// dummyHandler is a simple handler that returns 200 OK
func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
}

func TestCORS_AllowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(dummyHandler())

	req := httptest.NewRequest("GET", "/api/v1/profiles", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Expected Allow-Origin 'http://localhost:5173', got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Expected Allow-Credentials 'true', got %q", got)
	}
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(dummyHandler())

	req := httptest.NewRequest("GET", "/api/v1/profiles", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 (request still served), got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no Allow-Origin header, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("Expected no Allow-Credentials header, got %q", got)
	}
}

func TestCORS_PreflightAllowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(dummyHandler())

	req := httptest.NewRequest("OPTIONS", "/api/v1/profiles", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Errorf("Expected Allow-Origin 'http://localhost:5173', got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, OPTIONS" {
		t.Errorf("Expected Allow-Methods, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Headers"); got != "Authorization, Content-Type" {
		t.Errorf("Expected Allow-Headers, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("Expected Max-Age '86400', got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("Expected Allow-Credentials 'true', got %q", got)
	}
	// Body should be empty — preflight should not reach inner handler
	if w.Body.String() != "" {
		t.Errorf("Expected empty body for preflight, got %q", w.Body.String())
	}
}

func TestCORS_PreflightDisallowedOrigin(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(dummyHandler())

	req := httptest.NewRequest("OPTIONS", "/api/v1/profiles", nil)
	req.Header.Set("Origin", "http://evil.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// OPTIONS from disallowed origin: no CORS headers, pass through to next handler
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no Allow-Origin header, got %q", got)
	}
	if got := w.Header().Get("Access-Control-Allow-Methods"); got != "" {
		t.Errorf("Expected no Allow-Methods header, got %q", got)
	}
}

func TestCORS_NoOriginHeader(t *testing.T) {
	handler := CORS([]string{"http://localhost:5173"})(dummyHandler())

	req := httptest.NewRequest("GET", "/api/v1/profiles", nil)
	// No Origin header — same-origin request
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no Allow-Origin header, got %q", got)
	}
}

func TestCORS_EmptyConfig(t *testing.T) {
	handler := CORS([]string{})(dummyHandler())

	req := httptest.NewRequest("GET", "/api/v1/profiles", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no Allow-Origin with empty config, got %q", got)
	}
}

func TestCORS_NilConfig(t *testing.T) {
	handler := CORS(nil)(dummyHandler())

	req := httptest.NewRequest("GET", "/api/v1/profiles", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Expected no Allow-Origin with nil config, got %q", got)
	}
}

func TestCORS_VaryHeaderAlwaysSet(t *testing.T) {
	tests := []struct {
		name   string
		origin string
	}{
		{"with allowed origin", "http://localhost:5173"},
		{"with disallowed origin", "http://evil.com"},
		{"with no origin", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CORS([]string{"http://localhost:5173"})(dummyHandler())

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if got := w.Header().Get("Vary"); got != "Origin" {
				t.Errorf("Expected Vary 'Origin', got %q", got)
			}
		})
	}
}

func TestCORS_MultipleAllowedOrigins(t *testing.T) {
	origins := []string{"http://localhost:5173", "https://dashboard.example.com"}
	handler := CORS(origins)(dummyHandler())

	tests := []struct {
		name          string
		origin        string
		expectAllowed bool
	}{
		{"first origin", "http://localhost:5173", true},
		{"second origin", "https://dashboard.example.com", true},
		{"unknown origin", "http://other.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Origin", tt.origin)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			got := w.Header().Get("Access-Control-Allow-Origin")
			if tt.expectAllowed && got != tt.origin {
				t.Errorf("Expected Allow-Origin %q, got %q", tt.origin, got)
			}
			if !tt.expectAllowed && got != "" {
				t.Errorf("Expected no Allow-Origin, got %q", got)
			}
		})
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./internal/middleware/ -run TestCORS -v`
Expected: FAIL — `CORS` function not defined

---

### Task 2: CORS middleware — implementation

**Files:**
- Create: `server/internal/middleware/cors.go`

**Step 1: Implement CORS middleware**

```go
package middleware

import "net/http"

// CORS returns middleware that handles Cross-Origin Resource Sharing.
// allowedOrigins is a list of origins permitted to make cross-origin requests.
// If empty, no CORS headers are set (secure by default).
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	originSet := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		originSet[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Always set Vary: Origin for correct HTTP caching
			w.Header().Set("Vary", "Origin")

			origin := r.Header.Get("Origin")
			if origin == "" || !originSet[origin] {
				next.ServeHTTP(w, r)
				return
			}

			// Origin is allowed — set CORS response headers
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "86400")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

**Step 2: Run tests to verify they pass**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./internal/middleware/ -run TestCORS -v`
Expected: All PASS

**Step 3: Commit**

```bash
git add server/internal/middleware/cors.go server/internal/middleware/cors_test.go
git commit -m "feat: add CORS middleware with tests

Closes #62 partially — middleware only, not yet wired."
```

---

### Task 3: Wire CORS into server and add parseCORSOrigins

**Files:**
- Modify: `server/cmd/main.go:36-50` (add CORS parsing + middleware wrapping)
- Modify: `server/cmd/main_test.go` (add parseCORSOrigins tests)

**Step 1: Write parseCORSOrigins tests in main_test.go**

Append to `server/cmd/main_test.go`:

```go
func TestParseCORSOrigins_Empty(t *testing.T) {
	result := parseCORSOrigins("")
	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %v", result)
	}
}

func TestParseCORSOrigins_Single(t *testing.T) {
	result := parseCORSOrigins("http://localhost:5173")
	if len(result) != 1 || result[0] != "http://localhost:5173" {
		t.Errorf("Expected [http://localhost:5173], got %v", result)
	}
}

func TestParseCORSOrigins_Multiple(t *testing.T) {
	result := parseCORSOrigins("http://localhost:5173,https://dashboard.example.com")
	if len(result) != 2 {
		t.Fatalf("Expected 2 origins, got %d", len(result))
	}
	if result[0] != "http://localhost:5173" {
		t.Errorf("Expected first origin 'http://localhost:5173', got %q", result[0])
	}
	if result[1] != "https://dashboard.example.com" {
		t.Errorf("Expected second origin 'https://dashboard.example.com', got %q", result[1])
	}
}

func TestParseCORSOrigins_WhitespaceHandling(t *testing.T) {
	result := parseCORSOrigins(" http://localhost:5173 , https://dashboard.example.com ")
	if len(result) != 2 {
		t.Fatalf("Expected 2 origins, got %d", len(result))
	}
	if result[0] != "http://localhost:5173" {
		t.Errorf("Expected trimmed first origin, got %q", result[0])
	}
	if result[1] != "https://dashboard.example.com" {
		t.Errorf("Expected trimmed second origin, got %q", result[1])
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/ -run TestParseCORSOrigins -v`
Expected: FAIL — `parseCORSOrigins` not defined

**Step 3: Add parseCORSOrigins and wire CORS middleware in main.go**

Add `"strings"` to imports.

Add `parseCORSOrigins` function after `parseMaxBodySize`:

```go
// parseCORSOrigins parses the CORS_ALLOWED_ORIGINS environment variable.
// Returns a slice of origin strings. Empty input returns nil.
func parseCORSOrigins(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			origins = append(origins, p)
		}
	}
	return origins
}
```

In `main()`, after `parseMaxBodySize` line (around line 37), add:

```go
corsOrigins := parseCORSOrigins(os.Getenv("CORS_ALLOWED_ORIGINS"))
```

Replace line 50 (`handler := middleware.MaxBodySize(maxBodySize)(mux)`) with:

```go
handler := middleware.CORS(corsOrigins)(middleware.MaxBodySize(maxBodySize)(mux))
```

After the mode log line (line 56), add:

```go
if len(corsOrigins) > 0 {
	log.Printf("CORS allowed origins: %v", corsOrigins)
}
```

**Step 4: Run all tests**

Run: `cd /home/mark/Projects/devtools-sync/server && go test ./cmd/ -v && go test ./internal/middleware/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add server/cmd/main.go server/cmd/main_test.go
git commit -m "feat: wire CORS middleware into server with env var config

CORS_ALLOWED_ORIGINS env var (comma-separated) configures allowed origins.
Empty/unset means no CORS headers (secure by default)."
```

---

### Task 4: Update env.example

**Files:**
- Modify: `env.example:43-48`

**Step 1: Add CORS config to env.example**

In the "Dashboard Configuration" section, add after the `VITE_API_URL` line:

```
# Allowed CORS origins (comma-separated)
# Required if dashboard runs on a different origin than the API server
# Example: http://localhost:5173,https://dashboard.example.com
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

**Step 2: Commit**

```bash
git add env.example
git commit -m "docs: add CORS_ALLOWED_ORIGINS to env.example"
```
