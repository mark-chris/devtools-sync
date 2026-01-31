package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mark-chris/devtools-sync/server/internal/auth"
)

// UserGetter is a function that retrieves a user by ID
type UserGetter func(userID string) (*auth.User, error)

// RequireAuth is middleware that validates JWT tokens and attaches user to context
func RequireAuth(authService *auth.AuthService, userGetter UserGetter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "Missing or invalid authorization header",
				})
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			claims, err := authService.ValidateAccessToken(token)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "Invalid or expired token",
				})
				return
			}

			// Load user from database
			user, err := userGetter(claims.UserID)
			if err != nil || user == nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "User not found",
				})
				return
			}

			// Check if user is active
			if !user.IsActive {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "User not found or inactive",
				})
				return
			}

			// Attach user to context
			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole is middleware that checks if the user has the required role
func RequireRole(minRole string) func(http.Handler) http.Handler {
	roleHierarchy := map[string]int{
		"viewer":  1,
		"manager": 2,
		"admin":   3,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context (set by RequireAuth)
			user, ok := r.Context().Value("user").(*auth.User)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error": "User not found in context",
				})
				return
			}

			// Check role hierarchy
			userLevel, userExists := roleHierarchy[user.Role]
			minLevel, minExists := roleHierarchy[minRole]

			if !userExists || !minExists || userLevel < minLevel {
				writeJSON(w, http.StatusForbidden, map[string]string{
					"error": "Insufficient permissions",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
