package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/inits/jwt"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/service"
)

// AuthMiddleware returns the JWT auth middleware (lazy initialization)
func AuthMiddleware() gin.HandlerFunc {
	return jwt.NewJWTValidator().AuthMiddleware()
}

func RequireMerchantAccess() gin.HandlerFunc {
	teamService := service.NewTeamService()
	jwtValidator := jwt.NewJWTValidator()

	return func(c *gin.Context) {
		// Get merchant ID from URL parameter
		merchantID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid merchant ID",
			})
			c.Abort()
			return
		}

		userID, err := jwtValidator.GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Check if user has access to merchant
		hasAccess, err := teamService.IsUserInMerchant(merchantID, userID)
		if err != nil || !hasAccess {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "access denied - you are not a member of this merchant",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
