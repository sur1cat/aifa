package handler

import (
	"strconv"
	"strings"
	"time"

	"habitflow/internal/domain"
	"habitflow/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	txRepo *repository.TransactionRepository
}

func NewTransactionHandler(txRepo *repository.TransactionRepository) *TransactionHandler {
	return &TransactionHandler{txRepo: txRepo}
}

// Request/Response types
type CreateTransactionRequest struct {
	Title    string  `json:"title" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Type     string  `json:"type" binding:"required,oneof=income expense"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
}

type UpdateTransactionRequest struct {
	Title    string  `json:"title"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
}

type TransactionResponse struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Amount    float64 `json:"amount"`
	Type      string  `json:"type"`
	Category  string  `json:"category"`
	Date      string  `json:"date"`
	CreatedAt string  `json:"created_at"`
}

type SummaryResponse struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

var expenseTransactionCategories = map[string]struct{}{
	"food":          {},
	"transport":     {},
	"shopping":      {},
	"entertainment": {},
	"health":        {},
	"education":     {},
	"bills":         {},
	"gift":          {},
	"other":         {},
}

var incomeTransactionCategories = map[string]struct{}{
	"salary":     {},
	"freelance":  {},
	"investment": {},
	"gift":       {},
	"other":      {},
}

func toTransactionResponse(tx *domain.Transaction) *TransactionResponse {
	return &TransactionResponse{
		ID:        tx.ID.String(),
		Title:     tx.Title,
		Amount:    tx.Amount,
		Type:      string(tx.Type),
		Category:  tx.Category,
		Date:      tx.Date,
		CreatedAt: tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func normalizeTransactionCategory(category string) string {
	normalized := strings.ToLower(strings.TrimSpace(category))
	if normalized == "" {
		return "other"
	}
	return normalized
}

func categoryAllowedForType(txType domain.TransactionType, category string) bool {
	switch txType {
	case domain.TransactionTypeExpense:
		_, ok := expenseTransactionCategories[category]
		return ok
	case domain.TransactionTypeIncome:
		_, ok := incomeTransactionCategories[category]
		return ok
	default:
		return false
	}
}

func normalizeCategoryForType(txType domain.TransactionType, category string) string {
	normalized := normalizeTransactionCategory(category)
	if categoryAllowedForType(txType, normalized) {
		return normalized
	}
	return "other"
}

// ListTransactions returns all transactions for the current user
// GET /api/v1/transactions
func (h *TransactionHandler) ListTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	// Check for month query params
	yearStr := c.Query("year")
	monthStr := c.Query("month")

	// Pagination params with defaults
	limit := 50
	offset := 0
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	var transactions []*domain.Transaction
	var total int
	var err error

	if yearStr != "" && monthStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			respondValidationError(c, "Invalid year parameter")
			return
		}
		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			respondValidationError(c, "Invalid month parameter")
			return
		}
		transactions, total, err = h.txRepo.GetByUserIDAndMonth(c.Request.Context(), userID, year, month, limit, offset)
	} else {
		transactions, err = h.txRepo.GetByUserID(c.Request.Context(), userID)
		total = len(transactions)
	}

	if err != nil {
		respondInternalError(c, "Failed to get transactions")
		return
	}

	response := make([]*TransactionResponse, len(transactions))
	for i, tx := range transactions {
		response[i] = toTransactionResponse(tx)
	}

	respondPaginated(c, response, PaginationMeta{Limit: limit, Offset: offset, Total: total})
}

// CreateTransaction creates a new transaction
// POST /api/v1/transactions
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	// Validate type
	txType := domain.TransactionType(req.Type)
	if txType != domain.TransactionTypeIncome && txType != domain.TransactionTypeExpense {
		respondValidationError(c, "Invalid transaction type")
		return
	}

	// Set defaults
	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	tx := &domain.Transaction{
		UserID:   userID,
		Title:    req.Title,
		Amount:   req.Amount,
		Type:     txType,
		Category: normalizeCategoryForType(txType, req.Category),
		Date:     date,
	}

	if err := h.txRepo.Create(c.Request.Context(), tx); err != nil {
		respondInternalError(c, "Failed to create transaction")
		return
	}

	respondCreated(c, toTransactionResponse(tx))
}

// GetTransaction returns a single transaction
// GET /api/v1/transactions/:id
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid transaction ID")
		return
	}

	// Verify ownership
	owns, err := h.txRepo.VerifyOwnership(c.Request.Context(), txID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Transaction not found")
		return
	}

	tx, err := h.txRepo.GetByID(c.Request.Context(), txID)
	if err != nil {
		respondNotFound(c, "Transaction not found")
		return
	}

	respondOK(c, toTransactionResponse(tx))
}

// UpdateTransaction updates a transaction
// PUT /api/v1/transactions/:id
func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid transaction ID")
		return
	}

	// Verify ownership
	owns, err := h.txRepo.VerifyOwnership(c.Request.Context(), txID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Transaction not found")
		return
	}

	var req UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	tx, err := h.txRepo.GetByID(c.Request.Context(), txID)
	if err != nil {
		respondNotFound(c, "Transaction not found")
		return
	}

	// Update fields if provided
	if req.Title != "" {
		tx.Title = req.Title
	}
	if req.Amount != 0 {
		tx.Amount = req.Amount
	}
	if req.Type != "" {
		tx.Type = domain.TransactionType(req.Type)
	}
	if req.Category != "" {
		tx.Category = normalizeCategoryForType(tx.Type, req.Category)
	}
	if req.Date != "" {
		tx.Date = req.Date
	}
	tx.Category = normalizeCategoryForType(tx.Type, tx.Category)

	if err := h.txRepo.Update(c.Request.Context(), tx); err != nil {
		respondInternalError(c, "Failed to update transaction")
		return
	}

	respondOK(c, toTransactionResponse(tx))
}

// DeleteTransaction deletes a transaction
// DELETE /api/v1/transactions/:id
func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "Invalid transaction ID")
		return
	}

	// Verify ownership
	owns, err := h.txRepo.VerifyOwnership(c.Request.Context(), txID, userID)
	if err != nil || !owns {
		respondNotFound(c, "Transaction not found")
		return
	}

	if err := h.txRepo.Delete(c.Request.Context(), txID); err != nil {
		respondInternalError(c, "Failed to delete transaction")
		return
	}

	respondMessage(c, "Transaction deleted successfully")
}

// GetSummary returns income/expense summary for a month
// GET /api/v1/transactions/summary
func (h *TransactionHandler) GetSummary(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	// Get year and month from query params, default to current month
	now := time.Now()
	yearStr := c.DefaultQuery("year", strconv.Itoa(now.Year()))
	monthStr := c.DefaultQuery("month", strconv.Itoa(int(now.Month())))

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondValidationError(c, "Invalid year parameter")
		return
	}
	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 {
		respondValidationError(c, "Invalid month parameter")
		return
	}

	income, expense, err := h.txRepo.GetSummary(c.Request.Context(), userID, year, month)
	if err != nil {
		respondInternalError(c, "Failed to get summary")
		return
	}

	respondOK(c, SummaryResponse{
		Income:  income,
		Expense: expense,
		Balance: income - expense,
	})
}
