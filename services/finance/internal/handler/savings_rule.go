package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/sur1cat/aifa/finance-service/internal/domain"
	"github.com/sur1cat/aifa/finance-service/internal/middleware"
	"github.com/sur1cat/aifa/finance-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SavingsRuleHandler struct {
	rules *repository.SavingsRuleRepository
}

func NewSavingsRuleHandler(r *repository.SavingsRuleRepository) *SavingsRuleHandler {
	return &SavingsRuleHandler{rules: r}
}

type savingsRuleDTO struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Amount    float64 `json:"amount"`
	Period    *string `json:"period,omitempty"`
	GoalTitle *string `json:"goal_title,omitempty"`
	Active    bool    `json:"active"`
	CreatedAt string  `json:"created_at"`
}

func toRuleDTO(r *domain.SavingsRule) savingsRuleDTO {
	return savingsRuleDTO{
		ID:        r.ID.String(),
		Kind:      string(r.Kind),
		Amount:    r.Amount,
		Period:    r.Period,
		GoalTitle: r.GoalTitle,
		Active:    r.Active,
		CreatedAt: r.CreatedAt.Format(time.RFC3339),
	}
}

func (h *SavingsRuleHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	activeOnly := c.Query("active") != "false"
	rules, err := h.rules.List(c.Request.Context(), userID, activeOnly)
	if err != nil {
		slog.Error("list savings rules", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to list rules")
		return
	}
	dtos := make([]savingsRuleDTO, len(rules))
	for i, r := range rules {
		dtos[i] = toRuleDTO(r)
	}
	respondOK(c, dtos)
}

type createRuleRequest struct {
	Kind      string  `json:"kind"      binding:"required,oneof=monthly_savings on_income_savings spending_alert"`
	Amount    float64 `json:"amount"    binding:"required,gt=0"`
	Period    *string `json:"period,omitempty"`
	GoalTitle *string `json:"goal_title,omitempty"`
}

func (h *SavingsRuleHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	var req createRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	rule := &domain.SavingsRule{
		UserID:    userID,
		Kind:      domain.SavingsRuleKind(req.Kind),
		Amount:    req.Amount,
		Period:    req.Period,
		GoalTitle: req.GoalTitle,
	}
	if err := h.rules.Create(c.Request.Context(), rule); err != nil {
		slog.Error("create savings rule", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to create rule")
		return
	}
	respondCreated(c, toRuleDTO(rule))
}

func (h *SavingsRuleHandler) Deactivate(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid rule ID")
		return
	}
	if err := h.rules.Deactivate(c.Request.Context(), id, userID); errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Rule not found")
		return
	} else if err != nil {
		slog.Error("deactivate savings rule", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to deactivate rule")
		return
	}
	respondOK(c, gin.H{"deactivated": true})
}

func (h *SavingsRuleHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid rule ID")
		return
	}
	if err := h.rules.Delete(c.Request.Context(), id, userID); errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Rule not found")
		return
	} else if err != nil {
		slog.Error("delete savings rule", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete rule")
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

// DailySpent — сколько потрачено сегодня (для клиента).
func (h *SavingsRuleHandler) DailySpent(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	total, err := h.rules.DailySpent(c.Request.Context(), userID)
	if err != nil {
		slog.Error("daily spent", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get daily spent")
		return
	}
	respondOK(c, gin.H{"daily_spent": total})
}
