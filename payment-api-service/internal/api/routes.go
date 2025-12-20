package api

import (
	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/payment-gateway/payment-api-service/inits/logger"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/handler"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/middleware"
	"github.com/rhaloubi/payment-gateway/payment-api-service/internal/service"
	"go.uber.org/zap"
)

func SetupRoutes(router *gin.Engine) {

	healthHandler := handler.NewHealthHandler()

	paymentHandler, err := handler.NewPaymentHandler()
	if err != nil {
		logger.Log.Fatal("Failed to initialize payment handler", zap.Error(err))
	}

	// NEW: Initialize payment intent handler
	paymentService, _ := service.NewPaymentService()
	paymentIntentHandler := handler.NewPaymentIntentHandler(paymentService)

	transactionHandler, err := handler.NewTransactionHandler()
	if err != nil {
		logger.Log.Fatal("Failed to initialize transaction handler", zap.Error(err))
	}

	router.Use(middleware.ErrorHandlerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RequestLoggerMiddleware())

	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/ready", healthHandler.ReadinessCheck)

	// =========================================================================
	// EXISTING API (v1) - Requires API Key
	// =========================================================================
	v1 := router.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware())
	v1.Use(middleware.RateLimitMiddleware())
	v1.Use(middleware.IdempotencyMiddleware())
	v1.Use(middleware.SanitizedBodyLoggerMiddleware())
	v1.Use(middleware.AuditLogMiddleware())
	{
		payments := v1.Group("/payments")
		{
			payments.POST("/authorize", paymentHandler.AuthorizePayment)
			payments.POST("/sale", paymentHandler.SalePayment)

			payments.POST("/:id/capture", paymentHandler.CapturePayment)
			payments.POST("/:id/void", paymentHandler.VoidPayment)
			payments.POST("/:id/refund", paymentHandler.RefundPayment)

			payments.GET("/:id", paymentHandler.GetPayment)
		}

		transactions := v1.Group("/transactions")
		{
			transactions.GET("/", transactionHandler.ListTransactions)
			transactions.GET("/:id", transactionHandler.GetTransaction)
		}

		// NEW: Payment Intents (Server-to-Server)
		paymentIntents := v1.Group("/payment-intents")
		{
			paymentIntents.POST("", paymentIntentHandler.CreatePaymentIntent)
			paymentIntents.POST("/:id/cancel", paymentIntentHandler.CancelPaymentIntent)
		}
	}

	// =========================================================================
	// NEW: PUBLIC API (No Auth) - For Hosted Checkout
	// =========================================================================
	public := router.Group("/api/public")
	public.Use(middleware.CORSMiddleware())
	public.Use(middleware.RequestLoggerMiddleware())
	{
		intents := public.Group("/payment-intents")
		{
			// Get payment intent (browser-safe, no secrets)
			intents.GET("/:id", paymentIntentHandler.GetPaymentIntent)

			// Confirm payment intent (process payment)
			intents.POST("/:id/confirm", paymentIntentHandler.ConfirmPaymentIntent)
		}
	}
}
