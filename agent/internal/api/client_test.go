package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
