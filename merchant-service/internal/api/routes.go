package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/client"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/handler"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/middleware"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/service"
)

func SetupMerchantRoutes() {
	router := inits.R

	authClient := client.NewAuthServiceClient()
	merchantHandler := handler.NewMerchantHandler()
	teamHandler := handler.NewTeamHandler()
	settingsHandler := handler.NewSettingsHandler()
	apiKeyHandler := handler.NewAPIKeyHandler(authClient, service.NewTeamService())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "health check",
		})
	})

	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware())
	{
		// Merchant routes
		merchants := v1.Group("/merchants")
		{
			merchants.POST("", merchantHandler.CreateMerchant)
			merchants.GET("", merchantHandler.ListUserMerchants)

			apiKeys := merchants.Group("/api-keys")
			{
				apiKeys.POST("", apiKeyHandler.CreateAPIKey)
				apiKeys.GET("/merchant/:merchant_id", apiKeyHandler.GetMerchantAPIKeys)
				apiKeys.PATCH("/:merchant_id/:id/deactivate", apiKeyHandler.DeactivateAPIKey)
				apiKeys.DELETE("/:merchant_id/:id", apiKeyHandler.DeleteAPIKey)

			}

			merchantGroup := merchants.Group("/:id")
			merchantGroup.Use(middleware.RequireMerchantAccess())
			{
				// Read operations - available to all roles
				merchantGroup.GET("", middleware.RequireRolePermission("read"), merchantHandler.GetMerchant)
				merchantGroup.GET("/details", middleware.RequireRolePermission("read"), merchantHandler.GetMerchantDetails)
				merchantGroup.GET("/team", middleware.RequireRolePermission("read"), teamHandler.GetTeamMembers)
				merchantGroup.GET("/invitations", middleware.RequireRolePermission("read"), teamHandler.GetPendingInvitations)
				merchantGroup.GET("/settings", middleware.RequireRolePermission("read"), settingsHandler.GetSettings)

				// Update operations - Owner and Admin only
				merchantGroup.PATCH("", middleware.RequireRolePermission("update"), merchantHandler.UpdateMerchant)
				merchantGroup.PATCH("/settings", middleware.RequireRolePermission("update"), settingsHandler.UpdateSettings)
				merchantGroup.PATCH("/team/:user_id", middleware.RequireRolePermission("update"), teamHandler.UpdateTeamMemberRole)

				// Create operations - Owner, Admin, and Manager
				merchantGroup.POST("/team/invite", middleware.RequireRolePermission("create"), teamHandler.InviteTeamMember)

				// Delete operations - Owner only (Admin cannot delete)
				merchantGroup.DELETE("", middleware.RequireRolePermission("delete"), merchantHandler.DeleteMerchant)
				merchantGroup.DELETE("/team/:user_id", middleware.RequireRolePermission("delete"), teamHandler.RemoveTeamMember)
			}
		}

		// Invitation routes (public with auth)
		invitations := v1.Group("/invitations")
		{
			invitations.POST("/:token/accept", teamHandler.AcceptInvitation)
			invitations.DELETE("/:id", teamHandler.CancelInvitation)
		}
	}
}
