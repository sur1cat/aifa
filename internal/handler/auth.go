package handler

import (
	"habitflow/internal/domain"
	"habitflow/internal/repository"
	"habitflow/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	jwtManager     *auth.JWTManager
	googleVerifier *auth.GoogleTokenVerifier
	appleVerifier  *auth.AppleTokenVerifier
	userRepo       *repository.UserRepository
	tokenRepo      *repository.TokenRepository
}

func NewAuthHandler(
	jwtManager *auth.JWTManager,
	googleVerifier *auth.GoogleTokenVerifier,
	appleVerifier *auth.AppleTokenVerifier,
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
) *AuthHandler {
	return &AuthHandler{
		jwtManager:     jwtManager,
		googleVerifier: googleVerifier,
		appleVerifier:  appleVerifier,
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
	}
}

type AuthResponse struct {
	User      *UserResponse   `json:"user"`
	Tokens    *auth.TokenPair `json:"tokens"`
	IsNewUser bool            `json:"is_new_user"`
}

type UserResponse struct {
	ID           string  `json:"id"`
	Email        *string `json:"email,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	Name         *string `json:"name,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	AuthProvider string  `json:"auth_provider"`
	CreatedAt    string  `json:"created_at"`
}

func toUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:           user.ID.String(),
		Email:        user.Email,
		Phone:        user.Phone,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		AuthProvider: string(user.AuthProvider),
		CreatedAt:    user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type GoogleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

func (h *AuthHandler) GoogleSignIn(c *gin.Context) {
	var req GoogleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	googleUser, err := h.googleVerifier.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		respondUnauthorized(c, "Invalid Google token")
		return
	}

	var email, name, avatar *string
	if googleUser.Email != "" {
		email = &googleUser.Email
	}
	if googleUser.Name != "" {
		name = &googleUser.Name
	}
	if googleUser.Picture != "" {
		avatar = &googleUser.Picture
	}

	user, isNew, err := h.userRepo.FindOrCreateByProvider(
		c.Request.Context(),
		domain.AuthProviderGoogle,
		googleUser.Sub,
		email, name, avatar,
	)
	if err != nil {
		respondInternalError(c, "Failed to process user")
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.ID)
	if err != nil {
		respondInternalError(c, "Failed to generate tokens")
		return
	}

	response := AuthResponse{
		User:      toUserResponse(user),
		Tokens:    tokens,
		IsNewUser: isNew,
	}

	if isNew {
		respondCreated(c, response)
	} else {
		respondOK(c, response)
	}
}

type AppleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
	User    string `json:"user"`
}

func (h *AuthHandler) AppleSignIn(c *gin.Context) {
	var req AppleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	appleUser, err := h.appleVerifier.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		respondUnauthorized(c, "Invalid Apple token")
		return
	}

	var email, name *string
	if appleUser.Email != "" {
		email = &appleUser.Email
	}
	if userName := auth.ParseAppleUserName(req.User); userName != "" {
		name = &userName
	}

	user, isNew, err := h.userRepo.FindOrCreateByProvider(
		c.Request.Context(),
		domain.AuthProviderApple,
		appleUser.Sub,
		email, name, nil,
	)
	if err != nil {
		respondInternalError(c, "Failed to process user")
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.ID)
	if err != nil {
		respondInternalError(c, "Failed to generate tokens")
		return
	}

	response := AuthResponse{
		User:      toUserResponse(user),
		Tokens:    tokens,
		IsNewUser: isNew,
	}

	if isNew {
		respondCreated(c, response)
	} else {
		respondOK(c, response)
	}
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	if invalidated, _ := h.tokenRepo.IsInvalidated(c.Request.Context(), req.RefreshToken); invalidated {
		respondUnauthorized(c, "Token has been revoked")
		return
	}

	tokens, err := h.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		respondUnauthorized(c, "Invalid or expired refresh token")
		return
	}

	respondOK(c, tokens)
}

func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		respondUnauthorized(c, "Not authenticated")
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), userID.(uuid.UUID))
	if err != nil {
		respondNotFound(c, "User not found")
		return
	}

	respondOK(c, toUserResponse(user))
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	claims, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {

		respondMessage(c, "Logged out successfully")
		return
	}

	_ = h.tokenRepo.Invalidate(c.Request.Context(), req.RefreshToken, claims.UserID, claims.ExpiresAt.Time)
	respondMessage(c, "Logged out successfully")
}

func (h *AuthHandler) DeleteAccount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		respondUnauthorized(c, "Not authenticated")
		return
	}

	if err := h.userRepo.Delete(c.Request.Context(), userID.(uuid.UUID)); err != nil {
		respondInternalError(c, "Failed to delete account")
		return
	}

	respondMessage(c, "Account deleted successfully")
}
