package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

// Helper function to add user to context
func contextWithUser(ctx context.Context, user *auth.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// RED: Test creating invite as admin
func TestInviteHandler_AsAdmin(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	adminUser := &auth.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  "admin",
	}

	var storedInvite *auth.UserInvite
	storeInvite := func(invite *auth.UserInvite) error {
		storedInvite = invite
		return nil
	}

	handler := NewInviteHandler(authService, storeInvite)

	body := map[string]string{
		"email": "newuser@example.com",
		"role":  "viewer",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/invite", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// Add user to context (normally done by RequireAuth middleware)
	req = req.WithContext(contextWithUser(req.Context(), adminUser))
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

	// Verify invite URL is present
	inviteURL, ok := response["invite_url"].(string)
	if !ok || inviteURL == "" {
		t.Error("response missing invite_url")
	}

	// Verify invite was stored
	if storedInvite == nil {
		t.Fatal("invite was not stored")
	}

	if storedInvite.Email != "newuser@example.com" {
		t.Errorf("stored invite email = %v, want newuser@example.com", storedInvite.Email)
	}

	if storedInvite.Role != "viewer" {
		t.Errorf("stored invite role = %v, want viewer", storedInvite.Role)
	}

	if storedInvite.InvitedBy != adminUser.ID {
		t.Errorf("stored invite invited_by = %v, want %v", storedInvite.InvitedBy, adminUser.ID)
	}
}

// RED: Test creating invite with invalid role
func TestInviteHandler_InvalidRole(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	adminUser := &auth.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  "admin",
	}

	storeInvite := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewInviteHandler(authService, storeInvite)

	body := map[string]string{
		"email": "newuser@example.com",
		"role":  "superadmin", // Invalid role
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/invite", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUser(req.Context(), adminUser))
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// RED: Test creating invite with invalid email
func TestInviteHandler_InvalidEmail(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	adminUser := &auth.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  "admin",
	}

	storeInvite := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewInviteHandler(authService, storeInvite)

	body := map[string]string{
		"email": "invalid-email", // Invalid email
		"role":  "viewer",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/invite", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUser(req.Context(), adminUser))
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestCanInviteRole(t *testing.T) {
	tests := []struct {
		inviter string
		target  string
		want    bool
	}{
		{"admin", "admin", true},
		{"admin", "manager", true},
		{"admin", "viewer", true},
		{"manager", "admin", false},
		{"manager", "manager", true},
		{"manager", "viewer", true},
		{"viewer", "admin", false},
		{"viewer", "manager", false},
		{"viewer", "viewer", true},
		{"unknown", "viewer", false},
		{"admin", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.inviter+"_invites_"+tt.target, func(t *testing.T) {
			got := canInviteRole(tt.inviter, tt.target)
			if got != tt.want {
				t.Errorf("canInviteRole(%q, %q) = %v, want %v", tt.inviter, tt.target, got, tt.want)
			}
		})
	}
}

func TestInviteHandler_ManagerCanInviteViewer(t *testing.T) {
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	managerUser := &auth.User{
		ID:    uuid.New(),
		Email: "manager@example.com",
		Role:  "manager",
	}

	storeInvite := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewInviteHandler(authService, storeInvite)

	body := map[string]string{"email": "new@example.com", "role": "viewer"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/users/invite", bytes.NewReader(bodyBytes))
	req = req.WithContext(contextWithUser(req.Context(), managerUser))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestInviteHandler_ManagerCanInviteManager(t *testing.T) {
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	managerUser := &auth.User{
		ID:    uuid.New(),
		Email: "manager@example.com",
		Role:  "manager",
	}

	storeInvite := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewInviteHandler(authService, storeInvite)

	body := map[string]string{"email": "new@example.com", "role": "manager"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/users/invite", bytes.NewReader(bodyBytes))
	req = req.WithContext(contextWithUser(req.Context(), managerUser))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestInviteHandler_ManagerCannotInviteAdmin(t *testing.T) {
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	managerUser := &auth.User{
		ID:    uuid.New(),
		Email: "manager@example.com",
		Role:  "manager",
	}

	storeInvite := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewInviteHandler(authService, storeInvite)

	body := map[string]string{"email": "new@example.com", "role": "admin"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/users/invite", bytes.NewReader(bodyBytes))
	req = req.WithContext(contextWithUser(req.Context(), managerUser))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestInviteHandler_AdminCanInviteAdmin(t *testing.T) {
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	adminUser := &auth.User{
		ID:    uuid.New(),
		Email: "admin@example.com",
		Role:  "admin",
	}

	storeInvite := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewInviteHandler(authService, storeInvite)

	body := map[string]string{"email": "new@example.com", "role": "admin"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/users/invite", bytes.NewReader(bodyBytes))
	req = req.WithContext(contextWithUser(req.Context(), adminUser))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// RED: Test accepting invite with valid token
func TestAcceptInviteHandler_ValidToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	inviteToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(inviteToken)

	storedInvite := &auth.UserInvite{
		ID:         uuid.New(),
		Email:      "newuser@example.com",
		TokenHash:  tokenHash,
		Role:       "viewer",
		InvitedBy:  uuid.New(),
		AcceptedAt: nil,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		CreatedAt:  time.Now(),
	}

	getInviteByToken := func(tokenHash string) (*auth.UserInvite, error) {
		if tokenHash == storedInvite.TokenHash {
			return storedInvite, nil
		}
		return nil, nil
	}

	var createdUser *auth.User
	createUser := func(user *auth.User) error {
		createdUser = user
		return nil
	}

	var updatedInvite *auth.UserInvite
	markInviteAccepted := func(invite *auth.UserInvite) error {
		updatedInvite = invite
		return nil
	}

	handler := NewAcceptInviteHandler(authService, getInviteByToken, createUser, markInviteAccepted)

	body := map[string]string{
		"token":        inviteToken,
		"password":     "SecurePass123!",
		"display_name": "New User",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/accept-invite", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusOK {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify user was created
	if createdUser == nil {
		t.Fatal("user was not created")
	}

	if createdUser.Email != storedInvite.Email {
		t.Errorf("created user email = %v, want %v", createdUser.Email, storedInvite.Email)
	}

	if createdUser.Role != storedInvite.Role {
		t.Errorf("created user role = %v, want %v", createdUser.Role, storedInvite.Role)
	}

	if createdUser.DisplayName != "New User" {
		t.Errorf("created user display_name = %v, want New User", createdUser.DisplayName)
	}

	if !createdUser.IsActive {
		t.Error("created user is not active")
	}

	// Verify password was hashed
	if createdUser.PasswordHash == "SecurePass123!" {
		t.Error("password was not hashed")
	}

	// Verify invite was marked accepted
	if updatedInvite == nil {
		t.Fatal("invite was not updated")
	}

	if updatedInvite.AcceptedAt == nil {
		t.Error("invite accepted_at was not set")
	}
}

// RED: Test accepting invite with invalid password
func TestAcceptInviteHandler_InvalidPassword(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	inviteToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(inviteToken)

	storedInvite := &auth.UserInvite{
		ID:         uuid.New(),
		Email:      "newuser@example.com",
		TokenHash:  tokenHash,
		Role:       "viewer",
		InvitedBy:  uuid.New(),
		AcceptedAt: nil,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		CreatedAt:  time.Now(),
	}

	getInviteByToken := func(tokenHash string) (*auth.UserInvite, error) {
		return storedInvite, nil
	}

	createUser := func(user *auth.User) error {
		return nil
	}

	markInviteAccepted := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewAcceptInviteHandler(authService, getInviteByToken, createUser, markInviteAccepted)

	body := map[string]string{
		"token":        inviteToken,
		"password":     "weak", // Too weak
		"display_name": "New User",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/accept-invite", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// RED: Test accepting invite with expired token
func TestAcceptInviteHandler_ExpiredToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	inviteToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(inviteToken)

	storedInvite := &auth.UserInvite{
		ID:         uuid.New(),
		Email:      "newuser@example.com",
		TokenHash:  tokenHash,
		Role:       "viewer",
		InvitedBy:  uuid.New(),
		AcceptedAt: nil,
		ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired
		CreatedAt:  time.Now().Add(-49 * time.Hour),
	}

	getInviteByToken := func(tokenHash string) (*auth.UserInvite, error) {
		return storedInvite, nil
	}

	createUser := func(user *auth.User) error {
		return nil
	}

	markInviteAccepted := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewAcceptInviteHandler(authService, getInviteByToken, createUser, markInviteAccepted)

	body := map[string]string{
		"token":        inviteToken,
		"password":     "SecurePass123!",
		"display_name": "New User",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/accept-invite", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// RED: Test accepting already accepted invite
func TestAcceptInviteHandler_AlreadyAccepted(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	authService := auth.NewAuthService(secretKey)

	inviteToken, _ := authService.GenerateRefreshToken()
	tokenHash := authService.HashToken(inviteToken)

	acceptedAt := time.Now().Add(-1 * time.Hour)
	storedInvite := &auth.UserInvite{
		ID:         uuid.New(),
		Email:      "newuser@example.com",
		TokenHash:  tokenHash,
		Role:       "viewer",
		InvitedBy:  uuid.New(),
		AcceptedAt: &acceptedAt, // Already accepted
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		CreatedAt:  time.Now().Add(-24 * time.Hour),
	}

	getInviteByToken := func(tokenHash string) (*auth.UserInvite, error) {
		return storedInvite, nil
	}

	createUser := func(user *auth.User) error {
		return nil
	}

	markInviteAccepted := func(invite *auth.UserInvite) error {
		return nil
	}

	handler := NewAcceptInviteHandler(authService, getInviteByToken, createUser, markInviteAccepted)

	body := map[string]string{
		"token":        inviteToken,
		"password":     "SecurePass123!",
		"display_name": "New User",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest("POST", "/users/accept-invite", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	if w.Code != http.StatusBadRequest {
		t.Errorf("response code = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
