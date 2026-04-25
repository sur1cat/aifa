package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/sur1cat/aifa/finance-service/internal/domain"
	"github.com/sur1cat/aifa/finance-service/internal/events"
	"github.com/sur1cat/aifa/finance-service/internal/middleware"
	"github.com/sur1cat/aifa/finance-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var labelsRu = map[string]string{
	"food": "Продукты", "cafe": "Кафе и рестораны", "transport": "Транспорт",
	"health": "Здоровье", "entertainment": "Развлечения", "utilities": "Коммунальные услуги",
	"shopping": "Покупки", "education": "Образование", "travel": "Путешествия",
	"transfer": "Переводы",
}

type BudgetHandler struct {
	budgets *repository.BudgetRepository
	pub     *events.Publisher
}

func NewBudgetHandler(b *repository.BudgetRepository, pub *events.Publisher) *BudgetHandler {
	return &BudgetHandler{budgets: b, pub: pub}
}

type budgetDTO struct {
	ID           string  `json:"id"`
	Category     string  `json:"category"`
	LabelRu      string  `json:"label_ru"`
	MonthlyLimit float64 `json:"monthly_limit"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

func toBudgetDTO(b *domain.Budget) budgetDTO {
	return budgetDTO{
		ID:           b.ID.String(),
		Category:     b.Category,
		LabelRu:      labelsRu[b.Category],
		MonthlyLimit: b.MonthlyLimit,
		CreatedAt:    b.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    b.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *BudgetHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	bs, err := h.budgets.List(c.Request.Context(), userID)
	if err != nil {
		slog.Error("list budgets", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to list budgets")
		return
	}
	dtos := make([]budgetDTO, len(bs))
	for i, b := range bs {
		dtos[i] = toBudgetDTO(b)
	}
	respondOK(c, dtos)
}

type upsertBudgetRequest struct {
	Category     string  `json:"category"      binding:"required"`
	MonthlyLimit float64 `json:"monthly_limit" binding:"required,gt=0"`
}

func (h *BudgetHandler) Upsert(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	var req upsertBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	b := &domain.Budget{UserID: userID, Category: req.Category, MonthlyLimit: req.MonthlyLimit}
	if err := h.budgets.Create(c.Request.Context(), b); err != nil {
		slog.Error("upsert budget", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to save budget")
		return
	}
	respondCreated(c, toBudgetDTO(b))
}

func (h *BudgetHandler) Update(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid budget ID")
		return
	}
	var req struct {
		MonthlyLimit float64 `json:"monthly_limit" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	b := &domain.Budget{ID: id, UserID: userID, MonthlyLimit: req.MonthlyLimit}
	if err := h.budgets.Update(c.Request.Context(), b); errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Budget not found")
		return
	} else if err != nil {
		slog.Error("update budget", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update budget")
		return
	}
	respondOK(c, gin.H{"updated": true})
}

func (h *BudgetHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid budget ID")
		return
	}
	if err := h.budgets.Delete(c.Request.Context(), id, userID); errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Budget not found")
		return
	} else if err != nil {
		slog.Error("delete budget", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete budget")
		return
	}
	respondOK(c, gin.H{"deleted": true})
}

