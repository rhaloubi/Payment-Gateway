package inits

import (
	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/payment-api-service/config"
)

var R *gin.Engine

func init() {
	//Set Gin mode
	ginMode := config.GetEnv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	gin.SetMode(ginMode)

	// Create Gin router
	R = gin.New()
}
