package database

import (
	"fmt"
	"log"
	"net/url"
	"os"
)

// ValidateDatabaseURL checks that the database connection URL uses appropriate SSL/TLS settings
// for the current environment (production vs development)
func ValidateDatabaseURL(dbURL string, isDev bool) error {
	if dbURL == "" {
		return fmt.Errorf("DATABASE_URL environment variable is required")
	}

	// In development mode, allow any SSL configuration
	if isDev {
		return nil
	}

	// In production, enforce SSL/TLS
	parsed, err := url.Parse(dbURL)
	if err != nil {
		return fmt.Errorf("invalid DATABASE_URL format: %w", err)
	}

	sslMode := parsed.Query().Get("sslmode")

	// Allowed SSL modes for production
	allowedModes := map[string]bool{
		"require":     true, // Require SSL but don't verify server certificate
		"verify-ca":   true, // Require SSL and verify that server certificate is signed by a trusted CA
		"verify-full": true, // Require SSL, verify CA, and verify server hostname matches certificate
	}

	if !allowedModes[sslMode] {
		log.Printf("ERROR: Database SSL/TLS is required in production environment")
		log.Printf("Current sslmode: %q", sslMode)
		log.Printf("Allowed modes: require, verify-ca, verify-full")
		log.Printf("Example: DATABASE_URL=postgres://user:pass@host:5432/db?sslmode=verify-full")
		return fmt.Errorf("database SSL required in production (sslmode=%q not allowed, must be one of: require, verify-ca, verify-full)", sslMode)
	}

	// Log successful SSL mode in production
	log.Printf("Database SSL mode validated: %s", sslMode)

	return nil
}

// IsDevelopmentMode checks if the server is running in development mode
// This is a convenience wrapper around the auth package's check
func IsDevelopmentMode() bool {
	env := os.Getenv("ENVIRONMENT")
	return env == "development" || env == "dev" || env == ""
}
