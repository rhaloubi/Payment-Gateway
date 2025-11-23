package middleware

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"go.uber.org/zap"
)

// InternalServiceMiddleware authenticates internal service-to-service calls
func InternalServiceMiddleware() gin.HandlerFunc {
	// Get internal service secret from environment
	internalSecret := os.Getenv("INTERNAL_SERVICE_SECRET")
	if internalSecret == "" {
		logger.Log.Warn("INTERNAL_SERVICE_SECRET not set - internal endpoints are INSECURE")
		internalSecret = "dev-internal-secret-change-in-production"
	}

	return func(c *gin.Context) {
		// Check for internal service header
		serviceName := c.GetHeader("X-Internal-Service")
		serviceSecret := c.GetHeader("X-Internal-Secret")

		if serviceName == "" || serviceSecret == "" {
			logger.Log.Warn("Internal endpoint accessed without proper headers",
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
			)

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "internal service authentication required",
			})
			c.Abort()
			return
		}

		// Validate secret
		if serviceSecret != internalSecret {
			logger.Log.Warn("Invalid internal service secret",
				zap.String("service", serviceName),
				zap.String("ip", c.ClientIP()),
			)

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid internal service credentials",
			})
			c.Abort()
			return
		}

		// Whitelist allowed services
		allowedServices := []string{
			"transaction-service",
			"payment-api-service",
			"fraud-detection-service",
		}

		allowed := false
		for _, s := range allowedServices {
			if s == serviceName {
				allowed = true
				break
			}
		}

		if !allowed {
			logger.Log.Warn("Unknown internal service attempted access",
				zap.String("service", serviceName),
				zap.String("ip", c.ClientIP()),
			)

			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "service not authorized",
			})
			c.Abort()
			return
		}

		// Set service context
		c.Set("internal_service", serviceName)
		c.Set("auth_type", "internal")

		logger.Log.Debug("Internal service authenticated",
			zap.String("service", serviceName),
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}
