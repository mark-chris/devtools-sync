package auth

import (
	"time"

	"github.com/google/uuid"
)

// AuditEvent represents an audit log event type
type AuditEvent string

const (
	// Authentication events
	AuditLoginSuccess       AuditEvent = "auth.login.success"
	AuditLoginFailure       AuditEvent = "auth.login.failure"
	AuditRefreshSuccess     AuditEvent = "auth.refresh.success"
	AuditRefreshFailure     AuditEvent = "auth.refresh.failure"
	AuditLogout             AuditEvent = "auth.logout"
	AuditSessionRevoked     AuditEvent = "auth.session.revoked"

	// User management events
	AuditInviteCreated      AuditEvent = "user.invite.created"
	AuditInviteAccepted     AuditEvent = "user.invite.accepted"
)

// AuditActorType represents the type of actor performing the action
type AuditActorType string

const (
	ActorTypeUser   AuditActorType = "user"
	ActorTypeAgent  AuditActorType = "agent"
	ActorTypeSystem AuditActorType = "system"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID          uuid.UUID
	EventType   AuditEvent
	ActorType   AuditActorType
	ActorID     *uuid.UUID
	GroupID     *uuid.UUID
	TargetType  string
	TargetID    *uuid.UUID
	Details     map[string]interface{}
	ClientIP    string
	UserAgent   string
	CreatedAt   time.Time
}

// AuditLogger is an interface for logging audit events
type AuditLogger interface {
	Log(entry *AuditLog) error
}

// InMemoryAuditLogger is a simple in-memory audit logger for development
type InMemoryAuditLogger struct {
	logs []AuditLog
}

// NewInMemoryAuditLogger creates a new in-memory audit logger
func NewInMemoryAuditLogger() *InMemoryAuditLogger {
	return &InMemoryAuditLogger{
		logs: make([]AuditLog, 0),
	}
}

// Log adds an audit log entry
func (l *InMemoryAuditLogger) Log(entry *AuditLog) error {
	if entry.ID == uuid.Nil {
		entry.ID = uuid.New()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	l.logs = append(l.logs, *entry)
	return nil
}

// GetLogs returns all audit logs (for testing/development)
func (l *InMemoryAuditLogger) GetLogs() []AuditLog {
	return l.logs
}

// CreateLoginAuditLog creates an audit log for login attempts
func CreateLoginAuditLog(success bool, userID *uuid.UUID, email, clientIP, userAgent string) *AuditLog {
	eventType := AuditLoginFailure
	if success {
		eventType = AuditLoginSuccess
	}

	details := map[string]interface{}{
		"email": email,
	}

	if !success {
		details["reason"] = "invalid_credentials"
	}

	return &AuditLog{
		EventType:  eventType,
		ActorType:  ActorTypeUser,
		ActorID:    userID,
		TargetType: "user",
		TargetID:   userID,
		Details:    details,
		ClientIP:   clientIP,
		UserAgent:  userAgent,
	}
}

// CreateInviteAuditLog creates an audit log for invite creation
func CreateInviteAuditLog(inviterID, inviteID uuid.UUID, email, role string) *AuditLog {
	return &AuditLog{
		EventType:  AuditInviteCreated,
		ActorType:  ActorTypeUser,
		ActorID:    &inviterID,
		TargetType: "user_invite",
		TargetID:   &inviteID,
		Details: map[string]interface{}{
			"email": email,
			"role":  role,
		},
	}
}
