package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/tokenization-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/service"
	"go.uber.org/zap"
)

func IdempotencyMiddleware() gin.HandlerFunc {
	idempotencyService := service.NewIdempotencyService()

	return func(c *gin.Context) {
		idempotencyKey := c.GetHeader("Idempotency-Key")

		if idempotencyKey == "" {
			c.Next()
			return
		}

		if err := idempotencyService.ValidateIdempotencyKey(idempotencyKey); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid idempotency key: " + err.Error(),
			})
			c.Abort()
			return
		}

		merchantIDStr, exists := c.Get("merchant_id")
		if !exists {
			userID, userExists := c.Get("user_id")
			if !userExists {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error":   "authentication required for idempotency",
				})
				c.Abort()
				return
			}
			merchantIDStr = userID
		}

		merchantID, err := uuid.Parse(merchantIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid merchant_id",
			})
			c.Abort()
			return
		}

		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			logger.Log.Error("Failed to read request body", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to process request",
			})
			c.Abort()
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		cachedResponse, err := idempotencyService.CheckRequest(
			merchantID,
			idempotencyKey,
			requestBody,
			c.Request.URL.Path,
			c.Request.Method,
		)

		if err != nil {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		if cachedResponse != nil {
			c.Data(cachedResponse.ResponseStatus, "application/json", cachedResponse.ResponseBody)
			c.Abort()
			return
		}

		c.Set("idempotency_key", idempotencyKey)
		c.Set("idempotency_merchant_id", merchantID)
		c.Set("idempotency_request_body", requestBody)
		c.Set("idempotency_endpoint", c.Request.URL.Path)
		c.Set("idempotency_method", c.Request.Method)

		responseWriter := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			err := idempotencyService.StoreResponse(
				merchantID,
				idempotencyKey,
				requestBody,
				responseWriter.body.Bytes(),
				c.Writer.Status(),
				c.Request.URL.Path,
				c.Request.Method,
				c.ClientIP(),
				c.Request.UserAgent(),
			)

			if err != nil {
				logger.Log.Error("Failed to store idempotent response",
					zap.Error(err),
					zap.String("key", idempotencyKey),
				)
				// Don't fail the request, just log the error
			}
		}
	}
}

// responseBodyWriter captures response body for caching
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
