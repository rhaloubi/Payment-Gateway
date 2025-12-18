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
	authClient := service.NewAuthClient(cfg)

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

			// Protected auth routes
			authProtected := auth.Group("")
			authProtected.Use(middleware.AuthenticateJWT(authClient, cfg))
			{
				authProtected.GET("/profile", handler.ProxyRequest(cfg, "auth", circuitBreaker))
				authProtected.POST("/logout", handler.ProxyRequest(cfg, "auth", circuitBreaker))
				authProtected.POST("/change-password", handler.ProxyRequest(cfg, "auth", circuitBreaker))
				authProtected.GET("/sessions", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			}
		}

		// API Keys routes (JWT required)
		apiKeys := api.Group("/api-keys")
		apiKeys.Use(middleware.AuthenticateJWT(authClient, cfg))
		{
			apiKeys.POST("", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			apiKeys.GET("/merchant/:merchant_id", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			apiKeys.PATCH("/:id/deactivate", handler.ProxyRequest(cfg, "auth", circuitBreaker))
			apiKeys.DELETE("/:id", handler.ProxyRequest(cfg, "auth", circuitBreaker))
		}

		// Roles routes (JWT required)
		roles := api.Group("/roles")
		roles.Use(middleware.AuthenticateJWT(authClient, cfg))
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
		merchants.Use(middleware.AuthenticateJWT(authClient, cfg))
		{
			merchants.POST("", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.GET("", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.GET("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.PUT("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
			merchants.DELETE("/:id", handler.ProxyRequest(cfg, "merchant", circuitBreaker))
		}

		// Payment routes (API Key required)
		payments := api.Group("/payments")
		payments.Use(middleware.AuthenticateAPIKey(authClient, cfg))
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
	}

	return r
}
