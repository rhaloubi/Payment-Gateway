package api

import (
	"github.com/gin-gonic/gin"
	//"github.com/rhaloubi/go-learning/db/controllers"
)

var R *gin.Engine

func Routes() {
	R = gin.Default()
	// Define your routes here
	R.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "health check",
		})
	})
	/*R.POST("/accounts", controllers.CreateAccount)
	R.GET("/accounts", controllers.GetAccounts)
	R.GET("/accounts/:id", controllers.GetAccountByID)
	R.PUT("/accounts/:id", controllers.UpdateAccount)
	R.PATCH("/accounts/:id", controllers.UpdateAccount)
	R.DELETE("/accounts/:id", controllers.DeleteAccount)*/
}
