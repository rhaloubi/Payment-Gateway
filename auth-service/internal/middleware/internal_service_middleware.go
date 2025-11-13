package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func InternalServiceMiddleware() gin.HandlerFunc {
	// Get shared secret from environment (for development)
	internalSecret := os.Getenv("INTERNAL_SERVICE_SECRET")
	if internalSecret == "" {
		// For development, accept any internal request
		// In production, this should FAIL if secret is not set
		return func(c *gin.Context) {
			serviceName := c.GetHeader("X-Internal-Service")
			if serviceName == "" {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "Internal service header required",
				})
				c.Abort()
				return
			}
			c.Next()
		}
	}

	return func(c *gin.Context) {
		// Validate service header
		serviceName := c.GetHeader("X-Internal-Service")
		if serviceName == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Internal service header required",
			})
			c.Abort()
			return
		}

		// Validate shared secret
		secret := c.GetHeader("X-Internal-Secret")
		if secret != internalSecret {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Invalid internal service credentials",
			})
			c.Abort()
			return
		}

		// Set service name in context for logging
		c.Set("internal_service", serviceName)
		c.Next()
	}
}

func GetInternalServiceName(c *gin.Context) string {
	service, exists := c.Get("internal_service")
	if !exists {
		return "unknown"
	}
	return service.(string)
}
