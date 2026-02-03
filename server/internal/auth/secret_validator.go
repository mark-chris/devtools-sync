package auth

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Known weak/default secrets that should never be used in production
var knownWeakSecrets = []string{
	"local-dev-jwt-secret-not-for-production",
	"changeme",
	"secret",
	"password",
	"test",
	"dev",
	"development",
}

// ValidateSecret validates the JWT secret meets security requirements
// isDev indicates if the server is running in development mode
func ValidateSecret(secret string, isDev bool) error {
	// Check if secret is empty
	if secret == "" {
		return errors.New("JWT_SECRET environment variable is required")
	}

	// Check against known weak secrets FIRST (before length check)
	// This allows dev mode to accept them with a warning
	for _, weak := range knownWeakSecrets {
		if secret == weak {
			if isDev {
				log.Printf("WARNING: Using default JWT secret '%s...' - NOT FOR PRODUCTION USE", secret[:min(20, len(secret))])
				return nil
			}
			return fmt.Errorf("default/weak JWT secret not allowed in production environment")
		}
	}

	// Check minimum length (for non-weak secrets)
	if len(secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters (got %d)", len(secret))
	}

	return nil
}

// IsDevelopmentMode determines if the server is running in development mode
// Checks ENVIRONMENT and GO_ENV environment variables
func IsDevelopmentMode() bool {
	env := os.Getenv("ENVIRONMENT")
	if env == "development" || env == "dev" {
		return true
	}

	goEnv := os.Getenv("GO_ENV")
	if goEnv == "development" || goEnv == "dev" {
		return true
	}

	// Default to production mode for safety
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
