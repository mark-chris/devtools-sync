package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/mark-chris/devtools-sync/server/internal/auth"
	"github.com/mark-chris/devtools-sync/server/internal/middleware"
)

// UserByEmailFunc is a function that retrieves a user by email
type UserByEmailFunc func(email string) (*auth.User, error)

// StoreRefreshTokenFunc is a function that stores a refresh token
type StoreRefreshTokenFunc func(rt *auth.RefreshToken) error

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response body
type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// NewLoginHandler creates a new login handler.
// If rateLimiter is non-nil, the rate limit for the client IP is reset on successful login.
// If auditLogger is non-nil, login attempts (success and failure) are audit-logged.
func NewLoginHandler(
	authService *auth.AuthService,
	userByEmail UserByEmailFunc,
	storeRefreshToken StoreRefreshTokenFunc,
	rateLimiter *auth.RateLimiter,
	auditLogger auth.AuditLogger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Invalid request body",
			})
			return
		}

		// Get user by email
		user, err := userByEmail(req.Email)
		if err != nil || user == nil {
			if auditLogger != nil {
				_ = auditLogger.Log(auth.CreateLoginAuditLog(false, nil, req.Email, middleware.GetClientIP(r), r.UserAgent()))
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid credentials",
			})
			return
		}

		// Check if user is active
		if !user.IsActive {
			if auditLogger != nil {
				_ = auditLogger.Log(auth.CreateLoginAuditLog(false, &user.ID, req.Email, middleware.GetClientIP(r), r.UserAgent()))
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid credentials",
			})
			return
		}

		// Verify password
		if err := authService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
			if auditLogger != nil {
				_ = auditLogger.Log(auth.CreateLoginAuditLog(false, &user.ID, req.Email, middleware.GetClientIP(r), r.UserAgent()))
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid credentials",
			})
			return
		}

		// Generate access token
		accessToken, err := authService.GenerateAccessToken(user)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate access token",
			})
			return
		}

		// Generate refresh token
		refreshToken, err := authService.GenerateRefreshToken()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate refresh token",
			})
			return
		}

		// Store refresh token in database
		refreshTokenRecord := &auth.RefreshToken{
			UserID:     user.ID,
			TokenHash:  authService.HashToken(refreshToken),
			ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
			CreatedAt:  time.Now(),
		}

		if err := storeRefreshToken(refreshTokenRecord); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to store refresh token",
			})
			return
		}

		// Reset rate limit on successful login
		if rateLimiter != nil {
			rateLimiter.ResetLimit(middleware.GetClientIP(r))
		}

		// Audit log successful login
		if auditLogger != nil {
			_ = auditLogger.Log(auth.CreateLoginAuditLog(true, &user.ID, req.Email, middleware.GetClientIP(r), r.UserAgent()))
		}

		// Set refresh token cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "refresh_token",
			Value:    refreshToken,
			Path:     "/",
			Expires:  time.Now().Add(7 * 24 * time.Hour),
			MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})

		// Return access token
		writeJSON(w, http.StatusOK, LoginResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   900, // 15 minutes in seconds
		})
	}
}

// GetRefreshTokenFunc is a function that retrieves a refresh token by hash
type GetRefreshTokenFunc func(tokenHash string) (*auth.RefreshToken, error)

// GetUserByIDFunc is a function that retrieves a user by ID
type GetUserByIDFunc func(userID string) (*auth.User, error)

// UpdateRefreshTokenFunc is a function that updates a refresh token
type UpdateRefreshTokenFunc func(rt *auth.RefreshToken) error

