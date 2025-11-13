package jwt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTUtil handles JWT operations
type JWTUtil struct {
	secretKey string
}

// JWTClaims represents JWT claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// NewJWTUtil creates a new JWT utility
func NewJWTUtil() *JWTUtil {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		secretKey = "default-secret-key-change-in-production"
	}
	return &JWTUtil{
		secretKey: secretKey,
	}
}

// GenerateAccessToken generates a new access token
func (u *JWTUtil) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	claims := JWTClaims{
		UserID: userID.String(),
		Email:  email,
		Type:   "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "payment-gateway",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.secretKey))
}

// GenerateRefreshToken generates a new refresh token
func (u *JWTUtil) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	claims := JWTClaims{
		UserID: userID.String(),
		Type:   "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "payment-gateway",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(u.secretKey))
}

// ValidateAccessToken validates an access token
func (u *JWTUtil) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	return u.validateToken(tokenString, "access")
}

// ValidateRefreshToken validates a refresh token
func (u *JWTUtil) ValidateRefreshToken(tokenString string) (*JWTClaims, error) {
	return u.validateToken(tokenString, "refresh")
}

// validateToken validates a JWT token
func (u *JWTUtil) validateToken(tokenString, expectedType string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(u.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Type != expectedType {
		return nil, errors.New("invalid token type")
	}

	return claims, nil
}

// HashToken hashes a token using SHA-256 for storage
func (u *JWTUtil) HashToken(token string) string {
	return HashSHA256(token)
}

// HashSHA256 hashes a string using SHA-256
func HashSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}
