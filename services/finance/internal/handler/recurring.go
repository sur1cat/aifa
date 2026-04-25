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

type RecurringHandler struct {
	recurring *repository.RecurringRepository
	tx        *repository.TransactionRepository
}

func NewRecurringHandler(r *repository.RecurringRepository, tx *repository.TransactionRepository) *RecurringHandler {
	return &RecurringHandler{recurring: r, tx: tx}
}

type recurringDTO struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Amount            float64 `json:"amount"`
	Type              string  `json:"type"`
	Category          string  `json:"category"`
	Frequency         string  `json:"frequency"`
	StartDate         string  `json:"start_date"`
	NextDate          string  `json:"next_date"`
	EndDate           *string `json:"end_date,omitempty"`
	RemainingPayments *int    `json:"remaining_payments,omitempty"`
	IsActive          bool    `json:"is_active"`
	CreatedAt         string  `json:"created_at"`
}

func toRecurringDTO(r *domain.Recurring) recurringDTO {
	return recurringDTO{
		ID:                r.ID.String(),
		Title:             r.Title,
		Amount:            r.Amount,
		Type:              string(r.Type),
		Category:          r.Category,
		Frequency:         string(r.Frequency),
		StartDate:         r.StartDate,
		NextDate:          r.NextDate,
		EndDate:           r.EndDate,
		RemainingPayments: r.RemainingPayments,
		IsActive:          r.IsActive,
		CreatedAt:         r.CreatedAt.Format(time.RFC3339),
	}
}

var validFrequencies = map[domain.Frequency]bool{
	domain.FreqWeekly:    true,
	domain.FreqBiweekly:  true,
	domain.FreqMonthly:   true,
	domain.FreqQuarterly: true,
	domain.FreqYearly:    true,
}

type createRecurringRequest struct {
	Title             string  `json:"title" binding:"required"`
	Amount            float64 `json:"amount" binding:"required"`
	Type              string  `json:"type" binding:"required,oneof=income expense"`
	Category          string  `json:"category"`
	Frequency         string  `json:"frequency" binding:"required"`
	StartDate         string  `json:"start_date"`
	EndDate           *string `json:"end_date"`
	RemainingPayments *int    `json:"remaining_payments"`
}

func (h *RecurringHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	rows, err := h.recurring.ListByUser(c.Request.Context(), userID)
	if err != nil {
		slog.Error("list recurring", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get recurring transactions")
		return
	}
	dtos := make([]recurringDTO, len(rows))
	for i, r := range rows {
		dtos[i] = toRecurringDTO(r)
	}
	respondOK(c, dtos)
}

func (h *RecurringHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req createRecurringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	freq := domain.Frequency(req.Frequency)
	if !validFrequencies[freq] {
		respondError(c, http.StatusBadRequest, codeValidation, "invalid frequency")
		return
	}

	startDate := req.StartDate
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}

	x := &domain.Recurring{
		UserID:            userID,
		Title:             req.Title,
		Amount:            req.Amount,
		Type:              domain.TransactionType(req.Type),
		Category:          req.Category,
		Frequency:         freq,
		StartDate:         startDate,
		NextDate:          startDate,
		EndDate:           req.EndDate,
		RemainingPayments: req.RemainingPayments,
		IsActive:          true,
	}
	if err := h.recurring.Create(c.Request.Context(), x); err != nil {
		slog.Error("create recurring", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to create recurring transaction")
		return
	}
	respondCreated(c, toRecurringDTO(x))
}

func (h *RecurringHandler) Get(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid recurring transaction ID")
		return
	}
	x, err := h.recurring.GetOwnedByID(c.Request.Context(), id, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Recurring transaction not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load recurring transaction")
		return
	}
	respondOK(c, toRecurringDTO(x))
}

type updateRecurringRequest struct {
	Title             string  `json:"title"`
	Amount            float64 `json:"amount"`
	Type              string  `json:"type"`
	Category          string  `json:"category"`
	Frequency         string  `json:"frequency"`
	StartDate         string  `json:"start_date"`
	NextDate          string  `json:"next_date"`
	EndDate           *string `json:"end_date"`
	RemainingPayments *int    `json:"remaining_payments"`
	IsActive          *bool   `json:"is_active"`
}