// NewRefreshHandler creates a new refresh token handler.
// If auditLogger is non-nil, refresh attempts (success and failure) are audit-logged.
func NewRefreshHandler(
	authService *auth.AuthService,
	getRefreshToken GetRefreshTokenFunc,
	getUserByID GetUserByIDFunc,
	updateRefreshToken UpdateRefreshTokenFunc,
	auditLogger auth.AuditLogger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get refresh token from cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			if auditLogger != nil {
				_ = auditLogger.Log(&auth.AuditLog{
					EventType: auth.AuditRefreshFailure,
					ActorType: auth.ActorTypeUser,
					Details:   map[string]interface{}{"reason": "missing_cookie"},
					ClientIP:  middleware.GetClientIP(r),
					UserAgent: r.UserAgent(),
				})
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Missing refresh token",
			})
			return
		}

		// Hash the token to look it up
		tokenHash := authService.HashToken(cookie.Value)

		// Get token from database
		storedToken, err := getRefreshToken(tokenHash)
		if err != nil || storedToken == nil {
			if auditLogger != nil {
				_ = auditLogger.Log(&auth.AuditLog{
					EventType: auth.AuditRefreshFailure,
					ActorType: auth.ActorTypeUser,
					Details:   map[string]interface{}{"reason": "invalid_token"},
					ClientIP:  middleware.GetClientIP(r),
					UserAgent: r.UserAgent(),
				})
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid refresh token",
			})
			return
		}

		// Check if token is revoked
		if storedToken.RevokedAt != nil {
			if auditLogger != nil {
				_ = auditLogger.Log(&auth.AuditLog{
					EventType: auth.AuditRefreshFailure,
					ActorType: auth.ActorTypeUser,
					Details:   map[string]interface{}{"reason": "revoked_token"},
					ClientIP:  middleware.GetClientIP(r),
					UserAgent: r.UserAgent(),
				})
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid refresh token",
			})
			return
		}

		// Check if token is expired
		if time.Now().After(storedToken.ExpiresAt) {
			if auditLogger != nil {
				_ = auditLogger.Log(&auth.AuditLog{
					EventType: auth.AuditRefreshFailure,
					ActorType: auth.ActorTypeUser,
					Details:   map[string]interface{}{"reason": "expired_token"},
					ClientIP:  middleware.GetClientIP(r),
					UserAgent: r.UserAgent(),
				})
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "Invalid refresh token",
			})
			return
		}

		// Get user
		user, err := getUserByID(storedToken.UserID.String())
		if err != nil || user == nil {
			if auditLogger != nil {
				_ = auditLogger.Log(&auth.AuditLog{
					EventType: auth.AuditRefreshFailure,
					ActorType: auth.ActorTypeUser,
					Details:   map[string]interface{}{"reason": "user_not_found"},
					ClientIP:  middleware.GetClientIP(r),
					UserAgent: r.UserAgent(),
				})
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "User not found",
			})
			return
		}

		// Check if user is active
		if !user.IsActive {
			if auditLogger != nil {
				_ = auditLogger.Log(&auth.AuditLog{
					EventType: auth.AuditRefreshFailure,
					ActorType: auth.ActorTypeUser,
					Details:   map[string]interface{}{"reason": "user_inactive"},
					ClientIP:  middleware.GetClientIP(r),
					UserAgent: r.UserAgent(),
				})
			}
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "User not found or inactive",
			})
			return
		}

		// Generate new access token
		accessToken, err := authService.GenerateAccessToken(user)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "Failed to generate access token",
			})
			return
		}

		// Update last_used_at
		now := time.Now()
		storedToken.LastUsedAt = &now
		_ = updateRefreshToken(storedToken) // Ignore error - don't fail request if update fails

		// Audit log successful refresh
		if auditLogger != nil {
			_ = auditLogger.Log(&auth.AuditLog{
				EventType: auth.AuditRefreshSuccess,
				ActorType: auth.ActorTypeUser,
				ActorID:   &user.ID,
				Details:   map[string]interface{}{},
				ClientIP:  middleware.GetClientIP(r),
				UserAgent: r.UserAgent(),
			})
		}

		// Return new access token
		writeJSON(w, http.StatusOK, LoginResponse{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			ExpiresIn:   900,
		})
	}
}

// RevokeRefreshTokenFunc is a function that revokes a refresh token
type RevokeRefreshTokenFunc func(rt *auth.RefreshToken) error

// NewLogoutHandler creates a new logout handler.
// If auditLogger is non-nil, logout events are audit-logged.
func NewLogoutHandler(
	authService *auth.AuthService,
	getRefreshToken GetRefreshTokenFunc,
	revokeRefreshToken RevokeRefreshTokenFunc,
	auditLogger auth.AuditLogger,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get refresh token from cookie
		cookie, err := r.Cookie("refresh_token")
		if err != nil {
			// No cookie - still return 200 (idempotent)
			clearRefreshTokenCookie(w)
			writeJSON(w, http.StatusOK, map[string]string{
				"message": "Logged out successfully",
			})
			return
		}

		// Hash the token to look it up
		tokenHash := authService.HashToken(cookie.Value)

		// Get token from database
		storedToken, err := getRefreshToken(tokenHash)
		if err == nil && storedToken != nil {
			// Revoke the token
			now := time.Now()
			storedToken.RevokedAt = &now
			_ = revokeRefreshToken(storedToken) // Ignore error - logout is idempotent

			if auditLogger != nil {
				_ = auditLogger.Log(&auth.AuditLog{
					EventType: auth.AuditLogout,
					ActorType: auth.ActorTypeUser,
					ActorID:   &storedToken.UserID,
					Details:   map[string]interface{}{},
					ClientIP:  middleware.GetClientIP(r),
					UserAgent: r.UserAgent(),
				})
			}
		}

		// Clear cookie regardless of token validity (idempotent)
		clearRefreshTokenCookie(w)

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "Logged out successfully",
		})
	}
}

// clearRefreshTokenCookie clears the refresh token cookie
func clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data) // Ignore error - response already started
}
