package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

func TestRateLimitMiddleware_AllowsUnderLimit(t *testing.T) {
	rl := auth.NewRateLimiter(time.Hour, time.Hour, 1000)
	defer rl.Stop()

	handler := RateLimit(rl, 5, 15*time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("attempt %d: expected 200, got %d", i+1, rr.Code)
		}
	}
}

func TestRateLimitMiddleware_BlocksOverLimit(t *testing.T) {
	rl := auth.NewRateLimiter(time.Hour, time.Hour, 1000)
	defer rl.Stop()

	handler := RateLimit(rl, 3, 15*time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use up the limit
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = "10.0.0.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", rr.Code)
	}

	// Check response body
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["error"] == "" {
		t.Error("expected error message in response")
	}
}

func TestRateLimitMiddleware_RetryAfterHeader(t *testing.T) {
	rl := auth.NewRateLimiter(time.Hour, time.Hour, 1000)
	defer rl.Stop()

	window := 15 * time.Minute
	handler := RateLimit(rl, 1, window)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Use up the limit
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Next request should have Retry-After
	req = httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	retryAfter := rr.Header().Get("Retry-After")
	if retryAfter != "900" {
		t.Errorf("expected Retry-After '900', got '%s'", retryAfter)
	}
}

func TestRateLimitMiddleware_DifferentIPsIndependent(t *testing.T) {
	rl := auth.NewRateLimiter(time.Hour, time.Hour, 1000)
	defer rl.Stop()

	handler := RateLimit(rl, 1, 15*time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// IP 1 uses its limit
	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// IP 2 should still be allowed
	req = httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "10.0.0.2:12345"
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for different IP, got %d", rr.Code)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18, 150.172.238.178")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := GetClientIP(req)
	if ip != "203.0.113.50" {
		t.Errorf("expected '203.0.113.50', got '%s'", ip)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "203.0.113.100")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := GetClientIP(req)
	if ip != "203.0.113.100" {
		t.Errorf("expected '203.0.113.100', got '%s'", ip)
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:54321"

	ip := GetClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", ip)
	}
}

func TestGetClientIP_RemoteAddrNoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1"

	ip := GetClientIP(req)
	if ip != "192.168.1.1" {
		t.Errorf("expected '192.168.1.1', got '%s'", ip)
	}
}

func TestGetClientIP_IPv6(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "[::1]:54321"

	ip := GetClientIP(req)
	if ip != "::1" {
		t.Errorf("expected '::1', got '%s'", ip)
	}
}

func TestGetClientIP_XForwardedForPrecedence(t *testing.T) {
	// X-Forwarded-For takes precedence over X-Real-IP
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	req.Header.Set("X-Real-IP", "10.0.0.2")
	req.RemoteAddr = "127.0.0.1:12345"

	ip := GetClientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("expected '10.0.0.1' (X-Forwarded-For precedence), got '%s'", ip)
	}
}
