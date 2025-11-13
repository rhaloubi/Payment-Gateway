package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rhaloubi/payment-gateway/merchant-service/internal/service"
)

// RequireMerchantAccess checks if user has access to merchant
func RequireMerchantAccess() gin.HandlerFunc {
	teamService := service.NewTeamService()

	return func(c *gin.Context) {
		merchantID, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid merchant ID",
			})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
			})
			c.Abort()
			return
		}

		userUUID, _ := uuid.Parse(userID.(string))

		hasAccess, err := teamService.IsUserInMerchant(merchantID, userUUID)
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
