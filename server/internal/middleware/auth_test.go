package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

// RED: Test RequireAuth with valid token
func TestRequireAuth_ValidToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	user := &auth.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Role:     "admin",
		IsActive: true,
	}

	token, err := authService.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("setup failed: GenerateAccessToken() error = %v", err)
	}

	// Create a test handler that checks if user is in context
	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		// Verify user is in context
		ctxUser, ok := r.Context().Value("user").(*auth.User)
		if !ok {
			t.Error("user not found in context")
			return
		}

		if ctxUser.ID != user.ID {
			t.Errorf("context user ID = %v, want %v", ctxUser.ID, user.ID)
		}

		w.WriteHeader(http.StatusOK)
	})

	// Create mock user getter
	userGetter := func(userID string) (*auth.User, error) {
		return user, nil
	}

	middleware := RequireAuth(authService, userGetter)
	handler := middleware(testHandler)

	// Create request with Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if !handlerCalled {
		t.Error("handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusOK)
	}
}

// RED: Test RequireAuth without Authorization header
func TestRequireAuth_MissingAuthHeader(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	userGetter := func(userID string) (*auth.User, error) {
		return nil, nil
	}

	middleware := RequireAuth(authService, userGetter)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if handlerCalled {
		t.Error("handler was called, want not called")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test RequireAuth with invalid token
func TestRequireAuth_InvalidToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	userGetter := func(userID string) (*auth.User, error) {
		return nil, nil
	}

	middleware := RequireAuth(authService, userGetter)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if handlerCalled {
		t.Error("handler was called, want not called")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test RequireAuth with inactive user
func TestRequireAuth_InactiveUser(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	user := &auth.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Role:     "admin",
		IsActive: false, // Inactive
	}

	token, err := authService.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("setup failed: GenerateAccessToken() error = %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	userGetter := func(userID string) (*auth.User, error) {
		return user, nil
	}

	middleware := RequireAuth(authService, userGetter)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if handlerCalled {
		t.Error("handler was called, want not called for inactive user")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test RequireAuth with expired token
func TestRequireAuth_ExpiredToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	// Create expired token manually
	claims := jwt.MapClaims{
		"sub":   uuid.New().String(),
		"email": "test@example.com",
		"role":  "admin",
		"iat":   time.Now().Add(-1 * time.Hour).Unix(),
		"exp":   time.Now().Add(-30 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	userGetter := func(userID string) (*auth.User, error) {
		return nil, nil
	}

	middleware := RequireAuth(authService, userGetter)
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if handlerCalled {
		t.Error("handler was called, want not called for expired token")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test RequireRole with sufficient role (admin >= admin)
func TestRequireRole_SufficientRole(t *testing.T) {
	// Setup
	user := &auth.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  "admin",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequireRole("admin")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", user))
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if !handlerCalled {
		t.Error("handler was not called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusOK)
	}
}

// RED: Test RequireRole with insufficient role (viewer < admin)
func TestRequireRole_InsufficientRole(t *testing.T) {
	// Setup
	user := &auth.User{
		ID:    uuid.New(),
		Email: "viewer@example.com",
		Role:  "viewer",
	}

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	middleware := RequireRole("admin")
	handler := middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(context.WithValue(req.Context(), "user", user))
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if handlerCalled {
		t.Error("handler was called, want not called")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusForbidden)
	}
}

// RED: Test RequireRole hierarchy (admin > manager > viewer)
func TestRequireRole_Hierarchy(t *testing.T) {
	tests := []struct {
		name         string
		userRole     string
		requiredRole string
		shouldAllow  bool
	}{
		{"admin can access admin", "admin", "admin", true},
		{"admin can access manager", "admin", "manager", true},
		{"admin can access viewer", "admin", "viewer", true},
		{"manager cannot access admin", "manager", "admin", false},
		{"manager can access manager", "manager", "manager", true},
		{"manager can access viewer", "manager", "viewer", true},
		{"viewer cannot access admin", "viewer", "admin", false},
		{"viewer cannot access manager", "viewer", "manager", false},
		{"viewer can access viewer", "viewer", "viewer", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &auth.User{
				ID:    uuid.New(),
				Email: "test@example.com",
				Role:  tt.userRole,
			}

			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			middleware := RequireRole(tt.requiredRole)
			handler := middleware(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			req = req.WithContext(context.WithValue(req.Context(), "user", user))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if tt.shouldAllow {
				if !handlerCalled {
					t.Error("handler was not called, want called")
				}
				if w.Code != http.StatusOK {
					t.Errorf("response code = %d, want %d", w.Code, http.StatusOK)
				}
			} else {
				if handlerCalled {
					t.Error("handler was called, want not called")
				}
				if w.Code != http.StatusForbidden {
					t.Errorf("response code = %d, want %d", w.Code, http.StatusForbidden)
				}
			}
		})
	}
}
