package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"habitflow/internal/domain"
	"habitflow/internal/repository"
)

type SavingsGoalHandler struct {
	repo *repository.SavingsGoalRepository
}

func NewSavingsGoalHandler(repo *repository.SavingsGoalRepository) *SavingsGoalHandler {
	return &SavingsGoalHandler{repo: repo}
}

type SetSavingsGoalRequest struct {
	MonthlyTarget float64 `json:"monthlyTarget" binding:"required,gt=0"`
}

func (h *SavingsGoalHandler) GetSavingsGoal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	goal, err := h.repo.Get(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}

	if goal == nil {
		c.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}

	income, expenses, savings, err := h.repo.GetCurrentSavings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}

	progress := 0.0
	if goal.MonthlyTarget > 0 {
		progress = savings / goal.MonthlyTarget
		if progress > 1 {
			progress = 1
		}
		if progress < 0 {
			progress = 0
		}
	}

	response := domain.SavingsGoalWithProgress{
		ID:              goal.ID,
		MonthlyTarget:   goal.MonthlyTarget,
		CurrentSavings:  savings,
		MonthlyIncome:   income,
		MonthlyExpenses: expenses,
		Progress:        progress,
		CreatedAt:       goal.CreatedAt,
		UpdatedAt:       goal.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *SavingsGoalHandler) SetSavingsGoal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req SetSavingsGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	goal := &domain.SavingsGoal{
		UserID:        userID,
		MonthlyTarget: req.MonthlyTarget,
	}

	if err := h.repo.Upsert(c.Request.Context(), goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}

	income, expenses, savings, err := h.repo.GetCurrentSavings(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}

	progress := 0.0
	if goal.MonthlyTarget > 0 {
		progress = savings / goal.MonthlyTarget
		if progress > 1 {
			progress = 1
		}
		if progress < 0 {
			progress = 0
		}
	}

	response := domain.SavingsGoalWithProgress{
		ID:              goal.ID,
		MonthlyTarget:   goal.MonthlyTarget,
		CurrentSavings:  savings,
		MonthlyIncome:   income,
		MonthlyExpenses: expenses,
		Progress:        progress,
		CreatedAt:       goal.CreatedAt,
		UpdatedAt:       goal.UpdatedAt,
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

func (h *SavingsGoalHandler) DeleteSavingsGoal(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	if err := h.repo.Delete(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "DB_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"deleted": true}})
}
