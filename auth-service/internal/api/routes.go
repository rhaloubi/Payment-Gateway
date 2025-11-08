package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/handler"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/middleware"
	//"github.com/rhaloubi/go-learning/db/controllers"
)

func Routes() {
	r := inits.R
	authHandler := handler.NewAuthHandler()
	// Define your routes here
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "health check",
		})
	})

	// /api/v1/*
	v1 := r.Group("/api/v1")
	{
		// Public auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}

		authProtected := v1.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware())
		{
			authProtected.GET("/profile", authHandler.GetProfile)
			authProtected.POST("/logout", authHandler.Logout)
			authProtected.POST("/change-password", authHandler.ChangePassword)
			authProtected.GET("/sessions", authHandler.GetSessions)
		}
		// ***
	}
}
