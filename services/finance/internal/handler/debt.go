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

type DebtHandler struct {
	debts *repository.DebtRepository
}

func NewDebtHandler(d *repository.DebtRepository) *DebtHandler {
	return &DebtHandler{debts: d}
}

type debtDTO struct {
	ID             string  `json:"id"`
	Counterparty   string  `json:"counterparty"`
	Direction      string  `json:"direction"`
	Amount         float64 `json:"amount"`
	OriginalAmount float64 `json:"original_amount"`
	Note           *string `json:"note,omitempty"`
	Settled        bool    `json:"settled"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

func toDebtDTO(d *domain.Debt) debtDTO {
	return debtDTO{
		ID:             d.ID.String(),
		Counterparty:   d.Counterparty,
		Direction:      string(d.Direction),
		Amount:         d.Amount,
		OriginalAmount: d.OriginalAmount,
		Note:           d.Note,
		Settled:        d.Settled,
		CreatedAt:      d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      d.UpdatedAt.Format(time.RFC3339),
	}
}

func (h *DebtHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var settledOnly *bool
	if s := c.Query("settled"); s == "true" {
		v := true
		settledOnly = &v
	} else if s == "false" {
		v := false
		settledOnly = &v
	}

	debts, err := h.debts.List(c.Request.Context(), userID, settledOnly)
	if err != nil {
		slog.Error("list debts", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to list debts")
		return
	}
	dtos := make([]debtDTO, len(debts))
	for i, d := range debts {
		dtos[i] = toDebtDTO(d)
	}
	respondOK(c, dtos)
}

type createDebtRequest struct {
	Counterparty string  `json:"counterparty" binding:"required"`
	Direction    string  `json:"direction"    binding:"required,oneof=i_owe they_owe"`
	Amount       float64 `json:"amount"       binding:"required,gt=0"`
	Note         *string `json:"note,omitempty"`
}

func (h *DebtHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	var req createDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	d := &domain.Debt{
		UserID:       userID,
		Counterparty: req.Counterparty,
		Direction:    domain.DebtDirection(req.Direction),
		Amount:       req.Amount,
		Note:         req.Note,
	}
	if err := h.debts.Create(c.Request.Context(), d); err != nil {
		slog.Error("create debt", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to create debt")
		return
	}
	respondCreated(c, toDebtDTO(d))
}

type patchDebtRequest struct {
	ReduceBy *float64 `json:"reduce_by,omitempty"`
	Settle   bool     `json:"settle,omitempty"`
}

func (h *DebtHandler) Patch(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid debt ID")
		return
	}
	var req patchDebtRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	var d *domain.Debt
	if req.Settle {
		d, err = h.debts.Settle(c.Request.Context(), id, userID)
	} else if req.ReduceBy != nil && *req.ReduceBy > 0 {
		d, err = h.debts.Patch(c.Request.Context(), id, userID, *req.ReduceBy)
	} else {
		respondError(c, http.StatusBadRequest, codeValidation, "Provide reduce_by or settle=true")
		return
	}

	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Debt not found")
		return
	} else if err != nil {
		slog.Error("patch debt", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update debt")
		return
	}
	respondOK(c, toDebtDTO(d))
}

func (h *DebtHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid debt ID")
		return
	}
	if err := h.debts.Delete(c.Request.Context(), id, userID); errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Debt not found")
		return
	} else if err != nil {
		slog.Error("delete debt", "err", err)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete debt")
		return
	}
	respondOK(c, gin.H{"deleted": true})
}
