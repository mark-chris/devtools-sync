package auth

import (
	"testing"
	"time"
)

// RED: Test password validation - valid password
func TestValidatePassword_Valid(t *testing.T) {
	password := "SecurePass123!"

	err := ValidatePassword(password)

	if err != nil {
		t.Errorf("ValidatePassword() error = %v, want nil for valid password", err)
	}
}

// RED: Test password validation - too short
func TestValidatePassword_TooShort(t *testing.T) {
	password := "Short1!"

	err := ValidatePassword(password)

	if err == nil {
		t.Error("ValidatePassword() error = nil, want error for password < 12 chars")
	}
}

// RED: Test password validation - missing uppercase
func TestValidatePassword_MissingUppercase(t *testing.T) {
	password := "securepass123!"

	err := ValidatePassword(password)

	if err == nil {
		t.Error("ValidatePassword() error = nil, want error for missing uppercase")
	}
}

// RED: Test password validation - missing lowercase
func TestValidatePassword_MissingLowercase(t *testing.T) {
	password := "SECUREPASS123!"

	err := ValidatePassword(password)

	if err == nil {
		t.Error("ValidatePassword() error = nil, want error for missing lowercase")
	}
}

// RED: Test password validation - missing number
func TestValidatePassword_MissingNumber(t *testing.T) {
	password := "SecurePassWord!"

	err := ValidatePassword(password)

	if err == nil {
		t.Error("ValidatePassword() error = nil, want error for missing number")
	}
}

// RED: Test password validation - missing special character
func TestValidatePassword_MissingSpecial(t *testing.T) {
	password := "SecurePass123"

	err := ValidatePassword(password)

	if err == nil {
		t.Error("ValidatePassword() error = nil, want error for missing special character")
	}
}

// RED: Test password validation - exactly 12 chars (boundary)
func TestValidatePassword_Exactly12Chars(t *testing.T) {
	password := "SecureP@ss12"

	err := ValidatePassword(password)

	if err != nil {
		t.Errorf("ValidatePassword() error = %v, want nil for 12-char password", err)
	}
}

// RED: Test creating invite data (business logic)
func TestCreateInviteData(t *testing.T) {
	// Setup
	authService := NewAuthService([]byte("test-secret"))
	email := "newuser@example.com"
	role := "viewer"
	invitedBy := "admin-user-id"

	// Act
	inviteData, token, err := CreateInviteData(authService, email, role, invitedBy)

	// Assert
	if err != nil {
		t.Fatalf("CreateInviteData() error = %v, want nil", err)
	}

	if inviteData == nil {
		t.Fatal("CreateInviteData() returned nil invite data")
	}

	if token == "" {
		t.Fatal("CreateInviteData() returned empty token")
	}

	// Verify invite data fields
	if inviteData.Email != email {
		t.Errorf("inviteData.Email = %v, want %v", inviteData.Email, email)
	}

	if inviteData.Role != role {
		t.Errorf("inviteData.Role = %v, want %v", inviteData.Role, role)
	}

	if inviteData.InvitedBy != invitedBy {
		t.Errorf("inviteData.InvitedBy = %v, want %v", inviteData.InvitedBy, invitedBy)
	}

	// Verify token hash is set
	if inviteData.TokenHash == "" {
		t.Error("inviteData.TokenHash is empty, want hash")
	}

	// Verify token hash matches the returned token
	expectedHash := authService.HashToken(token)
	if inviteData.TokenHash != expectedHash {
		t.Errorf("inviteData.TokenHash = %v, want %v", inviteData.TokenHash, expectedHash)
	}

	// Verify expiration is ~48 hours from now
	expectedExpiry := time.Now().Add(48 * time.Hour)
	if inviteData.ExpiresAt.Before(expectedExpiry.Add(-1*time.Minute)) ||
		inviteData.ExpiresAt.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("inviteData.ExpiresAt = %v, want within 1 min of %v", inviteData.ExpiresAt, expectedExpiry)
	}
}

// RED: Test validating invite token
func TestValidateInviteToken_Valid(t *testing.T) {
	// Setup
	authService := NewAuthService([]byte("test-secret"))
	inviteData, token, err := CreateInviteData(authService, "test@example.com", "viewer", "admin-id")
	if err != nil {
		t.Fatalf("setup failed: CreateInviteData() error = %v", err)
	}

	// Act
	isValid := ValidateInviteToken(authService, inviteData, token)

	// Assert
	if !isValid {
		t.Error("ValidateInviteToken() = false, want true for valid token")
	}
}

// RED: Test validating invite token with wrong token
func TestValidateInviteToken_WrongToken(t *testing.T) {
	// Setup
	authService := NewAuthService([]byte("test-secret"))
	inviteData, _, err := CreateInviteData(authService, "test@example.com", "viewer", "admin-id")
	if err != nil {
		t.Fatalf("setup failed: CreateInviteData() error = %v", err)
	}

	wrongToken := "wrong-token-123"

	// Act
	isValid := ValidateInviteToken(authService, inviteData, wrongToken)

	// Assert
	if isValid {
		t.Error("ValidateInviteToken() = true, want false for wrong token")
	}
}

// RED: Test validating expired invite
func TestValidateInviteToken_Expired(t *testing.T) {
	// Setup
	authService := NewAuthService([]byte("test-secret"))
	inviteData, token, err := CreateInviteData(authService, "test@example.com", "viewer", "admin-id")
	if err != nil {
		t.Fatalf("setup failed: CreateInviteData() error = %v", err)
	}

	// Manually expire the invite
	inviteData.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Act
	isValid := ValidateInviteToken(authService, inviteData, token)

	// Assert
	if isValid {
		t.Error("ValidateInviteToken() = true, want false for expired invite")
	}
}

// RED: Test validating already accepted invite
func TestValidateInviteToken_AlreadyAccepted(t *testing.T) {
	// Setup
	authService := NewAuthService([]byte("test-secret"))
	inviteData, token, err := CreateInviteData(authService, "test@example.com", "viewer", "admin-id")
	if err != nil {
		t.Fatalf("setup failed: CreateInviteData() error = %v", err)
	}

	// Mark as accepted
	now := time.Now()
	inviteData.AcceptedAt = &now

	// Act
	isValid := ValidateInviteToken(authService, inviteData, token)

	// Assert
	if isValid {
		t.Error("ValidateInviteToken() = true, want false for already accepted invite")
	}
}
