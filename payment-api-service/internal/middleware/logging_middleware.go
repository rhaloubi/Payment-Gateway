package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"go.uber.org/zap"
)

// RequestLoggerMiddleware logs all incoming requests (PCI-safe)
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()
		c.Set("request_id", requestID)

		startTime := time.Now()

		merchantID, _ := c.Get("merchant_id")
		merchantIDStr := ""
		if merchantID != nil {
			merchantIDStr = merchantID.(string)
		}

		logger.Log.Info("Incoming payment request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("merchant_id", merchantIDStr),
			zap.String("ip", c.ClientIP()),
		)

		c.Next()

		duration := time.Since(startTime)

		logger.Log.Info("Payment request completed",
			zap.String("request_id", requestID),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", duration.Milliseconds()),
		)
	}
}

// SanitizedBodyLoggerMiddleware logs request bodies (PCI-safe - NEVER logs card data)
func SanitizedBodyLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		if len(requestBody) > 0 {
			sanitized := sanitizePaymentRequest(requestBody)
			if sanitized != "" {
				requestID, _ := c.Get("request_id")
				logger.Log.Debug("Request body (sanitized)",
					zap.String("request_id", requestID.(string)),
					zap.String("body", sanitized),
				)
			}
		}

		c.Next()
	}
}

// sanitizePaymentRequest removes ALL sensitive data from payment requests
func sanitizePaymentRequest(body []byte) string {
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}

	// Remove card object completely (PCI requirement)
	if _, exists := data["card"]; exists {
		cardInfo := map[string]interface{}{
			"present": true,
		}
		data["card"] = cardInfo
	}

	// Keep safe fields only
	safeFields := map[string]interface{}{
		"amount":      data["amount"],
		"currency":    data["currency"],
		"description": data["description"],
	}

	sanitized, err := json.Marshal(safeFields)
	if err != nil {
		return ""
	}

	return string(sanitized)
}

// AuditLogMiddleware logs security-relevant payment actions
func AuditLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sensitiveEndpoints := []string{
			"/v1/payments/authorize",
			"/v1/payments/sale",
			"/v1/payments",
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

		requestID, _ := c.Get("request_id")
		merchantID, _ := c.Get("merchant_id")
		authType, _ := c.Get("auth_type")

		c.Next()

		logger.Log.Info("Payment audit log",
			zap.String("request_id", getString(requestID)),
			zap.String("action", c.Request.Method+" "+c.Request.URL.Path),
			zap.String("merchant_id", getString(merchantID)),
			zap.String("auth_type", getString(authType)),
			zap.Int("status", c.Writer.Status()),
			zap.String("ip", c.ClientIP()),
		)
	}
}

func getString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
