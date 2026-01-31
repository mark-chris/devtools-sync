# Authentication and Authorization System Design

**Issue:** #9 - Implement authentication and authorization system
**Date:** 2026-01-30
**Status:** Approved

## Overview

JWT-based authentication system for dashboard users with invite-only registration, database-backed refresh tokens, and role-based access control. API key authentication for agents remains unchanged.

## Design Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Authentication Scope | Dashboard users only | Agents use existing API key infrastructure |
| Registration Flow | Invite-only | Controlled access for enterprise team collaboration |
| Refresh Token Storage | Database-backed | Enables session management and instant revocation |
| Password Security | 12+ chars, complexity, bcrypt cost 12 | NIST guidance, enterprise-grade security |
| Token Expiration | Access: 15min, Refresh: 7d, Invite: 48h | Balance security and UX |

## Architecture

### Core Components

1. **Auth Service** - Password hashing (bcrypt), JWT generation/validation, token refresh logic
2. **User Service** - User CRUD, invite creation, password validation
3. **Auth Middleware** - JWT validation, user context attachment, role-based authorization
4. **Database Layer** - `refresh_tokens` and `user_invites` tables
5. **Auth Handlers** - HTTP endpoints for login, refresh, logout, invite, accept-invite

### Authentication Flow

1. User logs in with email/password → receives access token (15 min) + refresh token (7 days)
2. Access token stored in memory, refresh token in httpOnly cookie
3. When access token expires, client calls `/auth/refresh` with cookie → receives new access token
4. Logout deletes refresh token from database and clears cookie

### Authorization Flow

1. Middleware extracts JWT from Authorization header
2. Validates signature and expiration
3. Loads user from database, checks `is_active` and role
4. Attaches user to request context for handlers to check permissions

## Database Schema

### refresh_tokens

```sql
CREATE TABLE refresh_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash VARCHAR(64) NOT NULL UNIQUE,     -- SHA256 of actual token
  device_name VARCHAR(255),                    -- e.g., "Chrome on MacBook Pro"
  user_agent TEXT,                             -- full UA string
  client_ip INET,                              -- IP address at creation
  expires_at TIMESTAMPTZ NOT NULL,             -- 7 days from creation
  revoked_at TIMESTAMPTZ,                      -- NULL = active, set on logout
  last_used_at TIMESTAMPTZ,                    -- updated on each refresh
  created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at);
```

### user_invites

```sql
CREATE TABLE user_invites (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email VARCHAR(255) NOT NULL,
  token_hash VARCHAR(64) NOT NULL UNIQUE,      -- SHA256 of invite token
  role VARCHAR(20) NOT NULL,                   -- admin, manager, viewer
  invited_by UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  accepted_at TIMESTAMPTZ,                     -- NULL = pending
  expires_at TIMESTAMPTZ NOT NULL,             -- 48 hours from creation
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE(email, accepted_at)                   -- prevent duplicate pending invites
);
CREATE INDEX idx_user_invites_token_hash ON user_invites(token_hash);
CREATE INDEX idx_user_invites_email ON user_invites(email);
CREATE INDEX idx_user_invites_expires_at ON user_invites(expires_at);
```

**Key Design Decisions:**
- Store SHA256 hashes, not plaintext tokens
- Track device info for "active sessions" feature
- Unique constraint on `(email, accepted_at)` prevents multiple pending invites
- Cascade deletes ensure cleanup when users are removed

## API Endpoints

### Authentication Endpoints (`/api/v1/auth`)

#### POST /auth/login
- **Request:** `{"email": "user@example.com", "password": "SecurePass123!"}`
- **Response:** `{"access_token": "eyJ...", "token_type": "Bearer", "expires_in": 900}`
- **Cookie:** `refresh_token=<token>; Secure; HttpOnly; SameSite=Strict; Max-Age=604800`
- **Errors:** 401 if credentials invalid or user inactive

