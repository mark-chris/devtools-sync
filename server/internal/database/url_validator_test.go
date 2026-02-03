package database

import (
	"os"
	"testing"
)

func TestValidateDatabaseURL_EmptyURL(t *testing.T) {
	err := ValidateDatabaseURL("", false)
	if err == nil {
		t.Error("Expected error for empty DATABASE_URL")
	}
	if err != nil && err.Error() != "DATABASE_URL environment variable is required" {
		t.Errorf("Wrong error message: %v", err)
	}
}

func TestValidateDatabaseURL_InvalidFormat(t *testing.T) {
	err := ValidateDatabaseURL("not-a-valid-url", false)
	if err == nil {
		t.Error("Expected error for invalid URL format")
	}
}

func TestValidateDatabaseURL_ProductionWithSSLDisabled(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=disable"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err == nil {
		t.Error("Expected error for sslmode=disable in production")
	}
}

func TestValidateDatabaseURL_ProductionWithNoSSLMode(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err == nil {
		t.Error("Expected error for missing sslmode in production")
	}
}

func TestValidateDatabaseURL_ProductionWithRequire(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=require"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err != nil {
		t.Errorf("Expected no error for sslmode=require in production, got: %v", err)
	}
}

func TestValidateDatabaseURL_ProductionWithVerifyCA(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=verify-ca"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err != nil {
		t.Errorf("Expected no error for sslmode=verify-ca in production, got: %v", err)
	}
}

func TestValidateDatabaseURL_ProductionWithVerifyFull(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=verify-full"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err != nil {
		t.Errorf("Expected no error for sslmode=verify-full in production, got: %v", err)
	}
}

func TestValidateDatabaseURL_DevelopmentWithSSLDisabled(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=disable"
	err := ValidateDatabaseURL(dbURL, true) // development mode
	if err != nil {
		t.Errorf("Expected no error for sslmode=disable in development, got: %v", err)
	}
}

func TestValidateDatabaseURL_DevelopmentWithNoSSLMode(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db"
	err := ValidateDatabaseURL(dbURL, true) // development mode
	if err != nil {
		t.Errorf("Expected no error for missing sslmode in development, got: %v", err)
	}
}

func TestValidateDatabaseURL_DevelopmentWithSSLEnabled(t *testing.T) {
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=require"
	err := ValidateDatabaseURL(dbURL, true) // development mode
	if err != nil {
		t.Errorf("Expected no error for sslmode=require in development, got: %v", err)
	}
}

func TestIsDevelopmentMode_Development(t *testing.T) {
	originalEnv := os.Getenv("ENVIRONMENT")
	defer func() { _ = os.Setenv("ENVIRONMENT", originalEnv) }()

	tests := []struct {
		envValue string
		expected bool
	}{
		{"development", true},
		{"dev", true},
		{"", true}, // default to dev when not set
		{"production", false},
		{"prod", false},
		{"staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			_ = os.Setenv("ENVIRONMENT", tt.envValue)
			result := IsDevelopmentMode()
			if result != tt.expected {
				t.Errorf("IsDevelopmentMode() with ENVIRONMENT=%q = %v, want %v", tt.envValue, result, tt.expected)
			}
		})
	}
}

func TestValidateDatabaseURL_ProductionWithPrefer(t *testing.T) {
	// sslmode=prefer is not secure enough for production
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=prefer"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err == nil {
		t.Error("Expected error for sslmode=prefer in production (not strict enough)")
	}
}

func TestValidateDatabaseURL_ProductionWithAllow(t *testing.T) {
	// sslmode=allow is not secure enough for production
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=allow"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err == nil {
		t.Error("Expected error for sslmode=allow in production (not strict enough)")
	}
}

func TestValidateDatabaseURL_ComplexURLWithSSL(t *testing.T) {
	// Test with additional query parameters
	dbURL := "postgres://user:pass@localhost:5432/db?sslmode=verify-full&connect_timeout=10&application_name=devtools-sync"
	err := ValidateDatabaseURL(dbURL, false) // production mode
	if err != nil {
		t.Errorf("Expected no error for complex URL with sslmode=verify-full, got: %v", err)
	}
}
