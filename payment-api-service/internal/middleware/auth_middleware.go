package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/client"
	"go.uber.org/zap"
)

func AuthMiddleware() gin.HandlerFunc {
	authClient := client.NewAuthServiceClient()

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			logger.Log.Warn("No API key provided",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "API key required (X-API-Key header)",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(apiKey, "pk_") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid API key format",
			})
			c.Abort()
			return
		}

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

		c.Set("merchant_id", apiKeyData.MerchantID.String())
		c.Set("api_key_id", apiKeyData.KeyID.String())
		c.Set("api_key_name", apiKeyData.Name)
		c.Set("auth_type", "api_key")

		logger.Log.Debug("API key authentication successful",
			zap.String("merchant_id", apiKeyData.MerchantID.String()),
			zap.String("key_name", apiKeyData.Name),
		)

		c.Next()
	}
}
