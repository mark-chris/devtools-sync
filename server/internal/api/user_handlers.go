package api

import (
	"encoding/json"
	"net/http"
	"regexp"
	"time"

	"github.com/google/uuid"
	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const userContextKey contextKey = "user"

// StoreInviteFunc is a function that stores a user invite
type StoreInviteFunc func(invite *auth.UserInvite) error

// InviteRequest represents the invite request body
type InviteRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

// InviteResponse represents the invite response body
type InviteResponse struct {
	InviteURL string `json:"invite_url"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// NewInviteHandler creates a new invite handler
func NewInviteHandler(
	authService *auth.AuthService,
	storeInvite StoreInviteFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (set by RequireAuth middleware)
		user, ok := r.Context().Value(userContextKey).(*auth.User)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "User not found in context",
			})
			return
		}

		// Parse request
		var req InviteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Validate email
		if !emailRegex.MatchString(req.Email) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid email address",
			})
			return
		}

		// Validate role
		validRoles := map[string]bool{
			"viewer":  true,
			"manager": true,
			"admin":   true,
		}

		if !validRoles[req.Role] {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid role. Must be viewer, manager, or admin",
			})
			return
		}

		// Generate invite token
		inviteToken, err := authService.GenerateRefreshToken()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate invite token",
			})
			return
		}

		// Create invite record
		invite := &auth.UserInvite{
			ID:        uuid.New(),
			Email:     req.Email,
			TokenHash: authService.HashToken(inviteToken),
			Role:      req.Role,
			InvitedBy: user.ID,
			ExpiresAt: time.Now().Add(48 * time.Hour),
			CreatedAt: time.Now(),
		}

		// Store invite
		if err := storeInvite(invite); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to store invite",
			})
			return
		}

		// Generate invite URL
		inviteURL := "https://app.example.com/accept-invite?token=" + inviteToken

		writeJSON(w, http.StatusOK, InviteResponse{
			InviteURL: inviteURL,
		})
	}
}

// GetInviteByTokenFunc is a function that retrieves an invite by token hash
type GetInviteByTokenFunc func(tokenHash string) (*auth.UserInvite, error)

// CreateUserFunc is a function that creates a new user
type CreateUserFunc func(user *auth.User) error

// MarkInviteAcceptedFunc is a function that marks an invite as accepted
type MarkInviteAcceptedFunc func(invite *auth.UserInvite) error

// AcceptInviteRequest represents the accept invite request body
type AcceptInviteRequest struct {
	Token       string `json:"token"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// AcceptInviteResponse represents the accept invite response body
type AcceptInviteResponse struct {
	Message string `json:"message"`
}

// NewAcceptInviteHandler creates a new accept invite handler
func NewAcceptInviteHandler(
	authService *auth.AuthService,
	getInviteByToken GetInviteByTokenFunc,
	createUser CreateUserFunc,
	markInviteAccepted MarkInviteAcceptedFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var req AcceptInviteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Hash token to look it up
		tokenHash := authService.HashToken(req.Token)

		// Get invite
		invite, err := getInviteByToken(tokenHash)
		if err != nil || invite == nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid or expired invite token",
			})
			return
		}

		// Check if already accepted
		if invite.AcceptedAt != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid or expired invite token",
			})
			return
		}

		// Check if expired
		if time.Now().After(invite.ExpiresAt) {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid or expired invite token",
			})
			return
		}

		// Validate password
		if err := auth.ValidatePassword(req.Password); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
			return
		}

		// Hash password
		passwordHash, err := authService.HashPassword(req.Password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to hash password",
			})
			return
		}

		// Create user
		now := time.Now()
		user := &auth.User{
			ID:           uuid.New(),
			Email:        invite.Email,
			PasswordHash: passwordHash,
			DisplayName:  req.DisplayName,
			Role:         invite.Role,
			IsActive:     true,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		if err := createUser(user); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to create user",
			})
			return
		}

		// Mark invite as accepted
		invite.AcceptedAt = &now
		_ = markInviteAccepted(invite) // Ignore error - user already created, invite update is best-effort

		writeJSON(w, http.StatusOK, AcceptInviteResponse{
			Message: "Account created successfully",
		})
	}
}
