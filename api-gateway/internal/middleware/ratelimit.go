package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/api-gateway/internal/config"
	"github.com/rhaloubi/api-gateway/internal/service"
)

func RateLimiter(limiter *service.RateLimiter, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		allowed, remaining, resetTime := limiter.Allow(clientIP, cfg.RateLimiting.Global.RequestsPerHour, time.Hour)

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", cfg.RateLimiting.Global.RequestsPerHour))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"error":       "rate limit exceeded",
				"retry_after": resetTime.Sub(time.Now()).Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func EndpointRateLimit(limiter *service.RateLimiter, endpoint string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := c.ClientIP()
		if apiKey := c.GetHeader("X-API-Key"); apiKey != "" {
			identifier = apiKey
		}

		key := fmt.Sprintf("%s:%s", endpoint, identifier)
		allowed, remaining, resetTime := limiter.Allow(key, limit, window)

		c.Header("X-RateLimit-Endpoint", endpoint)
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":     false,
				"error":       fmt.Sprintf("rate limit exceeded for %s", endpoint),
				"retry_after": resetTime.Sub(time.Now()).Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
