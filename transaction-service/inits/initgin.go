package inits

import (
	"os"

	"github.com/gin-gonic/gin"
)

var R *gin.Engine

func init() {
	//Set Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}
	gin.SetMode(ginMode)

	// Create Gin router
	R = gin.New()
}
