package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/handler"
	"github.com/rhaloubi/payment-gateway/tokenization-service/internal/middleware"
)

func SetupRoutes(router *gin.Engine) {
	// Initialize handlers
	tokenizationHandler := handler.NewTokenizationHandler()

	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestLoggerMiddleware())

	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware())
	v1.Use(middleware.RateLimitMiddleware())
	v1.Use(middleware.IdempotencyMiddleware())
	v1.Use(middleware.SanitizedBodyLoggerMiddleware())
	v1.Use(middleware.AuditLogMiddleware())
	{
		// Tokenization
		v1.POST("/tokenize", tokenizationHandler.TokenizeCard)

		// Token Management
		tokens := v1.Group("/tokens")
		{
			tokens.GET("/:token/validate", tokenizationHandler.ValidateToken)
			tokens.GET("/:token", tokenizationHandler.GetTokenInfo)
			tokens.DELETE("/:token", tokenizationHandler.RevokeToken)
		}

		keys := v1.Group("/keys")
		{
			keys.GET("/statistics", tokenizationHandler.GetKeyStatistics)
			keys.POST("/rotate", tokenizationHandler.RotateKey)
		}
	}

	internal := router.Group("/internal/v1")
	internal.Use(middleware.InternalServiceMiddleware())
	internal.Use(middleware.RequestLoggerMiddleware())
	{
		internal.POST("/detokenize", tokenizationHandler.Detokenize)
	}
}