#### POST /auth/refresh
- **Request:** Empty body, reads `refresh_token` from cookie
- **Response:** `{"access_token": "eyJ...", "token_type": "Bearer", "expires_in": 900}`
- **Side Effect:** Updates `last_used_at` on refresh token record
- **Errors:** 401 if token expired, revoked, or invalid

#### POST /auth/logout
- **Request:** Empty body, reads `refresh_token` from cookie
- **Response:** `{"message": "Logged out successfully"}`
- **Side Effect:** Sets `revoked_at` on refresh token, clears cookie
- **Always:** Returns 200 (even if token already invalid)

#### GET /auth/sessions (requires auth)
- **Response:** Array of active sessions for current user
- **Format:** `[{"id": "uuid", "device_name": "Chrome on MacBook", "created_at": "...", "last_used_at": "..."}]`

#### DELETE /auth/sessions/:id (requires auth)
- **Action:** Revokes specific refresh token (for "sign out other devices")
- **Response:** 204 on success, 404 if not found or not owned by user

### User Management Endpoints (`/api/v1/users`)

#### POST /users/invite (requires admin role)
- **Request:** `{"email": "newuser@example.com", "role": "viewer"}`
- **Response:** `{"invite_url": "https://app.example.com/accept-invite?token=abc123"}`
- **Side Effect:** Creates invite record (email sending out of scope)

#### POST /users/accept-invite
- **Request:** `{"token": "abc123", "password": "SecurePass123!", "display_name": "John Doe"}`
- **Response:** `{"message": "Account created successfully"}`
- **Side Effect:** Creates user, marks invite as accepted
- **Errors:** 400 if token expired/invalid or password doesn't meet requirements

## Authentication Middleware

### RequireAuth Middleware

Validates JWT access tokens and attaches user to request context:

