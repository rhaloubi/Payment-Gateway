package handler

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/api-gateway/internal/config"
	"github.com/rhaloubi/api-gateway/internal/service"
)

func ProxyRequest(cfg *config.Config, targetService string, cb *service.CircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := cb.Allow(targetService); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error":   fmt.Sprintf("service temporarily unavailable: %s", targetService),
			})
			return
		}

		var serviceURL string
		var timeout time.Duration

		switch targetService {
		case "auth":
			serviceURL = cfg.Services.Auth.URL
			timeout = cfg.Services.Auth.Timeout
		case "merchant":
			serviceURL = cfg.Services.Merchant.URL
			timeout = cfg.Services.Merchant.Timeout
		case "payment":
			serviceURL = cfg.Services.Payment.URL
			timeout = cfg.Services.Payment.Timeout
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "unknown service",
			})
			return
		}

		targetURL := serviceURL + c.Request.URL.Path
		if c.Request.URL.RawQuery != "" {
			targetURL += "?" + c.Request.URL.RawQuery
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			cb.RecordFailure(targetService)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to create proxy request",
			})
			return
		}

		for key, values := range c.Request.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		proxyReq.Header.Set("X-Forwarded-For", c.ClientIP())
		proxyReq.Header.Set("X-Request-ID", c.GetString("request_id"))

		client := &http.Client{Timeout: timeout}
		start := time.Now()
		resp, err := client.Do(proxyReq)
		duration := time.Since(start)

		if err != nil {
			cb.RecordFailure(targetService)
			c.JSON(http.StatusBadGateway, gin.H{
				"success": false,
				"error":   fmt.Sprintf("service request failed: %v", err),
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 500 {
			cb.RecordFailure(targetService)
		} else {
			cb.RecordSuccess(targetService)
		}

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to read response",
			})
			return
		}

		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		c.Header("X-Gateway-Response-Time", fmt.Sprintf("%dms", duration.Milliseconds()))
		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
	}
}
