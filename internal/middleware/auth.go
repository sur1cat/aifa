package middleware

import (
	"net/http"
	"strings"

	"habitflow/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{jwtManager: jwtManager}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "MISSING_TOKEN",
					"message": "Authorization header is required",
				},
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN_FORMAT",
					"message": "Authorization header must be in format: Bearer <token>",
				},
			})
			return
		}

		claims, err := m.jwtManager.ValidateAccessToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired access token",
				},
			})
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}

	id, ok := userID.(uuid.UUID)
	return id, ok
}
