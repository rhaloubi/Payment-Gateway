package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"go.uber.org/zap"
)

// RequestLoggerMiddleware logs all incoming requests (PCI-safe)
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		// Start timer
		startTime := time.Now()

		// Get merchant ID
		merchantID, _ := c.Get("merchant_id")
		merchantIDStr := ""
		if merchantID != nil {
			merchantIDStr = merchantID.(string)
		}

		// Log request (sanitized)
		logger.Log.Info("Incoming request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("merchant_id", merchantIDStr),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)

		// Log response
		logger.Log.Info("Request completed",
			zap.String("request_id", requestID),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", duration.Milliseconds()),
			zap.Int("response_size", c.Writer.Size()),
		)
	}
}

// SanitizedBodyLoggerMiddleware logs request/response bodies (PCI-safe)
// NEVER logs full card numbers, CVV, or other sensitive data
func SanitizedBodyLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore body for handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Sanitize and log request body
		if len(requestBody) > 0 {
			sanitizedRequest := sanitizeRequestBody(requestBody)
			if sanitizedRequest != "" {
				requestID, _ := c.Get("request_id")
				logger.Log.Debug("Request body (sanitized)",
					zap.String("request_id", requestID.(string)),
					zap.String("body", sanitizedRequest),
				)
			}
		}

		c.Next()
	}
}

// sanitizeRequestBody removes sensitive data from request body
func sanitizeRequestBody(body []byte) string {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "" // Not JSON or invalid
	}

	// Remove sensitive fields
	sensitiveFields := []string{
		"card_number",
		"cvv",
		"cvv2",
		"cvc",
		"pan",
		"card_data",
	}

	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			// Mask card number (show only last 4)
			if field == "card_number" {
				cardNum, ok := data[field].(string)
				if ok && len(cardNum) >= 4 {
					data[field] = "****" + cardNum[len(cardNum)-4:]
				} else {
					data[field] = "****"
				}
			} else {
				// Completely remove CVV, etc.
				delete(data, field)
			}
		}
	}

	// Convert back to JSON
	sanitized, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(sanitized)
}

// MaskCardNumber masks card number (PCI requirement)
func MaskCardNumber(cardNumber string) string {
	if len(cardNumber) < 4 {
		return "****"
	}
	return "****" + cardNumber[len(cardNumber)-4:]
}

// AuditLogMiddleware logs security-relevant actions
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only audit sensitive operations
		sensitiveEndpoints := []string{
			"/v1/tokenize",
			"/v1/tokens",
			"/internal/v1/detokenize",
		}

		isSensitive := false
		for _, endpoint := range sensitiveEndpoints {
			if strings.Contains(c.Request.URL.Path, endpoint) {
				isSensitive = true
				break
			}
		}

		if !isSensitive {
			c.Next()
			return
		}

		// Collect audit info
		requestID, _ := c.Get("request_id")
		merchantID, _ := c.Get("merchant_id")
		userID, _ := c.Get("user_id")
		authType, _ := c.Get("auth_type")

		// Process request
		c.Next()

		// Log audit event
		logger.Log.Info("Audit log",
			zap.String("request_id", requestID.(string)),
			zap.String("action", c.Request.Method+" "+c.Request.URL.Path),
			zap.String("merchant_id", getString(merchantID)),
			zap.String("user_id", getString(userID)),
			zap.String("auth_type", getString(authType)),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		)
	}
}

// Helper to safely get string from interface
func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
