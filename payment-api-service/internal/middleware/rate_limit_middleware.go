package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"go.uber.org/zap"
)

var paymentRateLimit = struct {
	RequestsPerSecond int
	RequestsPerHour   int
}{
	RequestsPerSecond: 20,
	RequestsPerHour:   10000,
}

func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		merchantIDStr, exists := c.Get("merchant_id")
		if !exists {
			merchantIDStr = c.ClientIP()
		}

		merchantID := merchantIDStr.(string)

		allowedSecond, _ := checkRateLimit(
			merchantID,
			"second",
			paymentRateLimit.RequestsPerSecond,
			1*time.Second,
		)

		if !allowedSecond {
			logger.Log.Warn("Rate limit exceeded (per second)",
				zap.String("merchant_id", merchantID),
				zap.String("ip", c.ClientIP()),
			)

			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", paymentRateLimit.RequestsPerSecond))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "1")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded: too many requests per second",
			})
			c.Abort()
			return
		}

		allowedHour, _ := checkRateLimit(
			merchantID,
			"hour",
			paymentRateLimit.RequestsPerHour,
			1*time.Hour,
		)

		if !allowedHour {
			logger.Log.Warn("Rate limit exceeded (per hour)",
				zap.String("merchant_id", merchantID),
			)

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded: hourly limit reached",
			})
			c.Abort()
			return
		}

		remaining, _ := getRemainingRequests(merchantID, "second", paymentRateLimit.RequestsPerSecond)
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", paymentRateLimit.RequestsPerSecond))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

func checkRateLimit(key string, window string, limit int, ttl time.Duration) (bool, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:payment:%s:%s", key, window)

	count, err := inits.RDB.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		inits.RDB.Expire(ctx, redisKey, ttl)
	}

	return count <= int64(limit), nil
}

func getRemainingRequests(key string, window string, limit int) (int, error) {
	ctx := context.Background()
	redisKey := fmt.Sprintf("rate_limit:payment:%s:%s", key, window)

	count, err := inits.RDB.Get(ctx, redisKey).Int()
	if err != nil {
		return limit, nil
	}

	remaining := limit - count
	if remaining < 0 {
		return 0, nil
	}
	return remaining, nil
}
