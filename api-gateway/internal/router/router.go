package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rhaloubi/api-gateway/internal/config"
	"github.com/rhaloubi/api-gateway/internal/handler"
	"github.com/rhaloubi/api-gateway/internal/middleware"
	"github.com/rhaloubi/api-gateway/internal/service"
)

func Setup(cfg *config.Config) *gin.Engine {
	// Set mode
	if cfg.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// Global middleware
	r.Use(middleware.Logger(cfg))
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestID())

	// Initialize services
	rateLimiter := service.NewRateLimiter(cfg)
	circuitBreaker := service.NewCircuitBreaker(cfg)

	// Health and metrics endpoints (no auth required)
	r.GET("/health", handler.HealthCheck(cfg, circuitBreaker))
	r.GET("/metrics", handler.Metrics())

	// API routes with full middleware stack
	api := r.Group("/api/v1")
	{
		// Apply global rate limiting
		if cfg.RateLimiting.Enabled {
			api.Use(middleware.RateLimiter(rateLimiter, cfg))
		}

		// Authentication routes (no auth required)
		auth := api.Group("/auth")
		{
			// Special rate limits for auth endpoints
			auth.POST("/register",
				middleware.EndpointRateLimit(rateLimiter, "register", 3, time.Hour),
				handler.ProxyRequest(cfg, "auth", circuitBreaker),
			)

			auth.POST("/login",
				middleware.EndpointRateLimit(rateLimiter, "login", 5, time.Minute),
				handler.ProxyRequest(cfg, "auth", circuitBreaker),
			)

			auth.POST("/refresh", handler.ProxyRequest(cfg, "auth", circuitBreaker))

			auth.GET("/profile", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			auth.POST("/logout", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			auth.POST("/change-password", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			auth.GET("/sessions", handler.ProxyRequest(cfg, "auth", circuitBreaker))

		}

		// Roles routes (JWT required)
		roles := api.Group("/roles")
		{
			roles.GET("", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			roles.GET("/:id", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			roles.POST("/assign", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			roles.DELETE("/assign", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			roles.GET("/user/:user_id/merchant/:merchant_id", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			roles.GET("/user/:user_id/merchant/:merchant_id/permissions", handler.ProxyRequest(cfg, "auth", circuitBreaker))
		}

		// Merchant routes (JWT required)
		merchants := api.Group("/merchants")
		{
			merchants.POST("", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.GET("", handler.ProxyRequest(cfg, "merchant", circuitBreaker))

			// Merchant API Keys
			merchantApiKeys := merchants.Group("/api-keys")
			{
				merchantApiKeys.POST("", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
				merchantApiKeys.GET("/merchant/:merchant_id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
				merchantApiKeys.PATCH("/:merchant_id/:id/deactivate", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
				merchantApiKeys.DELETE("/:merchant_id/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			}

			merchants.GET("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.GET("/:id/details", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.GET("/:id/team", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.GET("/:id/invitations", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.GET("/:id/settings", handler.ProxyRequest(cfg, "merchant", circuitBreaker))

			merchants.PUT("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.PATCH("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.PATCH("/:id/settings", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.PATCH("/:id/team/:user_id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))

			merchants.POST("/:id/team/invite", handler.ProxyRequest(cfg, "merchant", circuitBreaker))

			merchants.DELETE("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.DELETE("/:id/team/:user_id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))

		}
		// Invitation routes (JWT required)
		invitations := api.Group("/invitations")
		{
			invitations.POST("/:token/accept", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			invitations.DELETE("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
		}

		// Payment routes (API Key required)
		payments := api.Group("/payments")
		payments.Use(middleware.EndpointRateLimit(rateLimiter, "payments", 20, time.Second))
		{
			payments.POST("/authorize", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			payments.POST("/sale", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			payments.POST("/:id/capture", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			payments.POST("/:id/void", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			payments.POST("/:id/refund", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			payments.GET("/:id", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			payments.GET("", handler.ProxyRequest(cfg, "payment", circuitBreaker))
		}
		transactions := api.Group("/transactions")
		{
			transactions.GET("", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			transactions.GET("/:id", handler.ProxyRequest(cfg, "payment", circuitBreaker))
		}
		paymentIntents := api.Group("/payment-intents")
		{
			paymentIntents.POST("", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			paymentIntents.POST("/:id/cancel", handler.ProxyRequest(cfg, "payment", circuitBreaker))
		}

	}
	public := r.Group("/api/public")
	{
		intents := public.Group("/payment-intents")
		{
			intents.GET("/:id", handler.ProxyRequest(cfg, "payment", circuitBreaker))
			intents.POST("/:id/confirm", handler.ProxyRequest(cfg, "payment", circuitBreaker))
		}
	}

	return r
}