func (h *RecurringHandler) Update(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid recurring transaction ID")
		return
	}

	var req updateRecurringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	x, err := h.recurring.GetOwnedByID(c.Request.Context(), id, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Recurring transaction not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load recurring transaction")
		return
	}

	if err := applyRecurringUpdate(x, req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if err := h.recurring.Update(c.Request.Context(), x); err != nil {
		slog.Error("update recurring", "err", err, "id", x.ID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update recurring transaction")
		return
	}
	respondOK(c, toRecurringDTO(x))
}

func applyRecurringUpdate(x *domain.Recurring, req updateRecurringRequest) error {
	if req.Title != "" {
		x.Title = req.Title
	}
	if req.Amount != 0 {
		x.Amount = req.Amount
	}
	if req.Type != "" {
		x.Type = domain.TransactionType(req.Type)
	}
	if req.Category != "" {
		x.Category = req.Category
	}
	if req.Frequency != "" {
		f := domain.Frequency(req.Frequency)
		if !validFrequencies[f] {
			return errors.New("invalid frequency")
		}
		x.Frequency = f
	}
	if req.StartDate != "" {
		x.StartDate = req.StartDate
	}
	if req.NextDate != "" {
		x.NextDate = req.NextDate
	}
	if req.EndDate != nil {
		x.EndDate = req.EndDate
	}
	if req.RemainingPayments != nil {
		x.RemainingPayments = req.RemainingPayments
	}
	if req.IsActive != nil {
		x.IsActive = *req.IsActive
	}
	return nil
}

func (h *RecurringHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid recurring transaction ID")
		return
	}
	err = h.recurring.Delete(c.Request.Context(), id, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Recurring transaction not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete recurring transaction")
		return
	}
	respondMessage(c, "Recurring transaction deleted successfully")
}

type projectionDTO struct {
	MonthlyIncome  float64 `json:"monthly_income"`
	MonthlyExpense float64 `json:"monthly_expense"`
	MonthlyNet     float64 `json:"monthly_net"`
}

func (h *RecurringHandler) Projection(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	income, expense, err := h.recurring.MonthlyProjection(c.Request.Context(), userID)
	if err != nil {
		slog.Error("projection", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get projection")
		return
	}
	respondOK(c, projectionDTO{MonthlyIncome: income, MonthlyExpense: expense, MonthlyNet: income - expense})
}

type processResultDTO struct {
	Processed           int `json:"processed"`
	TransactionsCreated int `json:"transactions_created"`
}

// Process advances every due recurring row forward one occurrence per call,
// inserting a concrete transaction for each advance. Protected against
// duplicate runs via AdvanceIfDue (CAS on next_date).
func (h *RecurringHandler) Process(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	ctx := c.Request.Context()
	today := time.Now().Format("2006-01-02")

	rows, err := h.recurring.ListByUser(ctx, userID)
	if err != nil {
		slog.Error("process recurring: list", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get recurring transactions")
		return
	}

	created := 0
	for _, x := range rows {
		if !x.IsActive {
			continue
		}
		if x.EndDate != nil && *x.EndDate < today {
			x.IsActive = false
			_ = h.recurring.Update(ctx, x)
			continue
		}

		for x.NextDate <= today {
			currentDate := x.NextDate
			newNextDate := x.NextDateFrom(currentDate)

			newRemaining := x.RemainingPayments
			newActive := x.IsActive
			if x.RemainingPayments != nil {
				r := *x.RemainingPayments - 1
				newRemaining = &r
				if r <= 0 {
					newActive = false
				}
			}
			if x.EndDate != nil && newNextDate > *x.EndDate {
				newActive = false
			}

			advanced, err := h.recurring.AdvanceIfDue(ctx, x.ID, currentDate, newNextDate, newRemaining, newActive)
			if err != nil || !advanced {
				break
			}

			t := &domain.Transaction{
				UserID:   userID,
				Title:    x.Title,
				Amount:   x.Amount,
				Type:     x.Type,
				Category: x.Category,
				Date:     currentDate,
			}
			if err := h.tx.Create(ctx, t); err != nil {
				slog.Error("process recurring: create tx", "err", err, "recurring_id", x.ID)
				continue
			}
			created++

			x.NextDate = newNextDate
			x.RemainingPayments = newRemaining
			x.IsActive = newActive
			if !newActive {
				break
			}
		}
	}

	respondOK(c, processResultDTO{Processed: len(rows), TransactionsCreated: created})
}
