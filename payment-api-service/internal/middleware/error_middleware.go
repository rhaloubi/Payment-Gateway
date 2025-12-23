package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"go.uber.org/zap"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")

				logger.Log.Error("Panic recovered",
					zap.String("request_id", requestID.(string)),
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "an internal error occurred",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key, Idempotency-Key , X-Client-Secret")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
