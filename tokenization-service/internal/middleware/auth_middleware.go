package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/client"

	"go.uber.org/zap"
)

// AuthMiddleware validates JWT tokens or API keys
func AuthMiddleware() gin.HandlerFunc {
	authClient := client.NewAuthClient()

	return func(c *gin.Context) {
		// Try JWT first (Authorization: Bearer xxx)
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate JWT
			jwtData, err := authClient.ValidateJWT(token)
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
			c.Set("user_id", jwtData.UserID.String())
			c.Set("merchant_id", jwtData.MerchantID.String())
			c.Set("email", jwtData.Email)
			c.Set("roles", jwtData.Roles)
			c.Set("auth_type", "jwt")

			logger.Log.Debug("JWT authentication successful",
				zap.String("user_id", jwtData.UserID.String()),
				zap.String("merchant_id", jwtData.MerchantID.String()),
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
			c.Set("permissions", apiKeyData.Permissions)
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

// RequirePermission middleware checks if user has specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	authClient := client.NewAuthClient()

	return func(c *gin.Context) {
		// Get user context
		userIDStr, userExists := c.Get("user_id")
		merchantIDStr, merchantExists := c.Get("merchant_id")

		if !merchantExists {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "merchant context required",
			})
			c.Abort()
			return
		}

		// If API key auth, check permissions array
		authType, _ := c.Get("auth_type")
		if authType == "api_key" {
			permissions, exists := c.Get("permissions")
			if exists {
				permList, ok := permissions.([]string)
				if ok {
					for _, p := range permList {
						if p == permission {
							c.Next()
							return
						}
					}
				}
			}

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "insufficient permissions",
			})
			c.Abort()
			return
		}

		// If JWT auth, check via auth service
		if userExists {
			userID := parseUUID(userIDStr.(string))
			merchantID := parseUUID(merchantIDStr.(string))

			hasPermission, err := authClient.CheckPermission(userID, merchantID, permission)
			if err != nil {
				logger.Log.Error("Permission check failed", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "failed to verify permissions",
				})
				c.Abort()
				return
			}

			if !hasPermission {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "insufficient permissions",
				})
				c.Abort()
				return
			}

			c.Next()
			return
		}

		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "authorization failed",
		})
		c.Abort()
	}
}

// Helper function
func parseUUID(s string) uuid.UUID {
	id, _ := uuid.Parse(s)
	return id
}
