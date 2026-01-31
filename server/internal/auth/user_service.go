package auth

import (
	"errors"
	"time"
	"unicode"
)

// ValidatePassword enforces password complexity requirements
func ValidatePassword(password string) error {
	if len(password) < 12 {
		return errors.New("password must be at least 12 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !(hasUpper && hasLower && hasNumber && hasSpecial) {
		return errors.New("password must contain uppercase, lowercase, number, and special character")
	}

	return nil
}

// CreateInviteData creates invite data with a token for a new user
func CreateInviteData(authService *AuthService, email, role, invitedBy string) (*InviteData, string, error) {
	// Generate invite token
	token, err := authService.GenerateRefreshToken()
	if err != nil {
		return nil, "", err
	}

	// Hash the token for storage
	tokenHash := authService.HashToken(token)

	// Create invite data
	inviteData := &InviteData{
		Email:      email,
		TokenHash:  tokenHash,
		Role:       role,
		InvitedBy:  invitedBy,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		AcceptedAt: nil,
	}

	return inviteData, token, nil
}

// ValidateInviteToken checks if an invite token is valid
func ValidateInviteToken(authService *AuthService, inviteData *InviteData, token string) bool {
	// Check if already accepted
	if inviteData.AcceptedAt != nil {
		return false
	}

	// Check if expired
	if time.Now().After(inviteData.ExpiresAt) {
		return false
	}

	// Check if token matches
	expectedHash := authService.HashToken(token)
	return inviteData.TokenHash == expectedHash
}
