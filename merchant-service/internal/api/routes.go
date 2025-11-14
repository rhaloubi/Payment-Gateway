package api

import (
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/handler"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/middleware"
)

func SetupMerchantRoutes() {
	router := inits.R

	merchantHandler := handler.NewMerchantHandler()
	teamHandler := handler.NewTeamHandler()
	settingsHandler := handler.NewSettingsHandler()

	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware())
	{
		// Merchant routes
		merchants := v1.Group("/merchants")
		{
			merchants.POST("", merchantHandler.CreateMerchant)
			merchants.GET("", merchantHandler.ListUserMerchants)

			// Merchant-specific routes (require merchant access)
			merchantGroup := merchants.Group("/:id")
			merchantGroup.Use(middleware.RequireMerchantAccess())
			{
				merchantGroup.GET("", merchantHandler.GetMerchant)
				merchantGroup.GET("/details", merchantHandler.GetMerchantDetails)
				merchantGroup.PATCH("", merchantHandler.UpdateMerchant)
				merchantGroup.DELETE("", merchantHandler.DeleteMerchant)

				merchantGroup.POST("/team/invite", teamHandler.InviteTeamMember)
				merchantGroup.GET("/team", teamHandler.GetTeamMembers)
				merchantGroup.DELETE("/team/:user_id", teamHandler.RemoveTeamMember)
				merchantGroup.PATCH("/team/:user_id", teamHandler.UpdateTeamMemberRole)
				merchantGroup.GET("/invitations", teamHandler.GetPendingInvitations)

				merchantGroup.GET("/settings", settingsHandler.GetSettings)
				merchantGroup.PATCH("/settings", settingsHandler.UpdateSettings)
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
