package handler

import (
	"log/slog"
	"net/http"

	"github.com/sur1cat/aifa/notification-service/internal/middleware"
	"github.com/sur1cat/aifa/notification-service/internal/repository"

	"github.com/gin-gonic/gin"
)

type PushHandler struct {
	tokens *repository.DeviceTokenRepository
}

func NewPushHandler(r *repository.DeviceTokenRepository) *PushHandler {
	return &PushHandler{tokens: r}
}

type registerRequest struct {
	Token    string `json:"token" binding:"required"`
	Platform string `json:"platform"`
}

func (h *PushHandler) Register(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	platform := req.Platform
	if platform == "" {
		platform = "ios"
	}

	if err := h.tokens.Register(c.Request.Context(), userID, req.Token, platform); err != nil {
		slog.Error("register push token", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to register token")
		return
	}
	respondOK(c, gin.H{"status": "registered"})
}

func (h *PushHandler) Unregister(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	if err := h.tokens.Unregister(c.Request.Context(), req.Token); err != nil {
		slog.Error("unregister push token", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to unregister token")
		return
	}
	respondOK(c, gin.H{"status": "unregistered"})
}
