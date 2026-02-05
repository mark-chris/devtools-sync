package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	baseURL := "http://localhost:8080"
	client := NewClient(baseURL)

	if client == nil {
		t.Fatal("expected client to be created, got nil")
	}

	if client.baseURL != baseURL {
		t.Errorf("expected baseURL %s, got %s", baseURL, client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestHealth(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		responseBody   interface{}
		wantError      bool
		wantStatus     string
		wantService    string
	}{
		{
			name:       "successful health check",
			statusCode: http.StatusOK,
			responseBody: HealthResponse{
				Status:  "healthy",
				Service: "devtools-sync-server",
			},
			wantError:   false,
			wantStatus:  "healthy",
			wantService: "devtools-sync-server",
		},
		{
			name:         "server error",
			statusCode:   http.StatusInternalServerError,
			responseBody: map[string]string{"error": "internal error"},
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/health" {
					t.Errorf("expected path /health, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			health, err := client.Health()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if !tt.wantError {
				if health.Status != tt.wantStatus {
					t.Errorf("expected status %s, got %s", tt.wantStatus, health.Status)
				}
				if health.Service != tt.wantService {
					t.Errorf("expected service %s, got %s", tt.wantService, health.Service)
				}
			}
		})
	}
}

func TestReadLimitedResponse(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		maxSize   int64
		wantError error
		wantLen   int
	}{
		{
			name:      "response within limit",
			data:      "hello world",
			maxSize:   100,
			wantError: nil,
			wantLen:   11,
		},
		{
			name:      "response at exact limit",
			data:      "12345",
			maxSize:   5,
			wantError: nil,
			wantLen:   5,
		},
		{
			name:      "response exceeds limit",
			data:      "this is too long",
			maxSize:   5,
			wantError: ErrResponseTooLarge,
			wantLen:   0,
		},
		{
			name:      "empty response",
			data:      "",
			maxSize:   100,
			wantError: nil,
			wantLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.data)
			body, err := readLimitedResponse(reader, tt.maxSize)

			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Errorf("expected error %v, got %v", tt.wantError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(body) != tt.wantLen {
				t.Errorf("expected body length %d, got %d", tt.wantLen, len(body))
			}
		})
	}
}

func TestHealthResponseTooLarge(t *testing.T) {
	// Create a server that returns a response larger than MaxResponseSize
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write more data than MaxResponseSize (1MB + extra)
		largeData := strings.Repeat("x", int(MaxResponseSize)+1000)
		_, _ = w.Write([]byte(largeData))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Health()

	if !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestUploadProfile(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantError  bool
	}{
		{
			name:       "successful upload",
			statusCode: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "successful upload with 201",
			statusCode: http.StatusCreated,
			wantError:  false,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/profiles" {
					t.Errorf("expected path /api/v1/profiles, got %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected method POST, got %s", r.Method)
				}

				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client := NewClient(server.URL)
			testProfile := &Profile{
				Name:       "test",
				Extensions: []Extension{{ID: "ext1", Version: "1.0.0", Enabled: true}},
			}

			err := client.UploadProfile(testProfile)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		})
	}
}

func TestListProfiles(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody interface{}
		wantError    bool
		wantProfiles []string
	}{
		{
			name:         "successful list",
			statusCode:   http.StatusOK,
			responseBody: []string{"profile1", "profile2"},
			wantError:    false,
			wantProfiles: []string{"profile1", "profile2"},
		},
		{
			name:         "empty list",
			statusCode:   http.StatusOK,
			responseBody: []string{},
			wantError:    false,
			wantProfiles: []string{},
		},
		{
			name:         "server error",
			statusCode:   http.StatusInternalServerError,
			responseBody: map[string]string{"error": "internal error"},
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/v1/profiles" {
					t.Errorf("expected path /api/v1/profiles, got %s", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			profiles, err := client.ListProfiles()

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if !tt.wantError {
				if len(profiles) != len(tt.wantProfiles) {
					t.Errorf("expected %d profiles, got %d", len(tt.wantProfiles), len(profiles))
				}
				for i, name := range tt.wantProfiles {
					if profiles[i] != name {
						t.Errorf("expected profile %s at index %d, got %s", name, i, profiles[i])
					}
				}
			}
		})
	}
}

func TestDownloadProfile(t *testing.T) {
	tests := []struct {
		name         string
		profileName  string
		statusCode   int
		responseBody interface{}
		wantError    bool
		errorContains string
	}{
		{
			name:        "successful download",
			profileName: "test",
			statusCode:  http.StatusOK,
			responseBody: Profile{
				Name:       "test",
				Extensions: []Extension{{ID: "ext1", Version: "1.0.0", Enabled: true}},
			},
			wantError: false,
		},
		{
			name:          "profile not found",
			profileName:   "nonexistent",
			statusCode:    http.StatusNotFound,
			responseBody:  map[string]string{"error": "not found"},
			wantError:     true,
			errorContains: "not found",
		},
		{
			name:         "server error",
			profileName:  "test",
			statusCode:   http.StatusInternalServerError,
			responseBody: map[string]string{"error": "internal error"},
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/api/v1/profiles/" + tt.profileName
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {
					t.Fatalf("failed to encode response: %v", err)
				}
			}))
			defer server.Close()

			client := NewClient(server.URL)
			profile, err := client.DownloadProfile(tt.profileName)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.wantError && tt.errorContains != "" {
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got: %s", tt.errorContains, err.Error())
				}
			}

			if !tt.wantError {
				if profile.Name != tt.profileName {
					t.Errorf("expected profile name %s, got %s", tt.profileName, profile.Name)
				}
			}
		})
	}
}

func TestRetryableRequest_Success(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)

	resp, err := client.retryableRequest(req)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetryableRequest_RetryOn503(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)

	resp, err := client.retryableRequest(req)
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetryableRequest_NoRetryOn400(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)

	resp, err := client.retryableRequest(req)
	if err != nil {
		t.Fatalf("expected response, got error: %v", err)
	}
	defer resp.Body.Close()

	if attempts != 1 {
		t.Errorf("expected 1 attempt (no retry), got %d", attempts)
	}

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
