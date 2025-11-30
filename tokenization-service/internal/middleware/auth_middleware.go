package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/client"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/jwt"

	"go.uber.org/zap"
)

// AuthMiddleware validates JWT tokens or API keys
func AuthMiddleware() gin.HandlerFunc {
	authClient := client.NewAuthServiceClient()
	authJWT := jwt.NewJWTValidator()

	return func(c *gin.Context) {
		// Try JWT first (Authorization: Bearer xxx)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate JWT
			jwtData, err := authJWT.ValidateToken(token)
			if err != nil {
				logger.Log.Warn("JWT validation failed",
					zap.Error(err),
					zap.String("ip", c.ClientIP()),
				)
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "invalid or expired token",
				})
				c.Abort()
				return
			}

			// Set user context
			c.Set("user_id", jwtData.UserID)
			c.Set("email", jwtData.Email)
			c.Set("auth_type", "jwt")

			logger.Log.Debug("JWT authentication successful",
				zap.String("user_id", jwtData.UserID),
			)

			c.Next()
			return
		}

		// Try API Key (X-API-Key: pk_xxx)
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			// Validate API key format
			if !strings.HasPrefix(apiKey, "pk_") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "invalid API key format",
				})
				c.Abort()
				return
			}

			// Validate API key
			apiKeyData, err := authClient.ValidateAPIKey(apiKey)
			if err != nil {
				logger.Log.Warn("API key validation failed",
					zap.Error(err),
					zap.String("ip", c.ClientIP()),
				)
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "invalid API key",
				})
				c.Abort()
				return
			}

			// Set merchant context
			c.Set("merchant_id", apiKeyData.MerchantID.String())
			c.Set("api_key_id", apiKeyData.KeyID.String())
			c.Set("api_key_name", apiKeyData.Name)
			c.Set("auth_type", "api_key")

			logger.Log.Debug("API key authentication successful",
				zap.String("merchant_id", apiKeyData.MerchantID.String()),
				zap.String("key_name", apiKeyData.Name),
			)

			c.Next()
			return
		}

		// No authentication provided
		logger.Log.Warn("No authentication provided",
			zap.String("ip", c.ClientIP()),
			zap.String("path", c.Request.URL.Path),
		)

		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "authentication required (Bearer token or X-API-Key header)",
		})
		c.Abort()
	}
}
