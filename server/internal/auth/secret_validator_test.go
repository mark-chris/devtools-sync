package auth

import (
	"os"
	"strings"
	"testing"
)

func TestValidateSecret_EmptySecret(t *testing.T) {
	err := ValidateSecret("", false)
	if err == nil {
		t.Fatal("ValidateSecret() error = nil, want error for empty secret")
	}

	if !strings.Contains(err.Error(), "required") {
		t.Errorf("ValidateSecret() error = %v, want error mentioning 'required'", err)
	}
}

func TestValidateSecret_TooShort(t *testing.T) {
	tests := []struct {
		name   string
		secret string
	}{
		{"1 char", "a"},
		{"10 chars", "1234567890"},
		{"31 chars", "1234567890123456789012345678901"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecret(tt.secret, false)
			if err == nil {
				t.Fatalf("ValidateSecret(%q) error = nil, want error for short secret", tt.secret)
			}

			if !strings.Contains(err.Error(), "32 characters") {
				t.Errorf("ValidateSecret() error = %v, want error mentioning minimum length", err)
			}
		})
	}
}

func TestValidateSecret_MinimumLength(t *testing.T) {
	// Exactly 32 characters - should pass
	secret := "12345678901234567890123456789012"
	err := ValidateSecret(secret, false)
	if err != nil {
		t.Errorf("ValidateSecret() error = %v, want nil for 32-char secret", err)
	}
}

func TestValidateSecret_WeakSecrets_Production(t *testing.T) {
	weakSecrets := []string{
		"local-dev-jwt-secret-not-for-production",
		"changeme",
		"secret",
		"password",
		"test",
		"dev",
		"development",
	}

	for _, secret := range weakSecrets {
		t.Run(secret, func(t *testing.T) {
			err := ValidateSecret(secret, false)
			if err == nil {
				t.Fatalf("ValidateSecret(%q, false) error = nil, want error in production", secret)
			}

			if !strings.Contains(err.Error(), "production") {
				t.Errorf("ValidateSecret() error = %v, want error mentioning 'production'", err)
			}
		})
	}
}

func TestValidateSecret_WeakSecrets_Development(t *testing.T) {
	// In development mode, weak secrets should be allowed with a warning
	weakSecrets := []string{
		"local-dev-jwt-secret-not-for-production",
		"changeme",
		"secret",
	}

	for _, secret := range weakSecrets {
		t.Run(secret, func(t *testing.T) {
			err := ValidateSecret(secret, true)
			if err != nil {
				t.Errorf("ValidateSecret(%q, true) error = %v, want nil in development", secret, err)
			}
		})
	}
}

func TestValidateSecret_StrongSecret(t *testing.T) {
	strongSecrets := []string{
		"a-very-strong-secret-key-with-sufficient-entropy",
		"12345678901234567890123456789012345678901234567890",
		"my-super-secure-production-jwt-secret-2024",
	}

	for _, secret := range strongSecrets {
		t.Run(secret, func(t *testing.T) {
			// Should pass in both dev and production
			if err := ValidateSecret(secret, false); err != nil {
				t.Errorf("ValidateSecret(%q, false) error = %v, want nil", secret, err)
			}

			if err := ValidateSecret(secret, true); err != nil {
				t.Errorf("ValidateSecret(%q, true) error = %v, want nil", secret, err)
			}
		})
	}
}

func TestIsDevelopmentMode_Development(t *testing.T) {
	tests := []struct {
		name  string
		env   string
		value string
	}{
		{"ENVIRONMENT=development", "ENVIRONMENT", "development"},
		{"ENVIRONMENT=dev", "ENVIRONMENT", "dev"},
		{"GO_ENV=development", "GO_ENV", "development"},
		{"GO_ENV=dev", "GO_ENV", "dev"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original value
			original := os.Getenv(tt.env)
			defer func() {
				_ = os.Setenv(tt.env, original)
			}()

			_ = os.Setenv(tt.env, tt.value)

			if !IsDevelopmentMode() {
				t.Errorf("IsDevelopmentMode() = false, want true for %s=%s", tt.env, tt.value)
			}
		})
	}
}

func TestIsDevelopmentMode_Production(t *testing.T) {
	// Save original values
	origEnv := os.Getenv("ENVIRONMENT")
	origGoEnv := os.Getenv("GO_ENV")
	defer func() {
		_ = os.Setenv("ENVIRONMENT", origEnv)
		_ = os.Setenv("GO_ENV", origGoEnv)
	}()

	tests := []struct {
		name        string
		environment string
		goEnv       string
	}{
		{"Empty vars", "", ""},
		{"ENVIRONMENT=production", "production", ""},
		{"ENVIRONMENT=prod", "prod", ""},
		{"GO_ENV=production", "", "production"},
		{"Both production", "production", "production"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("ENVIRONMENT", tt.environment)
			_ = os.Setenv("GO_ENV", tt.goEnv)

			if IsDevelopmentMode() {
				t.Errorf("IsDevelopmentMode() = true, want false for non-dev environment")
			}
		})
	}
}

func TestIsDevelopmentMode_DefaultsToProduction(t *testing.T) {
	// Save original values
	origEnv := os.Getenv("ENVIRONMENT")
	origGoEnv := os.Getenv("GO_ENV")
	defer func() {
		_ = os.Setenv("ENVIRONMENT", origEnv)
		_ = os.Setenv("GO_ENV", origGoEnv)
	}()

	// Clear both variables
	_ = os.Unsetenv("ENVIRONMENT")
	_ = os.Unsetenv("GO_ENV")

	if IsDevelopmentMode() {
		t.Error("IsDevelopmentMode() = true, want false (should default to production)")
	}
}
