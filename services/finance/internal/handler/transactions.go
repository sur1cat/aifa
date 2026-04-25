package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/sur1cat/aifa/finance-service/internal/domain"
	"github.com/sur1cat/aifa/finance-service/internal/events"
	"github.com/sur1cat/aifa/finance-service/internal/localai"
	"github.com/sur1cat/aifa/finance-service/internal/middleware"
	"github.com/sur1cat/aifa/finance-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	tx      *repository.TransactionRepository
	localAI *localai.Client
	budgets *repository.BudgetRepository
	pub     *events.Publisher
}

func NewTransactionHandler(r *repository.TransactionRepository, ai *localai.Client, b *repository.BudgetRepository, pub *events.Publisher) *TransactionHandler {
	return &TransactionHandler{tx: r, localAI: ai, budgets: b, pub: pub}
}

type txDTO struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Amount    float64 `json:"amount"`
	Type      string  `json:"type"`
	Category  string  `json:"category"`
	Date      string  `json:"date"`
	CreatedAt string  `json:"created_at"`
}

func toTxDTO(t *domain.Transaction) txDTO {
	return txDTO{
		ID:        t.ID.String(),
		Title:     t.Title,
		Amount:    t.Amount,
		Type:      string(t.Type),
		Category:  t.Category,
		Date:      t.Date,
		CreatedAt: t.CreatedAt.Format(time.RFC3339),
	}
}

func toTxDTOs(ts []*domain.Transaction) []txDTO {
	dtos := make([]txDTO, len(ts))
	for i, t := range ts {
		dtos[i] = toTxDTO(t)
	}
	return dtos
}

func (h *TransactionHandler) List(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	yearStr, monthStr := c.Query("year"), c.Query("month")
	limit, offset := paginationParams(c)

	if yearStr != "" && monthStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			respondError(c, http.StatusBadRequest, codeValidation, "invalid year")
			return
		}
		month, err := strconv.Atoi(monthStr)
		if err != nil || month < 1 || month > 12 {
			respondError(c, http.StatusBadRequest, codeValidation, "invalid month")
			return
		}
		txs, total, err := h.tx.ListByUserAndMonth(c.Request.Context(), userID, year, month, limit, offset)
		if err != nil {
			slog.Error("list transactions by month", "err", err, "user_id", userID)
			respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get transactions")
			return
		}
		respondPaginated(c, toTxDTOs(txs), paginationMeta{Limit: limit, Offset: offset, Total: total})
		return
	}

	txs, err := h.tx.ListByUser(c.Request.Context(), userID)
	if err != nil {
		slog.Error("list transactions", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get transactions")
		return
	}
	respondPaginated(c, toTxDTOs(txs), paginationMeta{Limit: limit, Offset: offset, Total: len(txs)})
}

type createTxRequest struct {
	Title    string  `json:"title" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Type     string  `json:"type" binding:"required,oneof=income expense"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
}

func (h *TransactionHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserID(c)

	var req createTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	date := req.Date
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	category := req.Category
	if category == "" && req.Type == "expense" {
		if cat, ok := h.localAI.CategorizeExpense(c.Request.Context(), req.Title); ok {
			category = cat
			slog.Info("auto-categorized transaction", "title", req.Title, "category", category)
		}
	}

	t := &domain.Transaction{
		UserID:   userID,
		Title:    req.Title,
		Amount:   req.Amount,
		Type:     domain.TransactionType(req.Type),
		Category: category,
		Date:     date,
	}
	if err := h.tx.Create(c.Request.Context(), t); err != nil {
		slog.Error("create transaction", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to create transaction")
		return
	}

	// Проверяем бюджет асинхронно — не блокируем ответ клиенту.
	if t.Type == domain.TypeExpense && t.Category != "" && h.budgets != nil && h.pub != nil {
		go h.checkBudget(userID, t.Category)
	}

	if h.pub != nil {
		_ = h.pub.Publish("transaction.created", map[string]any{
			"user_id":  userID.String(),
			"tx_id":    t.ID.String(),
			"type":     string(t.Type),
			"amount":   t.Amount,
			"category": t.Category,
		})
	}

	respondCreated(c, toTxDTO(t))
}

func (h *TransactionHandler) Get(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid transaction ID")
		return
	}
	t, err := h.tx.GetOwnedByID(c.Request.Context(), id, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Transaction not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load transaction")
		return
	}
	respondOK(c, toTxDTO(t))
}

type updateTxRequest struct {
	Title    string  `json:"title"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`
	Category string  `json:"category"`
	Date     string  `json:"date"`
}

func (h *TransactionHandler) Update(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid transaction ID")
		return
	}

	var req updateTxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	t, err := h.tx.GetOwnedByID(c.Request.Context(), id, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Transaction not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to load transaction")
		return
	}

	if req.Title != "" {
		t.Title = req.Title
	}
	if req.Amount != 0 {
		t.Amount = req.Amount
	}
	if req.Type != "" {
		t.Type = domain.TransactionType(req.Type)
	}
	if req.Category != "" {
		t.Category = req.Category
	}
	if req.Date != "" {
		t.Date = req.Date
	}

	if err := h.tx.Update(c.Request.Context(), t); err != nil {
		slog.Error("update transaction", "err", err, "id", t.ID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to update transaction")
		return
	}
	respondOK(c, toTxDTO(t))
}

func (h *TransactionHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid transaction ID")
		return
	}

	err = h.tx.Delete(c.Request.Context(), id, userID)
	if errors.Is(err, domain.ErrNotFound) {
		respondError(c, http.StatusNotFound, codeNotFound, "Transaction not found")
		return
	}
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to delete transaction")
		return
	}
	respondMessage(c, "Transaction deleted successfully")
}

type summaryDTO struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

func (h *TransactionHandler) Summary(c *gin.Context) {
	userID, _ := middleware.UserID(c)
	now := time.Now()

	year, err := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(now.Year())))
	if err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, "invalid year")
		return
	}
	month, err := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(now.Month()))))
	if err != nil || month < 1 || month > 12 {
		respondError(c, http.StatusBadRequest, codeValidation, "invalid month")
		return
	}

	income, expense, err := h.tx.SumMonth(c.Request.Context(), userID, year, month)
	if err != nil {
		slog.Error("summary", "err", err, "user_id", userID)
		respondError(c, http.StatusInternalServerError, codeInternal, "Failed to get summary")
		return
	}
	respondOK(c, summaryDTO{Income: income, Expense: expense, Balance: income - expense})
}

func (h *TransactionHandler) checkBudget(userID uuid.UUID, category string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	budget, err := h.budgets.GetByCategory(ctx, userID, category)
	if err != nil {
		return // бюджет не установлен — ничего не делаем
	}

	spent, err := h.budgets.MonthlySpent(ctx, userID, category)
	if err != nil {
		slog.Error("monthly spent query", "err", err, "user_id", userID, "category", category)
		return
	}

	if spent >= budget.MonthlyLimit {
		label := labelsRu[category]
		if label == "" {
			label = category
		}
		if err := h.pub.PublishBudgetExceeded(events.BudgetExceededPayload{
			UserID:       userID.String(),
			Category:     category,
			LabelRu:      label,
			MonthlyLimit: budget.MonthlyLimit,
			Spent:        spent,
		}); err != nil {
			slog.Error("publish budget.exceeded", "err", err, "user_id", userID)
		} else {
			slog.Info("budget exceeded — event published",
				"user_id", userID, "category", category,
				"limit", budget.MonthlyLimit, "spent", spent)
		}
	}
}
