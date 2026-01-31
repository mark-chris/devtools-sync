package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	secretKey []byte
}

// NewAuthService creates a new AuthService
func NewAuthService(secretKey []byte) *AuthService {
	return &AuthService{
		secretKey: secretKey,
	}
}

// GenerateAccessToken generates a JWT access token for a user
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

// ValidateAccessToken validates a JWT access token and returns the claims
func (s *AuthService) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claimsMap, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return &Claims{
		UserID: claimsMap["sub"].(string),
		Email:  claimsMap["email"].(string),
		Role:   claimsMap["role"].(string),
	}, nil
}

// HashPassword hashes a password using bcrypt with cost factor 12
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks if a password matches the hash
func (s *AuthService) VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateRefreshToken generates a cryptographically secure random token
func (s *AuthService) GenerateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashToken hashes a token using SHA256
func (s *AuthService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
