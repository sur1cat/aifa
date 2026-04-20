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

// Response types
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

// ========== Google Sign-In ==========

type GoogleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
}

// GoogleSignIn handles Google Sign-In
// POST /api/v1/auth/google
func (h *AuthHandler) GoogleSignIn(c *gin.Context) {
	var req GoogleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	// Verify Google token
	googleUser, err := h.googleVerifier.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		respondUnauthorized(c, "Invalid Google token")
		return
	}

	// Find or create user
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

	// Generate tokens
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

// ========== Apple Sign-In ==========

type AppleSignInRequest struct {
	IDToken string `json:"id_token" binding:"required"`
	User    string `json:"user"` // Only sent on first sign in
}

// AppleSignIn handles Apple Sign-In
// POST /api/v1/auth/apple
func (h *AuthHandler) AppleSignIn(c *gin.Context) {
	var req AppleSignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	// Verify Apple token
	appleUser, err := h.appleVerifier.VerifyIDToken(c.Request.Context(), req.IDToken)
	if err != nil {
		respondUnauthorized(c, "Invalid Apple token")
		return
	}

	// Parse name from user info (only on first sign in)
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

// ========== Token & User ==========

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshToken refreshes access token
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	// Check if token is invalidated (logged out)
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

// GetCurrentUser returns current user
// GET /api/v1/auth/me
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

// Logout handles logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	// Validate the refresh token to get claims
	claims, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		// Even if token is invalid/expired, return success (user is logged out anyway)
		respondMessage(c, "Logged out successfully")
		return
	}

	// Invalidate the refresh token (ignore error, logout always succeeds)
	_ = h.tokenRepo.Invalidate(c.Request.Context(), req.RefreshToken, claims.UserID, claims.ExpiresAt.Time)
	respondMessage(c, "Logged out successfully")
}

// DeleteAccount deletes user account
// DELETE /api/v1/auth/account
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
