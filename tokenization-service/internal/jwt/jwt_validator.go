package jwt

import (
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rhaloubi/payment-gateway/tokenization-service/config"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Type   string `json:"type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

type JWTValidator struct {
	secretKey string
}

func NewJWTValidator() *JWTValidator {
	secretKey := config.GetEnv("JWT_SECRET_KEY")
	if secretKey == "" {
		panic("JWT_SECRET_KEY environment variable is required")
	}
	return &JWTValidator{
		secretKey: secretKey,
	}
}

func (v *JWTValidator) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(v.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	if claims.Type != "access" {
		return nil, errors.New("invalid token type, expected access token")
	}

	return claims, nil
}
