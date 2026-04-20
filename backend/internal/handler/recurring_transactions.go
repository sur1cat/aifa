package handler

import (
	"net/http"
	"time"

	"habitflow/internal/domain"
	"habitflow/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RecurringTransactionHandler struct {
	rtRepo *repository.RecurringTransactionRepository
	txRepo *repository.TransactionRepository
}

func NewRecurringTransactionHandler(rtRepo *repository.RecurringTransactionRepository, txRepo *repository.TransactionRepository) *RecurringTransactionHandler {
	return &RecurringTransactionHandler{rtRepo: rtRepo, txRepo: txRepo}
}

// Request/Response types
type CreateRecurringRequest struct {
	Title             string  `json:"title" binding:"required"`
	Amount            float64 `json:"amount" binding:"required"`
	Type              string  `json:"type" binding:"required"`
	Category          string  `json:"category"`
	Frequency         string  `json:"frequency" binding:"required"`
	StartDate         string  `json:"start_date"`
	EndDate           *string `json:"end_date"`
	RemainingPayments *int    `json:"remaining_payments"`
}

type UpdateRecurringRequest struct {
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

type RecurringResponse struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Amount            float64 `json:"amount"`
	Type              string  `json:"type"`
	Category          string  `json:"category"`
	Frequency         string  `json:"frequency"`
	StartDate         string  `json:"start_date"`
	NextDate          string  `json:"next_date"`
	EndDate           *string `json:"end_date"`
	RemainingPayments *int    `json:"remaining_payments"`
	IsActive          bool    `json:"is_active"`
	CreatedAt         string  `json:"created_at"`
}

type ProjectionResponse struct {
	MonthlyIncome  float64 `json:"monthly_income"`
	MonthlyExpense float64 `json:"monthly_expense"`
	MonthlyNet     float64 `json:"monthly_net"`
}

func toRecurringResponse(rt *domain.RecurringTransaction) *RecurringResponse {
	return &RecurringResponse{
		ID:                rt.ID.String(),
		Title:             rt.Title,
		Amount:            rt.Amount,
		Type:              string(rt.Type),
		Category:          rt.Category,
		Frequency:         string(rt.Frequency),
		StartDate:         rt.StartDate,
		NextDate:          rt.NextDate,
		EndDate:           rt.EndDate,
		RemainingPayments: rt.RemainingPayments,
		IsActive:          rt.IsActive,
		CreatedAt:         rt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ListRecurringTransactions returns all recurring transactions for the current user
// GET /api/v1/recurring-transactions
func (h *RecurringTransactionHandler) ListRecurringTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	transactions, err := h.rtRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get recurring transactions"},
		})
		return
	}

	response := make([]*RecurringResponse, len(transactions))
	for i, rt := range transactions {
		response[i] = toRecurringResponse(rt)
	}

	c.JSON(http.StatusOK, gin.H{"data": response})
}

// CreateRecurringTransaction creates a new recurring transaction
// POST /api/v1/recurring-transactions
func (h *RecurringTransactionHandler) CreateRecurringTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req CreateRecurringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	// Validate type
	txType := domain.TransactionType(req.Type)
	if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid transaction type"},
		})
		return
	}

	// Validate frequency
	freq := domain.RecurrenceFrequency(req.Frequency)
	validFrequencies := map[domain.RecurrenceFrequency]bool{
		domain.FrequencyWeekly:    true,
		domain.FrequencyBiweekly:  true,
		domain.FrequencyMonthly:   true,
		domain.FrequencyQuarterly: true,
		domain.FrequencyYearly:    true,
	}
	if !validFrequencies[freq] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": "Invalid frequency"},
		})
		return
	}

	// Set defaults
	startDate := req.StartDate
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}

	rt := &domain.RecurringTransaction{
		UserID:            userID,
		Title:             req.Title,
		Amount:            req.Amount,
		Type:              txType,
		Category:          req.Category,
		Frequency:         freq,
		StartDate:         startDate,
		NextDate:          startDate,
		EndDate:           req.EndDate,
		RemainingPayments: req.RemainingPayments,
		IsActive:          true,
	}

	if err := h.rtRepo.Create(c.Request.Context(), rt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to create recurring transaction"},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": toRecurringResponse(rt)})
}

// GetRecurringTransaction returns a single recurring transaction
// GET /api/v1/recurring-transactions/:id
func (h *RecurringTransactionHandler) GetRecurringTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	rtID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_ID", "message": "Invalid recurring transaction ID"},
		})
		return
	}

	// Verify ownership
	owns, err := h.rtRepo.VerifyOwnership(c.Request.Context(), rtID, userID)
	if err != nil || !owns {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Recurring transaction not found"},
		})
		return
	}

	rt, err := h.rtRepo.GetByID(c.Request.Context(), rtID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Recurring transaction not found"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toRecurringResponse(rt)})
}

