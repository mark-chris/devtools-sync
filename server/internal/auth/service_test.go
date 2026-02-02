package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// RED: Test JWT access token generation
func TestGenerateAccessToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	user := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "admin",
	}

	// Act
	tokenString, err := service.GenerateAccessToken(user)

	// Assert
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v, want nil", err)
	}

	if tokenString == "" {
		t.Fatal("GenerateAccessToken() returned empty token")
	}

	// Verify token structure and claims
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		t.Fatalf("jwt.Parse() error = %v, want nil", err)
	}

	if !token.Valid {
		t.Fatal("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to parse claims")
	}

	// Verify claims
	if claims["sub"] != user.ID.String() {
		t.Errorf("claim 'sub' = %v, want %v", claims["sub"], user.ID.String())
	}

	if claims["email"] != user.Email {
		t.Errorf("claim 'email' = %v, want %v", claims["email"], user.Email)
	}

	if claims["role"] != user.Role {
		t.Errorf("claim 'role' = %v, want %v", claims["role"], user.Role)
	}

	// Verify expiration is ~15 minutes from now
	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatal("claim 'exp' is not a number")
	}

	expTime := time.Unix(int64(exp), 0)
	expectedExp := time.Now().Add(15 * time.Minute)

	// Allow 1 minute tolerance
	if expTime.Before(expectedExp.Add(-1*time.Minute)) || expTime.After(expectedExp.Add(1*time.Minute)) {
		t.Errorf("token expiration = %v, want within 1 min of %v", expTime, expectedExp)
	}

	// Verify issued at exists
	if _, ok := claims["iat"].(float64); !ok {
		t.Error("claim 'iat' is missing or not a number")
	}
}

// RED: Test JWT access token validation with valid token
func TestValidateAccessToken_ValidToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	user := &User{
		ID:    uuid.New(),
		Email: "test@example.com",
		Role:  "admin",
	}

	// Generate a valid token
	tokenString, err := service.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("setup failed: GenerateAccessToken() error = %v", err)
	}

	// Act
	claims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v, want nil", err)
	}

	if claims == nil {
		t.Fatal("ValidateAccessToken() returned nil claims")
	}

	if claims.UserID != user.ID.String() {
		t.Errorf("claims.UserID = %v, want %v", claims.UserID, user.ID.String())
	}

	if claims.Email != user.Email {
		t.Errorf("claims.Email = %v, want %v", claims.Email, user.Email)
	}

	if claims.Role != user.Role {
		t.Errorf("claims.Role = %v, want %v", claims.Role, user.Role)
	}
}

// RED: Test JWT access token validation with invalid token
func TestValidateAccessToken_InvalidToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	invalidToken := "invalid.jwt.token"

	// Act
	claims, err := service.ValidateAccessToken(invalidToken)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error")
	}

	if claims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", claims)
	}
}

// RED: Test JWT access token validation with expired token
func TestValidateAccessToken_ExpiredToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create an expired token manually
	claims := jwt.MapClaims{
		"sub":   uuid.New().String(),
		"email": "test@example.com",
		"role":  "admin",
		"iat":   time.Now().Add(-1 * time.Hour).Unix(),
		"exp":   time.Now().Add(-30 * time.Minute).Unix(), // Expired 30 min ago
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(expiredToken)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for expired token")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test password hashing
func TestHashPassword(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)
	password := "SecurePass123!"

	// Act
	hash, err := service.HashPassword(password)

	// Assert
	if err != nil {
		t.Fatalf("HashPassword() error = %v, want nil", err)
	}

	if hash == "" {
		t.Fatal("HashPassword() returned empty hash")
	}

	if hash == password {
		t.Error("HashPassword() returned plaintext password, want hash")
	}

	// Verify hash starts with bcrypt prefix
	if len(hash) < 60 {
		t.Errorf("HashPassword() hash length = %d, want >= 60 (bcrypt hash)", len(hash))
	}
}

// RED: Test same password produces different hashes (salted)
func TestHashPassword_DifferentSalts(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)
	password := "SecurePass123!"

	// Act
	hash1, err1 := service.HashPassword(password)
	hash2, err2 := service.HashPassword(password)

	// Assert
	if err1 != nil || err2 != nil {
		t.Fatalf("HashPassword() errors = %v, %v, want nil, nil", err1, err2)
	}

	if hash1 == hash2 {
		t.Error("HashPassword() produced same hash for same password, want different hashes (salted)")
	}
}

// RED: Test password verification with correct password
func TestVerifyPassword_CorrectPassword(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)
	password := "SecurePass123!"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("setup failed: HashPassword() error = %v", err)
	}

	// Act
	err = service.VerifyPassword(hash, password)

	// Assert
	if err != nil {
		t.Errorf("VerifyPassword() error = %v, want nil", err)
	}
}

