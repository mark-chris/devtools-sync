//go:build integration

package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mark-chris/devtools-sync/agent/internal/keychain"
)

func TestIntegration_AuthenticatedWorkflow(t *testing.T) {
	kc := keychain.NewMockKeychain()

	// Mock server with all endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/auth/login" && r.Method == http.MethodPost:
			var req LoginRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Email == "test@example.com" && req.Password == "password123" {
				resp := LoginResponse{
					AccessToken: "valid-token",
					TokenType:   "Bearer",
					ExpiresIn:   3600,
				}
				_ = json.NewEncoder(w).Encode(resp)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
			}

		case r.URL.Path == "/api/v1/profiles" && r.Method == http.MethodGet:
			auth := r.Header.Get("Authorization")
			if auth != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_ = json.NewEncoder(w).Encode([]string{"profile1", "profile2"})

		case r.URL.Path == "/api/v1/profiles" && r.Method == http.MethodPost:
			auth := r.Header.Get("Authorization")
			if auth != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusCreated)

		case r.URL.Path == "/api/v1/profiles/test-profile" && r.Method == http.MethodGet:
			auth := r.Header.Get("Authorization")
			if auth != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			profile := Profile{
				Name:       "test-profile",
				CreatedAt:  time.Now(),
				UpdatedAt:  time.Now(),
				Extensions: []Extension{{ID: "test.ext", Version: "1.0.0", Enabled: true}},
			}
			_ = json.NewEncoder(w).Encode(profile)

		case r.URL.Path == "/api/v1/profiles/test-profile" && r.Method == http.MethodPut:
			auth := r.Header.Get("Authorization")
			if auth != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)

		case r.URL.Path == "/api/v1/sync" && r.Method == http.MethodPost:
			auth := r.Header.Get("Authorization")
			if auth != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			resp := SyncResponse{
				Status: "success",
				Merged: []Extension{{ID: "synced.ext", Version: "2.0.0", Enabled: true}},
			}
			_ = json.NewEncoder(w).Encode(resp)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewAuthenticatedClient(server.URL, kc)

	// Test 1: Login
	t.Run("Login", func(t *testing.T) {
		err := client.Login("test@example.com", "password123")
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}

		token, _ := kc.Get(keychain.KeyAccessToken)
		if token != "valid-token" {
			t.Errorf("expected token 'valid-token', got '%s'", token)
		}
	})

	// Test 2: Upload Profile
	t.Run("UploadProfile", func(t *testing.T) {
		err := client.UploadProfile(&Profile{
			Name:       "test-profile",
			Extensions: []Extension{{ID: "test.ext", Version: "1.0.0", Enabled: true}},
		})
		if err != nil {
			t.Fatalf("Upload failed: %v", err)
		}
	})

	// Test 3: List Profiles
	t.Run("ListProfiles", func(t *testing.T) {
		profiles, err := client.ListProfiles()
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(profiles) != 2 {
			t.Errorf("expected 2 profiles, got %d", len(profiles))
		}
	})

	// Test 4: Download Profile
	t.Run("DownloadProfile", func(t *testing.T) {
		profile, err := client.DownloadProfile("test-profile")
		if err != nil {
			t.Fatalf("Download failed: %v", err)
		}
		if profile.Name != "test-profile" {
			t.Errorf("expected 'test-profile', got '%s'", profile.Name)
		}
	})

	// Test 5: Sync
	t.Run("Sync", func(t *testing.T) {
		syncReq := &SyncRequest{
			ProfileName: "test-profile",
			Extensions:  []Extension{{ID: "local.ext", Version: "1.0.0", Enabled: true}},
		}

		data, _ := json.Marshal(syncReq)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/sync", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(data)), nil
		}

		resp, err := client.AuthenticatedRequest(req)
		if err != nil {
			t.Fatalf("Sync failed: %v", err)
		}
		defer resp.Body.Close()

		var syncResp SyncResponse
		body, _ := readLimitedResponse(resp.Body, MaxResponseSize)
		_ = json.Unmarshal(body, &syncResp)

		if syncResp.Status != "success" {
			t.Errorf("expected 'success', got '%s'", syncResp.Status)
		}
	})

	// Test 6: Logout
	t.Run("Logout", func(t *testing.T) {
		err := client.Logout()
		if err != nil {
			t.Fatalf("Logout failed: %v", err)
		}

		_, err = kc.Get(keychain.KeyAccessToken)
		if err == nil {
			t.Error("expected token to be deleted")
		}
	})
}
