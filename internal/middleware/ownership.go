package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type OwnershipChecker interface {
	VerifyOwnership(ctx context.Context, resourceID, userID uuid.UUID) (bool, error)
}

func RequireOwnership(checker OwnershipChecker, paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {

		userIDRaw, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{"code": "UNAUTHORIZED", "message": "Not authenticated"},
			})
			c.Abort()
			return
		}
		userID := userIDRaw.(uuid.UUID)

		resourceIDStr := c.Param(paramName)
		if resourceIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "BAD_REQUEST", "message": "Missing resource ID"},
			})
			c.Abort()
			return
		}

		resourceID, err := uuid.Parse(resourceIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{"code": "BAD_REQUEST", "message": "Invalid resource ID"},
			})
			c.Abort()
			return
		}

		isOwner, err := checker.VerifyOwnership(c.Request.Context(), resourceID, userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{"code": "NOT_FOUND", "message": "Resource not found"},
			})
			c.Abort()
			return
		}

		if !isOwner {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{"code": "FORBIDDEN", "message": "Access denied"},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