// RED: Test password verification with wrong password
func TestVerifyPassword_WrongPassword(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)
	password := "SecurePass123!"
	wrongPassword := "WrongPass456!"

	hash, err := service.HashPassword(password)
	if err != nil {
		t.Fatalf("setup failed: HashPassword() error = %v", err)
	}

	// Act
	err = service.VerifyPassword(hash, wrongPassword)

	// Assert
	if err == nil {
		t.Error("VerifyPassword() error = nil, want error for wrong password")
	}
}

// RED: Test refresh token generation
func TestGenerateRefreshToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Act
	token, err := service.GenerateRefreshToken()

	// Assert
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v, want nil", err)
	}

	if token == "" {
		t.Fatal("GenerateRefreshToken() returned empty token")
	}

	// Verify token is base64 URL encoded (no special chars except -_)
	if len(token) < 32 {
		t.Errorf("GenerateRefreshToken() token length = %d, want >= 32", len(token))
	}
}

// RED: Test refresh tokens are unique
func TestGenerateRefreshToken_Unique(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Act
	token1, err1 := service.GenerateRefreshToken()
	token2, err2 := service.GenerateRefreshToken()

	// Assert
	if err1 != nil || err2 != nil {
		t.Fatalf("GenerateRefreshToken() errors = %v, %v, want nil, nil", err1, err2)
	}

	if token1 == token2 {
		t.Error("GenerateRefreshToken() produced same token twice, want unique tokens")
	}
}

// RED: Test token hashing
func TestHashToken(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)
	token := "test-token-123"

	// Act
	hash := service.HashToken(token)

	// Assert
	if hash == "" {
		t.Fatal("HashToken() returned empty hash")
	}

	if hash == token {
		t.Error("HashToken() returned plaintext token, want hash")
	}

	// SHA256 hash is 64 hex characters
	if len(hash) != 64 {
		t.Errorf("HashToken() hash length = %d, want 64 (SHA256 hex)", len(hash))
	}
}

// RED: Test token hashing is deterministic
func TestHashToken_Deterministic(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)
	token := "test-token-123"

	// Act
	hash1 := service.HashToken(token)
	hash2 := service.HashToken(token)

	// Assert
	if hash1 != hash2 {
		t.Error("HashToken() produced different hashes for same token, want deterministic")
	}
}

// RED: Test JWT validation with missing "sub" claim
func TestValidateAccessToken_MissingSubClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token without "sub" claim
	claims := jwt.MapClaims{
		"email": "test@example.com",
		"role":  "admin",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for missing 'sub' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with missing "email" claim
func TestValidateAccessToken_MissingEmailClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token without "email" claim
	claims := jwt.MapClaims{
		"sub":  uuid.New().String(),
		"role": "admin",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for missing 'email' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with missing "role" claim
func TestValidateAccessToken_MissingRoleClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token without "role" claim
	claims := jwt.MapClaims{
		"sub":   uuid.New().String(),
		"email": "test@example.com",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for missing 'role' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with wrong type for "sub" claim
func TestValidateAccessToken_WrongTypeSubClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token with numeric "sub" instead of string
	claims := jwt.MapClaims{
		"sub":   12345, // Wrong type - should be string
		"email": "test@example.com",
		"role":  "admin",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for wrong type 'sub' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with wrong type for "email" claim
func TestValidateAccessToken_WrongTypeEmailClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token with boolean "email" instead of string
	claims := jwt.MapClaims{
		"sub":   uuid.New().String(),
		"email": true, // Wrong type - should be string
		"role":  "admin",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for wrong type 'email' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with wrong type for "role" claim
func TestValidateAccessToken_WrongTypeRoleClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token with array "role" instead of string
	claims := jwt.MapClaims{
		"sub":   uuid.New().String(),
		"email": "test@example.com",
		"role":  []string{"admin", "user"}, // Wrong type - should be string
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for wrong type 'role' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with empty "sub" claim
func TestValidateAccessToken_EmptySubClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token with empty "sub"
	claims := jwt.MapClaims{
		"sub":   "", // Empty string
		"email": "test@example.com",
		"role":  "admin",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for empty 'sub' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with empty "email" claim
func TestValidateAccessToken_EmptyEmailClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token with empty "email"
	claims := jwt.MapClaims{
		"sub":   uuid.New().String(),
		"email": "", // Empty string
		"role":  "admin",
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for empty 'email' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}

// RED: Test JWT validation with empty "role" claim
func TestValidateAccessToken_EmptyRoleClaim(t *testing.T) {
	// Setup
	secretKey := []byte("test-secret-key-min-32-bytes-long!")
	service := NewAuthService(secretKey)

	// Create token with empty "role"
	claims := jwt.MapClaims{
		"sub":   uuid.New().String(),
		"email": "test@example.com",
		"role":  "", // Empty string
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		t.Fatalf("setup failed: token.SignedString() error = %v", err)
	}

	// Act
	parsedClaims, err := service.ValidateAccessToken(tokenString)

	// Assert
	if err == nil {
		t.Fatal("ValidateAccessToken() error = nil, want error for empty 'role' claim")
	}

	if parsedClaims != nil {
		t.Errorf("ValidateAccessToken() claims = %v, want nil", parsedClaims)
	}
}
