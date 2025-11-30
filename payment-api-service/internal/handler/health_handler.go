package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck handles GET /health
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "payment-api-service",
		"time":    time.Now().Format(time.RFC3339),
	})
}

// ReadinessCheck handles GET /ready
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	checks := map[string]bool{
		"database": false,
		"redis":    false,
	}

	// Check PostgreSQL
	if sqlDB, err := inits.DB.DB(); err == nil {
		if err := sqlDB.PingContext(ctx); err == nil {
			checks["database"] = true
		}
	}

	// Check Redis
	if err := inits.RDB.Ping(ctx).Err(); err == nil {
		checks["redis"] = true
	}

	// All checks must pass
	ready := checks["database"] && checks["redis"]

	status := http.StatusOK
	if !ready {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, gin.H{
		"ready":  ready,
		"checks": checks,
		"time":   time.Now().Format(time.RFC3339),
	})
}