// UpdateRecurringTransaction updates a recurring transaction
// PUT /api/v1/recurring-transactions/:id
func (h *RecurringTransactionHandler) UpdateRecurringTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	rtID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_ID", "message": "Invalid recurring transaction ID"},
		})
		return
	}

	// Verify ownership
	owns, err := h.rtRepo.VerifyOwnership(c.Request.Context(), rtID, userID)
	if err != nil || !owns {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Recurring transaction not found"},
		})
		return
	}

	var req UpdateRecurringRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "VALIDATION_ERROR", "message": err.Error()},
		})
		return
	}

	rt, err := h.rtRepo.GetByID(c.Request.Context(), rtID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Recurring transaction not found"},
		})
		return
	}

	// Update fields if provided
	if req.Title != "" {
		rt.Title = req.Title
	}
	if req.Amount != 0 {
		rt.Amount = req.Amount
	}
	if req.Type != "" {
		rt.Type = domain.TransactionType(req.Type)
	}
	if req.Category != "" {
		rt.Category = req.Category
	}
	if req.Frequency != "" {
		rt.Frequency = domain.RecurrenceFrequency(req.Frequency)
	}
	if req.StartDate != "" {
		rt.StartDate = req.StartDate
	}
	if req.NextDate != "" {
		rt.NextDate = req.NextDate
	}
	if req.EndDate != nil {
		rt.EndDate = req.EndDate
	}
	if req.RemainingPayments != nil {
		rt.RemainingPayments = req.RemainingPayments
	}
	if req.IsActive != nil {
		rt.IsActive = *req.IsActive
	}

	if err := h.rtRepo.Update(c.Request.Context(), rt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to update recurring transaction"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": toRecurringResponse(rt)})
}

// DeleteRecurringTransaction deletes a recurring transaction
// DELETE /api/v1/recurring-transactions/:id
func (h *RecurringTransactionHandler) DeleteRecurringTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	rtID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{"code": "INVALID_ID", "message": "Invalid recurring transaction ID"},
		})
		return
	}

	// Verify ownership
	owns, err := h.rtRepo.VerifyOwnership(c.Request.Context(), rtID, userID)
	if err != nil || !owns {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{"code": "NOT_FOUND", "message": "Recurring transaction not found"},
		})
		return
	}

	if err := h.rtRepo.Delete(c.Request.Context(), rtID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to delete recurring transaction"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": gin.H{"message": "Recurring transaction deleted successfully"}})
}

// GetProjection returns monthly projection for recurring transactions
// GET /api/v1/recurring-transactions/projection
func (h *RecurringTransactionHandler) GetProjection(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	income, expense, err := h.rtRepo.GetMonthlyProjection(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get projection"},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ProjectionResponse{
		MonthlyIncome:  income,
		MonthlyExpense: expense,
		MonthlyNet:     income - expense,
	}})
}

// ProcessRecurringTransactions creates transactions from recurring ones that are due
// POST /api/v1/recurring-transactions/process
func (h *RecurringTransactionHandler) ProcessRecurringTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	today := time.Now().Format("2006-01-02")

	// Get all active recurring transactions for the user
	transactions, err := h.rtRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to get recurring transactions"},
		})
		return
	}

	var created int
	for _, rt := range transactions {
		if !rt.IsActive {
			continue
		}

		// Check if end_date has passed
		if rt.EndDate != nil && *rt.EndDate < today {
			rt.IsActive = false
			h.rtRepo.Update(c.Request.Context(), rt)
			continue
		}

		// Process all due dates up to today
		for rt.NextDate <= today {
			currentDate := rt.NextDate
			newNextDate := rt.CalculateNextDate()

			// Decrement remaining payments if set
			newRemainingPayments := rt.RemainingPayments
			newIsActive := rt.IsActive
			if rt.RemainingPayments != nil {
				remaining := *rt.RemainingPayments - 1
				newRemainingPayments = &remaining
				if remaining <= 0 {
					newIsActive = false
				}
			}

			// Check if next_date exceeds end_date
			if rt.EndDate != nil && newNextDate > *rt.EndDate {
				newIsActive = false
			}

			// Atomically update next_date (prevents race condition)
			// If another process already updated, this returns false
			updated, err := h.rtRepo.UpdateNextDateAtomic(
				c.Request.Context(),
				rt.ID,
				currentDate,
				newNextDate,
				newRemainingPayments,
				newIsActive,
			)
			if err != nil {
				continue // Skip on error
			}
			if !updated {
				// Another process already processed this date, skip
				break
			}

			// Create actual transaction only after successful atomic update
			tx := &domain.Transaction{
				UserID:   userID,
				Title:    rt.Title,
				Amount:   rt.Amount,
				Type:     rt.Type,
				Category: rt.Category,
				Date:     currentDate,
			}

			if err := h.txRepo.Create(c.Request.Context(), tx); err != nil {
				continue // Skip on error, try next
			}
			created++

			// Update local state for next iteration
			rt.NextDate = newNextDate
			rt.RemainingPayments = newRemainingPayments
			rt.IsActive = newIsActive

			// If no longer active, stop processing this recurring transaction
			if !rt.IsActive {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"processed":            len(transactions),
			"transactions_created": created,
		},
	})
}
