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
