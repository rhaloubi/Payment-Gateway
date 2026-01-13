package jwt

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/config"
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

// Use this in all microservices
func (v *JWTValidator) AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := v.ValidateToken(authHeader)
		if err != nil {
			c.JSON(401, gin.H{
				"success": false,
				"error":   "Invalid or expired token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Set user info in context for handlers to use
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)

		c.Next()
	}
}

// GetUserIDFromContext extracts user ID from context
func (v *JWTValidator) GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, errors.New("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return uuid.Nil, errors.New("user_id is not a string")
	}

	return uuid.Parse(userIDStr)
}

// GetUserEmailFromContext extracts user email from context
func (v *JWTValidator) GetUserEmailFromContext(c *gin.Context) (string, error) {
	email, exists := c.Get("user_email")
	if !exists {
		return "", errors.New("user_email not found in context")
	}

	emailStr, ok := email.(string)
	if !ok {
		return "", errors.New("user_email is not a string")
	}

	return emailStr, nil
}
