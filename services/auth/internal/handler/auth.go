package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/sur1cat/aifa/auth-service/internal/domain"
	"github.com/sur1cat/aifa/auth-service/internal/events"
	"github.com/sur1cat/aifa/auth-service/internal/jwt"
	"github.com/sur1cat/aifa/auth-service/internal/oauth"
	"github.com/sur1cat/aifa/auth-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	jwt       *jwt.Manager
	google    *oauth.GoogleVerifier
	apple     *oauth.AppleVerifier
	users     *repository.UserRepository
	blacklist *repository.Blacklist
	publisher *events.Publisher
}

func NewAuthHandler(
	jwtMgr *jwt.Manager,
	google *oauth.GoogleVerifier,
	apple *oauth.AppleVerifier,
	users *repository.UserRepository,
	blacklist *repository.Blacklist,
	publisher *events.Publisher,
) *AuthHandler {
	return &AuthHandler{jwt: jwtMgr, google: google, apple: apple, users: users, blacklist: blacklist, publisher: publisher}
}

type userDTO struct {
	ID           string `json:"id"`
	AuthProvider string `json:"auth_provider"`
	CreatedAt    string `json:"created_at"`
}

func toDTO(u *domain.User) userDTO {
	return userDTO{ID: u.ID.String(), AuthProvider: string(u.AuthProvider), CreatedAt: u.CreatedAt.Format(time.RFC3339)}
}

type authResponse struct {
	User      userDTO        `json:"user"`
	Tokens    *jwt.TokenPair `json:"tokens"`
	IsNewUser bool           `json:"is_new_user"`
}

type providerProfile struct {
	Sub       string
	Email     string
	Name      string
	AvatarURL string
}

func (h *AuthHandler) signIn(c *gin.Context, provider domain.AuthProvider, p providerProfile) {
	user, isNew, err := h.users.FindOrCreate(c.Request.Context(), provider, p.Sub)
	if err != nil {
		slog.Error("find or create user", "err", err, "provider", provider)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to process user")
		return
	}

	if isNew {
		h.publisher.PublishUserProvisioned(events.UserProvisioned{
			UserID:    user.ID.String(),
			Provider:  string(provider),
			Email:     p.Email,
			Name:      p.Name,
			AvatarURL: p.AvatarURL,
		})
	}

	tokens, err := h.jwt.GenerateTokenPair(user.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to generate tokens")
		return
	}

	resp := authResponse{User: toDTO(user), Tokens: tokens, IsNewUser: isNew}
	if isNew {
		respondCreated(c, resp)
		return
	}
	respondOK(c, resp)
}

// revokeCurrentAccess blacklists the JTI from the request's access token so
// subsequent hits from the same token fail the middleware check.
func (h *AuthHandler) revokeCurrentAccess(c *gin.Context) {
	jti, ok := c.Value("jti").(string)
	if !ok {
		return
	}
	_ = h.blacklist.Revoke(c.Request.Context(), jti, h.jwt.AccessTTL())
}

type googleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

func (h *AuthHandler) GoogleSignIn(c *gin.Context) {
	var req googleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	info, err := h.google.Verify(c.Request.Context(), req.IDToken)
	if err != nil {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Invalid Google token")
		return
	}
	h.signIn(c, domain.ProviderGoogle, providerProfile{
		Sub:       info.Sub,
		Email:     info.Email,
		Name:      info.Name,
		AvatarURL: info.Picture,
	})
}

type appleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
	User    string `json:"user"`
}

func (h *AuthHandler) AppleSignIn(c *gin.Context) {
	var req appleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	info, err := h.apple.Verify(c.Request.Context(), req.IDToken)
	if err != nil {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Invalid Apple token")
		return
	}
	h.signIn(c, domain.ProviderApple, providerProfile{
		Sub:   info.Sub,
		Email: info.Email,
		Name:  oauth.ParseAppleUserName(req.User),
	})
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	claims, err := h.jwt.ValidateRefresh(req.RefreshToken)
	if err != nil {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Invalid or expired refresh token")
		return
	}

	revoked, err := h.blacklist.IsRevoked(c.Request.Context(), claims.ID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Could not verify token state")
		return
	}
	if revoked {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Refresh token has been revoked")
		return
	}

	tokens, err := h.jwt.GenerateTokenPair(claims.UserID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to generate tokens")
		return
	}
	respondOK(c, tokens)
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, ok := c.Value("userID").(uuid.UUID)
	if !ok {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Not authenticated")
		return
	}
	user, err := h.users.GetByID(c.Request.Context(), userID)
	if err != nil {
		respondError(c, http.StatusNotFound, codeNotFound, "User not found")
		return
	}
	respondOK(c, toDTO(user))
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	// Best-effort revocation; an invalid refresh token still lets us reply OK
	// because the access token is revoked below either way.
	if claims, err := h.jwt.ValidateRefresh(req.RefreshToken); err == nil {
		_ = h.blacklist.Revoke(c.Request.Context(), claims.ID, time.Until(claims.ExpiresAt.Time))
	}
	h.revokeCurrentAccess(c)

	respondMessage(c, "Logged out successfully")
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID, ok := c.Value("userID").(uuid.UUID)
	if !ok {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Not authenticated")
		return
	}

	if err := h.users.Delete(c.Request.Context(), userID); err != nil {
		slog.Error("delete user", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete account")
		return
	}

	h.publisher.PublishUserDeleted(events.UserDeleted{UserID: userID.String()})
	h.revokeCurrentAccess(c)

	respondMessage(c, "Account deleted successfully")
}
