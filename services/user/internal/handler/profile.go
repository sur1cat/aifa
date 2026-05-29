package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/sur1cat/aifa/user-service/internal/domain"
	"github.com/sur1cat/aifa/user-service/internal/middleware"
	"github.com/sur1cat/aifa/user-service/internal/repository"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profiles *repository.ProfileRepository
}

func NewProfileHandler(r *repository.ProfileRepository) *ProfileHandler {
	return &ProfileHandler{profiles: r}
}

func (h *ProfileHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Not authenticated")
		return
	}
	p, err := h.profiles.GetByID(c.Request.Context(), userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Profile not found")
		return
	}
	if err != nil {
		slog.Error("get profile", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load profile")
		return
	}
	respondOK(c, p)
}

type updateRequest struct {
	Name      *string `json:"name"`
	AvatarURL *string `json:"avatar_url"`
	Locale    *string `json:"locale"`
	Timezone  *string `json:"timezone"`
}

func (h *ProfileHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, codeUnauthorized, "Not authenticated")
		return
	}
	var req updateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	p, err := h.profiles.Update(c.Request.Context(), userID, repository.UpdateInput{
		Name:      req.Name,
		AvatarURL: req.AvatarURL,
		Locale:    req.Locale,
		Timezone:  req.Timezone,
	})
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Profile not found")
		return
	}
	if err != nil {
		slog.Error("update profile", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update profile")
		return
	}
	respondOK(c, p)
}
