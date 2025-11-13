package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/auth-service/inits"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/handler"
	"github.com/rhaloubi/payment-gateway/auth-service/internal/middleware"
)

func Routes() {
	r := inits.R
	authHandler := handler.NewAuthHandler()
	roleHandler := handler.NewRoleHandler()
	apiKeyHandler := handler.NewAPIKeyHandler()
	internalHandler := handler.NewInternalHandler()

	// Define your routes here
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "health check",
		})
	})

	internal := r.Group("/internal/v1")
	//internal.Use(middleware.InternalServiceMiddleware())
	{
		// Role assignment for merchant owners
		internal.POST("/roles/assign-merchant-owner", internalHandler.AssignMerchantOwnerRole)

		// Get user roles (for merchant service to check access)
		internal.GET("/users/:user_id/roles", internalHandler.GetUserRolesByUserID)
	}

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
		roles := v1.Group("/roles")
		roles.Use(middleware.AuthMiddleware())
		{
			roles.GET("", roleHandler.GetAllRoles)
			roles.GET("/:id", roleHandler.GetRoleByID)
			roles.POST("/assign", roleHandler.AssignRoleToUser)
			roles.DELETE("/assign", roleHandler.RemoveRoleFromUser)
			roles.GET("/user/:user_id/merchant/:merchant_id", roleHandler.GetUserRoles)
			roles.GET("/user/:user_id/merchant/:merchant_id/permissions", roleHandler.GetUserPermissions)
		}
		// ***
		apiKeys := v1.Group("/api-keys")
		apiKeys.Use(middleware.AuthMiddleware())
		{
			apiKeys.POST("", apiKeyHandler.CreateAPIKey)
			apiKeys.GET("/merchant/:merchant_id", apiKeyHandler.GetMerchantAPIKeys)
			apiKeys.PATCH("/:id/deactivate", apiKeyHandler.DeactivateAPIKey)
			apiKeys.DELETE("/:id", apiKeyHandler.DeleteAPIKey)
		}
	}
}
