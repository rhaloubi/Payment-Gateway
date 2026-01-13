package inits

import (
	"github.com/rhaloubi/payment-gateway/auth-service/config"

	"github.com/gin-gonic/gin"
)

var R *gin.Engine

func init() {
	// Set Gin mode
	ginMode := config.GetEnv("GIN_MODE")
	if ginMode == "" {
		ginMode = "release"
	}
	gin.SetMode(ginMode)

	// Create Gin router
	R = gin.New()
}
