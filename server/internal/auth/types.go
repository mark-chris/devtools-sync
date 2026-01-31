package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a dashboard user
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	DisplayName  string
	Role         string
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    *time.Time
}

// Claims represents JWT token claims
type Claims struct {
	UserID string
	Email  string
	Role   string
}

// RefreshToken represents a database refresh token record
type RefreshToken struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	TokenHash   string
	DeviceName  string
	UserAgent   string
	ClientIP    string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	LastUsedAt  *time.Time
	CreatedAt   time.Time
}

// UserInvite represents a database user invite record
type UserInvite struct {
	ID         uuid.UUID
	Email      string
	TokenHash  string
	Role       string
	InvitedBy  uuid.UUID
	AcceptedAt *time.Time
	ExpiresAt  time.Time
	CreatedAt  time.Time
}

// InviteData represents invite information for creation
type InviteData struct {
	Email      string
	TokenHash  string
	Role       string
	InvitedBy  string
	ExpiresAt  time.Time
	AcceptedAt *time.Time
}
