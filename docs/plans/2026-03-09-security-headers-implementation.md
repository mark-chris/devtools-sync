# Security Headers Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add security headers to the dashboard for both dev (Vite) and production (Go middleware) to protect against XSS, clickjacking, MIME sniffing, and other attacks.

**Architecture:** Security headers applied at two layers — Vite `server.headers` for dev, Go `SecurityHeaders` middleware for production. The Go middleware follows the existing middleware chain pattern (`CORS → MaxBodySize → SecurityHeaders → mux`).

**Tech Stack:** Vite 7, Go 1.25, net/http middleware

---

### Task 1: Go SecurityHeaders middleware — tests

**Files:**
- Create: `server/internal/middleware/security_headers_test.go`

**Step 1: Write the tests**

```go
package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders_AllHeadersSet(t *testing.T) {
	handler := SecurityHeaders(dummyHandler())

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	expected := map[string]string{
		"Content-Security-Policy":   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
		"X-Frame-Options":           "DENY",
		"X-Content-Type-Options":    "nosniff",
		"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
		"Referrer-Policy":           "strict-origin-when-cross-origin",
		"Permissions-Policy":        "camera=(), microphone=(), geolocation=()",
	}

	for header, want := range expected {
		got := w.Header().Get(header)
		if got != want {
			t.Errorf("Header %s: expected %q, got %q", header, want, got)
		}
	}
}

func TestSecurityHeaders_PassesThroughToHandler(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})
	handler := SecurityHeaders(inner)

	req := httptest.NewRequest("POST", "/api/v1/profiles", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", w.Code)
	}
	if w.Body.String() != "created" {
		t.Errorf("Expected body 'created', got %q", w.Body.String())
	}
}

func TestSecurityHeaders_DoesNotOverrideExistingHeaders(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "preserved")
		w.WriteHeader(http.StatusOK)
	})
	handler := SecurityHeaders(inner)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if got := w.Header().Get("X-Custom"); got != "preserved" {
		t.Errorf("Expected X-Custom 'preserved', got %q", got)
	}
	if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
		t.Errorf("Expected X-Frame-Options 'DENY', got %q", got)
	}
}
```

Note: `dummyHandler()` is already defined in `cors_test.go` in the same package.

**Step 2: Run tests to verify they fail**

Run: `cd server && go test -v ./internal/middleware/ -run TestSecurityHeaders`
Expected: FAIL — `SecurityHeaders` not defined

---

### Task 2: Go SecurityHeaders middleware — implementation

**Files:**
- Create: `server/internal/middleware/security_headers.go`

**Step 1: Write the middleware**

```go
package middleware

import "net/http"

// SecurityHeaders adds standard security headers to all responses.
// Headers protect against XSS, clickjacking, MIME sniffing, and other attacks.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}
```

**Step 2: Run tests to verify they pass**

Run: `cd server && go test -v ./internal/middleware/ -run TestSecurityHeaders`
Expected: PASS — all 3 tests

**Step 3: Run full middleware test suite**

Run: `cd server && go test -v ./internal/middleware/`
Expected: All tests pass (existing + new)

**Step 4: Commit**

```bash
git add server/internal/middleware/security_headers.go server/internal/middleware/security_headers_test.go
git commit -m "feat(security): add SecurityHeaders middleware with tests (issue #42)"
```

---

### Task 3: Wire SecurityHeaders into the server

**Files:**
- Modify: `server/cmd/main.go:67` (middleware chain)

**Step 1: Update the middleware chain**

Change line 67 from:
```go
	handler := middleware.CORS(corsOrigins)(middleware.MaxBodySize(maxBodySize)(mux))
```
To:
```go
	handler := middleware.CORS(corsOrigins)(middleware.MaxBodySize(maxBodySize)(middleware.SecurityHeaders(mux)))
```

**Step 2: Run server tests**

Run: `cd server && go test -v ./...`
Expected: All tests pass

**Step 3: Commit**

```bash
git add server/cmd/main.go
git commit -m "feat(security): wire SecurityHeaders middleware into server (issue #42)"
```

---

### Task 4: Add security headers to Vite dev server

**Files:**
- Modify: `dashboard/vite.config.js`

**Step 1: Add headers to vite config**

Replace the full config with:

```javascript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    strictPort: true,
    watch: {
      usePolling: true
    },
    headers: {
      'Content-Security-Policy': "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'",
      'X-Frame-Options': 'DENY',
      'X-Content-Type-Options': 'nosniff',
      'Referrer-Policy': 'strict-origin-when-cross-origin',
      'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
    }
  }
})
```

Note: No HSTS in dev — dev may not use TLS.

**Step 2: Verify dashboard builds**

Run: `cd dashboard && npm run build`
Expected: Build succeeds

**Step 3: Commit**

```bash
git add dashboard/vite.config.js
git commit -m "feat(security): add security headers to Vite dev server (issue #42)"
```
