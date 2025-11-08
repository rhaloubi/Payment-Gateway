package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/service"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware() gin.HandlerFunc {
	authService := service.NewAuthService()

	return func(c *gin.Context) {
		// Get token from header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token (remove "Bearer " prefix)
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid authorization format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Validate token
		user, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Set("user_id", user.ID.String())

		c.Next()
	}
}

// APIKeyMiddleware validates API keys
func APIKeyMiddleware() gin.HandlerFunc {
	apiKeyService := service.NewAPIKeyService()

	return func(c *gin.Context) {
		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "X-API-Key header required",
			})
			c.Abort()
			return
		}

		// Validate API key
		key, err := apiKeyService.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or inactive API key",
			})
			c.Abort()
			return
		}

		// Set merchant info in context
		c.Set("merchant_id", key.MerchantID.String())
		c.Set("api_key_id", key.ID.String())

		c.Next()
	}
}

// RequirePermission checks if user has a specific permission
func RequirePermission(resource, action string) gin.HandlerFunc {
	roleService := service.NewRoleService()

	return func(c *gin.Context) {
		// Get user ID and merchant ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
			})
			c.Abort()
			return
		}

		merchantID, exists := c.Get("merchant_id")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "merchant_id required",
			})
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := roleService.HasPermission(
			userID.(uuid.UUID),
			merchantID.(uuid.UUID),
			resource,
			action,
		)

		if err != nil || !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
