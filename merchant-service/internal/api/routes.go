package api

import (
	"github.com/rhaloubi/payment-gateway/merchant-service/inits"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/handler"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/middleware"
)

// SetupMerchantRoutes sets up all merchant-related routes
func SetupMerchantRoutes() {
	router := inits.R

	merchantHandler := handler.NewMerchantHandler()

	// API v1 group
	v1 := router.Group("/api/v1")
	//v1.Use(authMiddleware.AuthMiddleware())
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
			}
		}
	}
}
