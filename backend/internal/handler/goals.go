package handler

import (
	"net/http"
	"time"

	"habitflow/internal/domain"
	"habitflow/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GoalHandler struct {
	goalRepo *repository.GoalRepository
}

func NewGoalHandler(goalRepo *repository.GoalRepository) *GoalHandler {
	return &GoalHandler{goalRepo: goalRepo}
}

// Request/Response types
type CreateGoalRequest struct {
	Title       string  `json:"title" binding:"required"`
	Icon        string  `json:"icon"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
	Deadline    *string `json:"deadline"` // ISO8601 format
}

type UpdateGoalRequest struct {
	Title       string  `json:"title"`
	Icon        string  `json:"icon"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
	Deadline    *string `json:"deadline"`
	ArchivedAt  *string `json:"archived_at"`
}

type GoalResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Icon        string  `json:"icon"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
	Deadline    *string `json:"deadline"`
	CreatedAt   string  `json:"created_at"`
	ArchivedAt  *string `json:"archived_at"`
}

func toGoalResponse(g *domain.Goal) *GoalResponse {
	var deadline *string
	if g.Deadline != nil {
		formatted := g.Deadline.Format("2006-01-02T15:04:05Z07:00")
		deadline = &formatted
	}

	var archivedAt *string
	if g.ArchivedAt != nil {
		formatted := g.ArchivedAt.Format("2006-01-02T15:04:05Z07:00")
		archivedAt = &formatted
	}

	return &GoalResponse{
		ID:          g.ID.String(),
		Title:       g.Title,
		Icon:        g.Icon,
		TargetValue: g.TargetValue,
		Unit:        g.Unit,
		Deadline:    deadline,
		CreatedAt:   g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		ArchivedAt:  archivedAt,
	}
}

// ListGoals returns all goals for the current user
// GET /api/v1/goals
func (h *GoalHandler) ListGoals(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	goals, err := h.goalRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get goals"},
		})
		return
	}

	response := make([]*GoalResponse, len(goals))
	for i, goal := range goals {
		response[i] = toGoalResponse(goal)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// CreateGoal creates a new goal
// POST /api/v1/goals
func (h *GoalHandler) CreateGoal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	// Set defaults
	icon := req.Icon
	if icon == "" {
		icon = "🎯"
	}

	goal := &domain.Goal{
		UserID:      userID,
		Title:       req.Title,
		Icon:        icon,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
	}

	// Parse deadline if provided
	if req.Deadline != nil && *req.Deadline != "" {
		parsedTime, err := time.Parse(time.RFC3339, *req.Deadline)
		if err == nil {
			goal.Deadline = &parsedTime
		}
	}

	if err := h.goalRepo.Create(c.Request.Context(), goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create goal"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toGoalResponse(goal)})
}

// GetGoal returns a single goal
// GET /api/v1/goals/:id
func (h *GoalHandler) GetGoal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_ID", "message": "Invalid goal ID"},
		})
		return
	}

	// Verify ownership
	owns, err := h.goalRepo.VerifyOwnership(c.Request.Context(), goalID, userID)
	if err != nil || !owns {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Goal not found"},
		})
		return
	}

	goal, err := h.goalRepo.GetByID(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Goal not found"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGoalResponse(goal)})
}

// UpdateGoal updates a goal
// PUT /api/v1/goals/:id
func (h *GoalHandler) UpdateGoal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_ID", "message": "Invalid goal ID"},
		})
		return
	}

	// Verify ownership
	owns, err := h.goalRepo.VerifyOwnership(c.Request.Context(), goalID, userID)
	if err != nil || !owns {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Goal not found"},
		})
		return
	}

	var req UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	goal, err := h.goalRepo.GetByID(c.Request.Context(), goalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Goal not found"},
		})
		return
	}

	// Update fields if provided
	if req.Title != "" {
		goal.Title = req.Title
	}
	if req.Icon != "" {
		goal.Icon = req.Icon
	}
	if req.TargetValue != nil {
		goal.TargetValue = req.TargetValue
	}
	if req.Unit != nil {
		goal.Unit = req.Unit
	}
	// Handle deadline - can be set or cleared
	if req.Deadline != nil {
		if *req.Deadline == "" {
			goal.Deadline = nil
		} else {
			parsedTime, err := time.Parse(time.RFC3339, *req.Deadline)
			if err == nil {
				goal.Deadline = &parsedTime
			}
		}
	}
	// Handle archived_at - can be set or cleared
	if req.ArchivedAt != nil {
		if *req.ArchivedAt == "" {
			goal.ArchivedAt = nil
		} else {
			parsedTime, err := time.Parse(time.RFC3339, *req.ArchivedAt)
			if err == nil {
				goal.ArchivedAt = &parsedTime
			}
		}
	}

	if err := h.goalRepo.Update(c.Request.Context(), goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update goal"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toGoalResponse(goal)})
}

// DeleteGoal deletes a goal
// DELETE /api/v1/goals/:id
func (h *GoalHandler) DeleteGoal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_ID", "message": "Invalid goal ID"},
		})
		return
	}

	// Verify ownership
	owns, err := h.goalRepo.VerifyOwnership(c.Request.Context(), goalID, userID)
	if err != nil || !owns {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Goal not found"},
		})
		return
	}

	if err := h.goalRepo.Delete(c.Request.Context(), goalID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to delete goal"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "Goal deleted successfully"}})
}