```go
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract "Bearer <token>" from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
            writeJSON(w, 401, map[string]string{"error": "Missing or invalid authorization header"})
            return
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")

        // Validate JWT signature and expiration
        claims, err := authService.ValidateAccessToken(token)
        if err != nil {
            writeJSON(w, 401, map[string]string{"error": "Invalid or expired token"})
            return
        }

        // Load user from database and verify still active
        user, err := userService.GetByID(claims.UserID)
        if err != nil || !user.IsActive {
            writeJSON(w, 401, map[string]string{"error": "User not found or inactive"})
            return
        }

        // Attach user to context for downstream handlers
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### RequireRole Middleware

Checks user has minimum required role (chains after RequireAuth):

```go
func RequireRole(minRole string) func(http.Handler) http.Handler {
    roleHierarchy := map[string]int{"viewer": 1, "manager": 2, "admin": 3}

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user").(*User)

            if roleHierarchy[user.Role] < roleHierarchy[minRole] {
                writeJSON(w, 403, map[string]string{"error": "Insufficient permissions"})
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

### Usage Example

```go
// Public endpoint - no auth
http.Handle("/health", healthHandler)

// Protected endpoint - any authenticated user
http.Handle("/api/v1/auth/sessions", RequireAuth(sessionsHandler))

// Admin-only endpoint
http.Handle("/api/v1/users/invite", RequireAuth(RequireRole("admin")(inviteHandler)))
```

## Token Management

### JWT Access Token Structure

**Header:**
```json
{"alg": "HS256", "typ": "JWT"}
```

**Payload (Claims):**
```json
{
  "sub": "user-uuid",           // Subject: user ID
  "email": "user@example.com",  // User email
  "role": "admin",              // User role for quick auth checks
  "iat": 1706572800,            // Issued at (Unix timestamp)
  "exp": 1706573700             // Expires at (iat + 15 minutes)
}
```

**Signature:** HMAC-SHA256(base64(header) + "." + base64(payload), SECRET_KEY)

### Token Generation

```go
type AuthService struct {
    secretKey []byte  // Load from env var JWT_SECRET (min 32 bytes)
}

func (s *AuthService) GenerateAccessToken(user *User) (string, error) {
    claims := jwt.MapClaims{
        "sub":   user.ID.String(),
        "email": user.Email,
        "role":  user.Role,
        "iat":   time.Now().Unix(),
        "exp":   time.Now().Add(15 * time.Minute).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.secretKey)
}

func (s *AuthService) GenerateRefreshToken() (string, error) {
    // 32-byte cryptographically secure random token
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

func (s *AuthService) HashToken(token string) string {
    hash := sha256.Sum256([]byte(token))
    return hex.EncodeToString(hash[:])
}
```

### Token Validation

```go
func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        // Verify signing method
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return s.secretKey, nil
    })

    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }

    claims := token.Claims.(jwt.MapClaims)
    return &Claims{
        UserID: claims["sub"].(string),
        Email:  claims["email"].(string),
        Role:   claims["role"].(string),
    }, nil
}
```

**Security Notes:**
- JWT_SECRET must be at least 256 bits (32 bytes) for HS256
- Refresh/invite tokens are random, not JWTs (can't be decoded without database)
- Always hash tokens before storing (even in logs)

## Security Considerations

### 1. Rate Limiting

Prevent brute force attacks:

```go
// In-memory rate limiter (or use Redis in production)
var loginAttempts = make(map[string][]time.Time)  // email -> attempt timestamps

func checkRateLimit(email string, maxAttempts int, window time.Duration) error {
    now := time.Now()
    attempts := loginAttempts[email]

    // Filter to attempts within window
    var recent []time.Time
    for _, t := range attempts {
        if now.Sub(t) < window {
            recent = append(recent, t)
        }
    }

    if len(recent) >= maxAttempts {
        return fmt.Errorf("too many attempts, try again in %v", window)
    }

    loginAttempts[email] = append(recent, now)
    return nil
}
```

**Limits:**
- Login: 5 attempts per email per 15 minutes
- Invite acceptance: 3 attempts per token per 10 minutes
- Refresh: 10 requests per refresh token per minute

### 2. Audit Logging

Log all authentication events to existing `audit_log` table:

**Events to log:**
- `auth.login.success` / `auth.login.failure` - with client IP, user agent
- `auth.refresh.success` / `auth.refresh.failure`
- `auth.logout` - include which session was revoked
- `user.invite.created` - who invited whom, what role
- `user.invite.accepted` - new user registration
- `auth.session.revoked` - manual session termination

### 3. Password Validation

Enforce complexity at registration/password-change:

```go
func validatePassword(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters")
    }

    var hasUpper, hasLower, hasNumber, hasSpecial bool
    for _, char := range password {
        switch {
        case unicode.IsUpper(char): hasUpper = true
        case unicode.IsLower(char): hasLower = true
        case unicode.IsDigit(char): hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char): hasSpecial = true
        }
    }

    if !(hasUpper && hasLower && hasNumber && hasSpecial) {
        return errors.New("password must contain uppercase, lowercase, number, and special character")
    }

    return nil
}
```

### 4. Additional Hardening

- **CORS:** Restrict origins to dashboard domain only
- **Secure cookies:** Always set `Secure`, `HttpOnly`, `SameSite=Strict` on refresh tokens
- **Token rotation:** Issue new refresh token on each refresh (optional - can add later)
- **Account lockout:** After 10 failed login attempts in 1 hour, set `is_active=false` (requires admin unlock)
- **Cleanup job:** Cron task to delete expired refresh tokens and invites older than 30 days

## Testing Strategy

### 1. Unit Tests

**Auth Service Tests (`auth_service_test.go`):**
- `TestGenerateAccessToken` - verify JWT structure, claims, expiration
- `TestValidateAccessToken` - valid tokens pass, expired/invalid fail
- `TestHashPassword` - bcrypt hashing works, same password produces different hashes
- `TestVerifyPassword` - correct passwords verify, wrong ones don't
- `TestGenerateRefreshToken` - produces unique 32-byte tokens
- `TestHashToken` - consistent SHA256 hashing

**User Service Tests (`user_service_test.go`):**
- `TestCreateInvite` - creates valid invite with correct expiration
- `TestAcceptInvite` - creates user, marks invite accepted, rejects expired invites
- `TestValidatePassword` - enforces all complexity rules
- `TestGetByEmail` - finds existing users, returns error for missing users

### 2. Integration Tests

**Login Flow (`auth_integration_test.go`):**
```go
func TestLoginFlow(t *testing.T) {
    // Setup: create test user in DB
    user := createTestUser(t, "test@example.com", "ValidPass123!")

    // Login with correct credentials
    resp := login(t, "test@example.com", "ValidPass123!")
    assert.Equal(t, 200, resp.StatusCode)
    assert.NotEmpty(t, resp.Body.AccessToken)
    assert.NotEmpty(t, resp.Cookies["refresh_token"])

    // Verify refresh token stored in DB
    rt := getRefreshToken(t, user.ID)
    assert.False(t, rt.ExpiresAt.Before(time.Now()))

    // Login with wrong password
    resp = login(t, "test@example.com", "WrongPass123!")
    assert.Equal(t, 401, resp.StatusCode)
}
```

**Invite Flow (`invite_integration_test.go`):**
- Create invite as admin → verify invite record created
- Accept invite with valid token → verify user created, invite marked accepted
- Try to accept expired invite → verify 400 error
- Try to create duplicate pending invite → verify unique constraint prevents it

**Refresh Flow (`refresh_integration_test.go`):**
- Use refresh token → get new access token, verify `last_used_at` updated
- Use expired refresh token → verify 401 error
- Use revoked refresh token → verify 401 error

### 3. API Tests

**Middleware Tests (`middleware_test.go`):**
- Request without Authorization header → 401
- Request with invalid JWT → 401
- Request with valid JWT for inactive user → 401
- Request with valid JWT for active user → passes, user in context
- `RequireRole("admin")` with viewer user → 403

### 4. Test Coverage Goals

- **Unit tests:** 90%+ coverage on auth/user services
- **Integration tests:** Cover all happy paths + key error cases
- **API tests:** One test per endpoint (success + main failure modes)

### 5. Test Data Fixtures

- `testAdmin` - active admin user
- `testViewer` - active viewer user
- `testInactive` - inactive user (for testing lockout)
- `testInvite` - valid unexpired invite
- `testExpiredInvite` - expired invite

## Implementation Checklist

- [ ] Create database migrations for `refresh_tokens` and `user_invites` tables
- [ ] Implement Auth Service (JWT generation/validation, password hashing)
- [ ] Implement User Service (invite creation, password validation)
- [ ] Implement middleware (RequireAuth, RequireRole)
- [ ] Create HTTP handlers for auth endpoints
- [ ] Create HTTP handlers for user invite endpoints
- [ ] Add rate limiting for login/invite endpoints
- [ ] Add audit logging for auth events
- [ ] Write unit tests (auth service, user service)
- [ ] Write integration tests (login, refresh, invite flows)
- [ ] Write API tests (middleware, endpoints)
- [ ] Add cleanup cron job for expired tokens
- [ ] Document API endpoints in API reference
- [ ] Update README with authentication setup instructions

## Dependencies

- JWT library: `github.com/golang-jwt/jwt/v5`
- Bcrypt: `golang.org/x/crypto/bcrypt` (in Go standard library)
- Database migrations: Use existing migration tool (from Issue #8)

## Out of Scope (Future Work)

- Email sending for invite notifications (Issue #9 focuses on auth only)
- Password reset flow (add in future issue)
- Two-factor authentication (add in future issue)
- OAuth/SSO integration (add in future issue)
- Token rotation on refresh (optional enhancement)
