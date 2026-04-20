package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"habitflow/internal/repository"
)

type PushHandler struct {
	tokenRepo *repository.DeviceTokenRepository
}

func NewPushHandler(tokenRepo *repository.DeviceTokenRepository) *PushHandler {
	return &PushHandler{tokenRepo: tokenRepo}
}

type RegisterTokenRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform"`
}

// RegisterToken registers a device token for push notifications
func (h *PushHandler) RegisterToken(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	platform := req.Platform
	if platform == "" {
		platform = "ios"
	}

	err := h.tokenRepo.Register(c.Request.Context(), userID, req.Token, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to register token"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "registered"}})
}

// UnregisterToken removes a device token
func (h *PushHandler) UnregisterToken(c *gin.Context) {
	var req RegisterTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	err := h.tokenRepo.Unregister(c.Request.Context(), req.Token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to unregister token"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"status": "unregistered"}})
}
