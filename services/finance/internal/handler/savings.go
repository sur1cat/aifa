package handler

import (
	"log/slog"
	"net/http"

	"github.com/sur1cat/aifa/finance-service/internal/domain"
	"github.com/sur1cat/aifa/finance-service/internal/middleware"
	"github.com/sur1cat/aifa/finance-service/internal/repository"

	"github.com/gin-gonic/gin"
)

type SavingsHandler struct {
	savings *repository.SavingsRepository
	tx      *repository.TransactionRepository
}

func NewSavingsHandler(s *repository.SavingsRepository, tx *repository.TransactionRepository) *SavingsHandler {
	return &SavingsHandler{savings: s, tx: tx}
}

type savingsDTO struct {
	ID              string  `json:"id"`
	MonthlyTarget   float64 `json:"monthlyTarget"`
	CurrentSavings  float64 `json:"currentSavings"`
	MonthlyIncome   float64 `json:"monthlyIncome"`
	MonthlyExpenses float64 `json:"monthlyExpenses"`
	Progress        float64 `json:"progress"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func progressShare(savings, target float64) float64 {
	if target <= 0 {
		return 0
	}
	p := savings / target
	switch {
	case p > 1:
		return 1
	case p < 0:
		return 0
	default:
		return p
	}
}

func (h *SavingsHandler) Get(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	goal, err := h.savings.Get(c.Request.Context(), userID)
	if err != nil {
		slog.Error("get savings goal", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load savings goal")
		return
	}
	if goal == nil {
		respondOK(c, nil)
		return
	}

	income, expenses, err := h.tx.SumCurrentMonth(c.Request.Context(), userID)
	if err != nil {
		slog.Error("sum current month", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to compute savings progress")
		return
	}
	respondOK(c, buildSavingsDTO(goal, income, expenses))
}

type setSavingsRequest struct {
	MonthlyTarget float64 `json:"monthlyTarget" binding:"required,gt=0"`
}

func (h *SavingsHandler) Set(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req setSavingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	goal := &domain.SavingsGoal{UserID: userID, MonthlyTarget: req.MonthlyTarget}
	if err := h.savings.Upsert(c.Request.Context(), goal); err != nil {
		slog.Error("upsert savings goal", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to set savings goal")
		return
	}

	income, expenses, err := h.tx.SumCurrentMonth(c.Request.Context(), userID)
	if err != nil {
		slog.Error("sum current month", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to compute savings progress")
		return
	}
	respondOK(c, buildSavingsDTO(goal, income, expenses))
}

func (h *SavingsHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	if err := h.savings.Delete(c.Request.Context(), userID); err != nil {
		slog.Error("delete savings goal", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete savings goal")
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

func buildSavingsDTO(g *domain.SavingsGoal, income, expenses float64) savingsDTO {
	savings := income - expenses
	return savingsDTO{
		ID:              g.ID.String(),
		MonthlyTarget:   g.MonthlyTarget,
		CurrentSavings:  savings,
		MonthlyIncome:   income,
		MonthlyExpenses: expenses,
		Progress:        progressShare(savings, g.MonthlyTarget),
		CreatedAt:       g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:       g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
