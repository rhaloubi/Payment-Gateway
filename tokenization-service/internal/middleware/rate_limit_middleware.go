package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"go.uber.org/zap"
)

type RateLimitConfig struct {
	RequestsPerSecond int
	RequestsPerHour   int
	BurstSize         int
}

var defaultRateLimit = RateLimitConfig{
	RequestsPerSecond: 50,
	RequestsPerHour:   5000,
	BurstSize:         10,
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantIDStr, exists := c.Get("merchant_id")
		if !exists {
			merchantIDStr = c.ClientIP()
		}

		merchantID := merchantIDStr.(string)

		// Check per-second rate limit
		allowedSecond, err := checkRateLimit(
			merchantID,
			"second",
			defaultRateLimit.RequestsPerSecond,
			1*time.Second,
		)

		if err != nil {
			logger.Log.Error("Rate limit check failed", zap.Error(err))
			// Allow request on error (fail open)
			c.Next()
			return
		}

		if !allowedSecond {
			logger.Log.Warn("Rate limit exceeded (per second)",
				zap.String("merchant_id", merchantID),
				zap.String("ip", c.ClientIP()),
			)

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", defaultRateLimit.RequestsPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "1")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded: too many requests per second",
			})
			c.Abort()
			return
		}

		// Check per-hour rate limit
		allowedHour, err := checkRateLimit(
			merchantID,
			"hour",
			defaultRateLimit.RequestsPerHour,
			1*time.Hour,
		)

		if err != nil {
			logger.Log.Error("Rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if !allowedHour {
			logger.Log.Warn("Rate limit exceeded (per hour)",
				zap.String("merchant_id", merchantID),
				zap.String("ip", c.ClientIP()),
			)

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", defaultRateLimit.RequestsPerHour))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "3600")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded: hourly limit reached",
			})
			c.Abort()
			return
		}

		// Get remaining requests
		remaining, _ := getRemainingRequests(merchantID, "second", defaultRateLimit.RequestsPerSecond)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", defaultRateLimit.RequestsPerSecond))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// checkRateLimit checks if request is within rate limit using Redis
func checkRateLimit(key string, window string, limit int, ttl time.Duration) (bool, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:tokenization:%s:%s", key, window)

	// Increment counter
	count, err := inits.RDB.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	// Set expiry on first request
	if count == 1 {
		inits.RDB.Expire(ctx, redisKey, ttl)
	}

	// Check if within limit
	return count <= int64(limit), nil
}

// getRemainingRequests gets remaining requests in current window
func getRemainingRequests(key string, window string, limit int) (int, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:tokenization:%s:%s", key, window)

	count, err := inits.RDB.Get(ctx, redisKey).Int()
	if err != nil {
		// Key doesn't exist, all requests available
		return limit, nil
	}

	remaining := limit - count
	if remaining < 0 {
		return 0, nil
	}

	return remaining, nil
}

// CustomRateLimitMiddleware allows custom rate limits per endpoint
func CustomRateLimitMiddleware(requestsPerSecond int, requestsPerHour int) gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantIDStr, exists := c.Get("merchant_id")
		if !exists {
			merchantIDStr = c.ClientIP()
		}

		merchantID := merchantIDStr.(string)

		// Check per-second limit
		allowedSecond, _ := checkRateLimit(
			merchantID,
			"second",
			requestsPerSecond,
			1*time.Second,
		)

		if !allowedSecond {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Check per-hour limit
		allowedHour, _ := checkRateLimit(
			merchantID,
			"hour",
			requestsPerHour,
			1*time.Hour,
		)

		if !allowedHour {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "hourly rate limit exceeded",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
