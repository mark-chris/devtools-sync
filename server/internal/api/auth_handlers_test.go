package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

// RED: Test login with valid credentials
func TestLoginHandler_ValidCredentials(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	// Create a test user with known password
	password := "SecurePass123!"
	passwordHash, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("setup failed: HashPassword() error = %v", err)
	}

	testUser := &auth.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "admin",
		IsActive:     true,
	}

	// Mock user lookup
	userByEmail := func(email string) (*auth.User, error) {
		if email == testUser.Email {
			return testUser, nil
		}
		return nil, nil
	}

	// Mock refresh token storage
	var storedRefreshToken *auth.RefreshToken
	storeRefreshToken := func(rt *auth.RefreshToken) error {
		storedRefreshToken = rt
		return nil
	}

	handler := NewLoginHandler(authService, userByEmail, storeRefreshToken)

	// Create request
	body := map[string]string{
		"email":    testUser.Email,
		"password": password,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify access token in response
	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		t.Error("response missing access_token")
	}

	// Verify token type
	if response["token_type"] != "Bearer" {
		t.Errorf("token_type = %v, want Bearer", response["token_type"])
	}

	// Verify expires_in
	expiresIn, ok := response["expires_in"].(float64)
	if !ok || expiresIn != 900 { // 15 minutes = 900 seconds
		t.Errorf("expires_in = %v, want 900", expiresIn)
	}

	// Verify refresh token cookie is set
	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
			break
		}
	}

	if refreshTokenCookie == nil {
		t.Fatal("refresh_token cookie not set")
	}

	if refreshTokenCookie.Value == "" {
		t.Error("refresh_token cookie value is empty")
	}

	if !refreshTokenCookie.HttpOnly {
		t.Error("refresh_token cookie is not HttpOnly")
	}

	if !refreshTokenCookie.Secure {
		t.Error("refresh_token cookie is not Secure")
	}

	if refreshTokenCookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("refresh_token cookie SameSite = %v, want Strict", refreshTokenCookie.SameSite)
	}

	// Verify refresh token was stored
	if storedRefreshToken == nil {
		t.Fatal("refresh token was not stored")
	}

	if storedRefreshToken.UserID != testUser.ID {
		t.Errorf("stored refresh token UserID = %v, want %v", storedRefreshToken.UserID, testUser.ID)
	}
}

