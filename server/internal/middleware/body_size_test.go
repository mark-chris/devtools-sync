package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestMaxBodySize_WithinLimit(t *testing.T) {
	// Create a handler that reads the body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Read " + strconv.Itoa(len(body)) + " bytes"))
	})

	// Wrap with MaxBodySize middleware (1KB limit)
	middleware := MaxBodySize(1024)
	wrappedHandler := middleware(handler)

	// Create request with small body (100 bytes)
	body := bytes.Repeat([]byte("a"), 100)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMaxBodySize_ExceedsLimit(t *testing.T) {
	// Create a handler that tries to read the body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			// Error reading body - return 413
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			_, _ = w.Write([]byte(`{"error":"Request body too large"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with MaxBodySize middleware (1KB limit)
	middleware := MaxBodySize(1024)
	wrappedHandler := middleware(handler)

	// Create request with large body (2KB)
	body := bytes.Repeat([]byte("a"), 2048)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413, got %d", w.Code)
	}
}

func TestMaxBodySize_ExactLimit(t *testing.T) {
	// Create a handler that reads the body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Read " + strconv.Itoa(len(body)) + " bytes"))
	})

	// Wrap with MaxBodySize middleware (1KB limit)
	middleware := MaxBodySize(1024)
	wrappedHandler := middleware(handler)

	// Create request with exactly 1KB body
	body := bytes.Repeat([]byte("a"), 1024)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Assert - exactly at limit should succeed
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for exact limit, got %d", w.Code)
	}
}

func TestMaxBodySize_EmptyBody(t *testing.T) {
	// Create a handler that reads the body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read body: %v", err)
		}
		if len(body) != 0 {
			t.Errorf("Expected empty body, got %d bytes", len(body))
		}
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with MaxBodySize middleware
	middleware := MaxBodySize(1024)
	wrappedHandler := middleware(handler)

	// Create request with empty body
	req := httptest.NewRequest("POST", "/test", nil)
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestMaxBodySize_GETRequest(t *testing.T) {
	// GET requests shouldn't have bodies, but middleware should not interfere
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap with MaxBodySize middleware
	middleware := MaxBodySize(1024)
	wrappedHandler := middleware(handler)

	// Create GET request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleMaxBytesError_WithMaxBytesError(t *testing.T) {
	w := httptest.NewRecorder()
	// Create the specific error message that MaxBytesReader returns
	err := &maxBytesError{}

	handled := HandleMaxBytesError(w, err, 1024)

	if !handled {
		t.Error("HandleMaxBytesError() should have handled max bytes error")
	}

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status 413, got %d", w.Code)
	}
}

// maxBytesError simulates the error from MaxBytesReader
type maxBytesError struct{}

func (e *maxBytesError) Error() string {
	return "http: request body too large"
}

func TestMaxBodySize_MultipleReads(t *testing.T) {
	// Create a handler that reads the body multiple times
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First read
		chunk1 := make([]byte, 512)
		n1, err1 := r.Body.Read(chunk1)
		if err1 != nil && err1 != io.EOF {
			t.Logf("First read error: %v", err1)
		}

		// Second read
		chunk2 := make([]byte, 512)
		n2, err2 := r.Body.Read(chunk2)
		if err2 != nil && err2 != io.EOF {
			t.Logf("Second read error: %v", err2)
		}

		total := n1 + n2
		if total > 1024 {
			w.WriteHeader(http.StatusRequestEntityTooLarge)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// Wrap with MaxBodySize middleware
	middleware := MaxBodySize(1024)
	wrappedHandler := middleware(handler)

	// Create request with 1500 bytes (exceeds limit)
	body := bytes.Repeat([]byte("a"), 1500)
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Execute
	wrappedHandler.ServeHTTP(w, req)

	// Assert - should detect oversized body
	if w.Code == http.StatusOK {
		t.Error("Expected non-200 status for oversized body")
	}
}
