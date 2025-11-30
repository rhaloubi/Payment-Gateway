package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"go.uber.org/zap"
)

const idempotencyTTL = 24 * time.Hour

func IdempotencyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		idempotencyKey := c.GetHeader("Idempotency-Key")
		if idempotencyKey == "" {
			c.Next()
			return
		}

		if len(idempotencyKey) < 16 || len(idempotencyKey) > 255 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "idempotency key must be between 16 and 255 characters",
			})
			c.Abort()
			return
		}

		merchantID, exists := c.Get("merchant_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "authentication required for idempotency",
			})
			c.Abort()
			return
		}

		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("Failed to read request body", zap.Error(err))
			c.Next()
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		// Check for cached response
		ctx := context.Background()
		cacheKey := fmt.Sprintf("idempotency:payment:%s:%s", merchantID, idempotencyKey)

		cached, err := inits.RDB.Get(ctx, cacheKey).Result()
		if err == nil {
			// Cache hit - return cached response
			logger.Log.Info("Idempotency cache hit",
				zap.String("key", idempotencyKey),
				zap.String("merchant_id", merchantID.(string)),
			)

			var cachedResp map[string]interface{}
			if err = json.Unmarshal([]byte(cached), &cachedResp); err == nil {
				c.JSON(http.StatusOK, cachedResp)
				c.Abort()
				return
			}
		}

		hashKey := fmt.Sprintf("idempotency:hash:%s:%s", merchantID, idempotencyKey)
		requestHash := hashRequest(requestBody)

		existingHash, err := inits.RDB.Get(ctx, hashKey).Result()
		if err == nil && existingHash != "" && existingHash != requestHash {
			// Same key, different request = ERROR
			logger.Log.Warn("Idempotency key reused with different request",
				zap.String("key", idempotencyKey),
				zap.String("merchant_id", merchantID.(string)),
			)

			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   "idempotency key already used for different request",
			})
			c.Abort()
			return
		}

		inits.RDB.Set(ctx, hashKey, requestHash, idempotencyTTL)

		// Capture response
		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		c.Next()

		// Cache successful responses only (2xx)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			inits.RDB.Set(ctx, cacheKey, responseWriter.body.Bytes(), idempotencyTTL)

			logger.Log.Debug("Idempotency response cached",
				zap.String("key", idempotencyKey),
			)
		}
	}
}

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseBodyWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func hashRequest(body []byte) string {
	var normalized interface{}
	if err := json.Unmarshal(body, &normalized); err != nil {
		hash := sha256.Sum256(body)
		return hex.EncodeToString(hash[:])
	}

	normalizedBytes, err := json.Marshal(normalized)
	if err != nil {
		hash := sha256.Sum256(body)
		return hex.EncodeToString(hash[:])
	}

	hash := sha256.Sum256(normalizedBytes)
	return hex.EncodeToString(hash[:])
}