// RED: Test login with invalid credentials
func TestLoginHandler_InvalidCredentials(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	testUser := &auth.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "$2a$12$invalid",
		Role:         "admin",
		IsActive:     true,
	}

	userByEmail := func(email string) (*auth.User, error) {
		if email == testUser.Email {
			return testUser, nil
		}
		return nil, nil
	}

	storeRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewLoginHandler(authService, userByEmail, storeRefreshToken)

	body := map[string]string{
		"email":    testUser.Email,
		"password": "WrongPassword123!",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test login with inactive user
func TestLoginHandler_InactiveUser(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	password := "SecurePass123!"
	passwordHash, _ := authService.HashPassword(password)

	testUser := &auth.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: passwordHash,
		Role:         "admin",
		IsActive:     false, // Inactive
	}

	userByEmail := func(email string) (*auth.User, error) {
		if email == testUser.Email {
			return testUser, nil
		}
		return nil, nil
	}

	storeRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewLoginHandler(authService, userByEmail, storeRefreshToken)

	body := map[string]string{
		"email":    testUser.Email,
		"password": password,
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test login with non-existent user
func TestLoginHandler_UserNotFound(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	userByEmail := func(email string) (*auth.User, error) {
		return nil, nil // User not found
	}

	storeRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewLoginHandler(authService, userByEmail, storeRefreshToken)

	body := map[string]string{
		"email":    "nonexistent@example.com",
		"password": "AnyPassword123!",
	}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test login with malformed JSON
func TestLoginHandler_MalformedJSON(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	userByEmail := func(email string) (*auth.User, error) {
		return nil, nil
	}

	storeRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewLoginHandler(authService, userByEmail, storeRefreshToken)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// RED: Test refresh with valid token
func TestRefreshHandler_ValidToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	testUser := &auth.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Role:     "admin",
		IsActive: true,
	}

	// Generate and store refresh token
	refreshToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(refreshToken)

	storedToken := &auth.RefreshToken{
		UserID:    testUser.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	getRefreshToken := func(tokenHash string) (*auth.RefreshToken, error) {
		if tokenHash == storedToken.TokenHash {
			return storedToken, nil
		}
		return nil, nil
	}

	getUserByID := func(userID string) (*auth.User, error) {
		if userID == testUser.ID.String() {
			return testUser, nil
		}
		return nil, nil
	}

	var updatedToken *auth.RefreshToken
	updateRefreshToken := func(rt *auth.RefreshToken) error {
		updatedToken = rt
		return nil
	}

	handler := NewRefreshHandler(authService, getRefreshToken, getUserByID, updateRefreshToken)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify new access token
	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		t.Error("response missing access_token")
	}

	// Verify last_used_at was updated
	if updatedToken == nil {
		t.Fatal("refresh token was not updated")
	}

	if updatedToken.LastUsedAt == nil {
		t.Error("last_used_at was not set")
	}
}

// RED: Test refresh with missing cookie
func TestRefreshHandler_MissingCookie(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	getRefreshToken := func(tokenHash string) (*auth.RefreshToken, error) {
		return nil, nil
	}

	getUserByID := func(userID string) (*auth.User, error) {
		return nil, nil
	}

	updateRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewRefreshHandler(authService, getRefreshToken, getUserByID, updateRefreshToken)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test refresh with expired token
func TestRefreshHandler_ExpiredToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	refreshToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(refreshToken)

	storedToken := &auth.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt: time.Now().Add(-8 * 24 * time.Hour),
	}

	getRefreshToken := func(tokenHash string) (*auth.RefreshToken, error) {
		return storedToken, nil
	}

	getUserByID := func(userID string) (*auth.User, error) {
		return nil, nil
	}

	updateRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewRefreshHandler(authService, getRefreshToken, getUserByID, updateRefreshToken)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test refresh with revoked token
func TestRefreshHandler_RevokedToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	refreshToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(refreshToken)

	revokedAt := time.Now().Add(-1 * time.Hour)
	storedToken := &auth.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		RevokedAt: &revokedAt, // Revoked
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	getRefreshToken := func(tokenHash string) (*auth.RefreshToken, error) {
		return storedToken, nil
	}

	getUserByID := func(userID string) (*auth.User, error) {
		return nil, nil
	}

	updateRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewRefreshHandler(authService, getRefreshToken, getUserByID, updateRefreshToken)

	req := httptest.NewRequest("POST", "/auth/refresh", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusUnauthorized {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

// RED: Test logout with valid token
func TestLogoutHandler_ValidToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	refreshToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(refreshToken)

	storedToken := &auth.RefreshToken{
		UserID:    uuid.New(),
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		CreatedAt: time.Now(),
	}

	getRefreshToken := func(tokenHash string) (*auth.RefreshToken, error) {
		return storedToken, nil
	}

	var revokedToken *auth.RefreshToken
	revokeRefreshToken := func(rt *auth.RefreshToken) error {
		revokedToken = rt
		return nil
	}

	handler := NewLogoutHandler(authService, getRefreshToken, revokeRefreshToken)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: refreshToken,
	})
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify token was revoked
	if revokedToken == nil {
		t.Fatal("refresh token was not revoked")
	}

	if revokedToken.RevokedAt == nil {
		t.Error("revoked_at was not set")
	}

	// Verify cookie was cleared
	cookies := w.Result().Cookies()
	var refreshTokenCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "refresh_token" {
			refreshTokenCookie = cookie
			break
		}
	}

	if refreshTokenCookie == nil {
		t.Fatal("refresh_token cookie not found in response")
	}

	if refreshTokenCookie.Value != "" {
		t.Error("refresh_token cookie was not cleared")
	}

	if refreshTokenCookie.MaxAge != -1 {
		t.Errorf("refresh_token cookie MaxAge = %d, want -1", refreshTokenCookie.MaxAge)
	}
}

// RED: Test logout without cookie (should still return 200)
func TestLogoutHandler_MissingCookie(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	getRefreshToken := func(tokenHash string) (*auth.RefreshToken, error) {
		return nil, nil
	}

	revokeRefreshToken := func(rt *auth.RefreshToken) error {
		return nil
	}

	handler := NewLogoutHandler(authService, getRefreshToken, revokeRefreshToken)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert - should still return 200 (idempotent)
	if w.Code != http.StatusOK {
		t.Errorf("response code = %d, want %d (logout is idempotent)", w.Code, http.StatusOK)
	}
}
