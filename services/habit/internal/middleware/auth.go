package middleware

import (
	"net/http"
	"strings"

	"github.com/sur1cat/aifa/habit-service/internal/jwt"
	"github.com/sur1cat/aifa/habit-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Auth struct {
	jwt       *jwt.Validator
	blacklist *repository.Blacklist
}

func NewAuth(v *jwt.Validator, bl *repository.Blacklist) *Auth {
	return &Auth{jwt: v, blacklist: bl}
}

func (a *Auth) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			abort(c, http.StatusUnauthorized, "MISSING_TOKEN", "Authorization header is required")
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			abort(c, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Authorization header must be: Bearer <token>")
			return
		}

		claims, err := a.jwt.ValidateAccess(parts[1])
		if err != nil {
			abort(c, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired access token")
			return
		}

		revoked, err := a.blacklist.IsRevoked(c.Request.Context(), claims.ID)
		if err != nil {
			abort(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Could not verify token state")
			return
		}
		if revoked {
			abort(c, http.StatusUnauthorized, "TOKEN_REVOKED", "Token has been revoked")
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}

func UserID(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("userID")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

func abort(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}
