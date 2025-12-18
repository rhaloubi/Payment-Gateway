package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/api-gateway/internal/config"
	"github.com/rhaloubi/api-gateway/internal/service"
)

func AuthenticateJWT(authClient *service.AuthClient, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing authorization header",
			})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := authClient.ValidateJWT(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims["user_id"])
		c.Set("merchant_id", claims["merchant_id"])
		c.Set("email", claims["email"])
		c.Set("roles", claims["roles"])

		if userID, ok := claims["user_id"].(string); ok {
			c.Request.Header.Set("X-User-ID", userID)
		}
		if merchantID, ok := claims["merchant_id"].(string); ok {
			c.Request.Header.Set("X-Merchant-ID", merchantID)
		}

		c.Next()
	}
}

func AuthenticateAPIKey(authClient *service.AuthClient, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing API key",
			})
			c.Abort()
			return
		}

		if !strings.HasPrefix(apiKey, "pk_live_") && !strings.HasPrefix(apiKey, "pk_test_") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid API key format",
			})
			c.Abort()
			return
		}

		keyInfo, err := authClient.ValidateAPIKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or inactive API key",
			})
			c.Abort()
			return
		}

		c.Set("api_key_id", keyInfo["id"])
		c.Set("merchant_id", keyInfo["merchant_id"])
		c.Set("key_name", keyInfo["name"])

		if merchantID, ok := keyInfo["merchant_id"].(string); ok {
			c.Request.Header.Set("X-Merchant-ID", merchantID)
		}
		if keyID, ok := keyInfo["id"].(string); ok {
			c.Request.Header.Set("X-API-Key-ID", keyID)
		}

		c.Next()
	}
}
