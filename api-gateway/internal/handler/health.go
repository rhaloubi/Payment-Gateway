package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/api-gateway/internal/config"
	"github.com/rhaloubi/api-gateway/internal/service"
)

func HealthCheck(cfg *config.Config, cb *service.CircuitBreaker) gin.HandlerFunc {
	return func(c *gin.Context) {
		health := gin.H{
			"status":  "ok",
			"service": "api-gateway",
			"version": "1.0.0",
			"services": gin.H{
				"auth":     cb.GetState("auth").String(),
				"merchant": cb.GetState("merchant").String(),
				"payment":  cb.GetState("payment").String(),
			},
		}
		c.JSON(http.StatusOK, health)
	}
}
