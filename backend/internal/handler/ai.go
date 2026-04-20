package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"habitflow/internal/domain"
	"habitflow/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"habitflow/pkg/ai"
)

type AIHandler struct {
	client      *ai.Client
	txRepo      *repository.TransactionRepository
	goalRepo    *repository.GoalRepository
	rtRepo      *repository.RecurringTransactionRepository
	pendingRepo *repository.AIPendingCommandRepository
}

func NewAIHandler(client *ai.Client, txRepo *repository.TransactionRepository, goalRepo *repository.GoalRepository, rtRepo *repository.RecurringTransactionRepository, pendingRepo *repository.AIPendingCommandRepository) *AIHandler {
	return &AIHandler{
		client:      client,
		txRepo:      txRepo,
		goalRepo:    goalRepo,
		rtRepo:      rtRepo,
		pendingRepo: pendingRepo,
	}
}

type ChatMessageRequest struct {
	Agent   string `json:"agent" binding:"required"`
	Message string `json:"message" binding:"required"`
	Context string `json:"context,omitempty"`
}

type ChatMessageResponse struct {
	Response string `json:"response"`
}

type AICommandRequest struct {
	Message string `json:"message" binding:"required"`
	Context string `json:"context,omitempty"`
}

type AICommandTransactionData struct {
	Title    string   `json:"title"`
	Amount   *float64 `json:"amount"`
	Type     string   `json:"type"`
	Category string   `json:"category"`
	Date     string   `json:"date"`
}

type AICommandTransactionSelector struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Amount *float64 `json:"amount"`
	Date   string   `json:"date"`
}

type AICommandGoalData struct {
	Title       string  `json:"title"`
	Icon        string  `json:"icon"`
	TargetValue *int    `json:"target_value"`
	Unit        *string `json:"unit"`
	Deadline    *string `json:"deadline"`
}

type AICommandRecurringData struct {
	Title             string   `json:"title"`
	Amount            *float64 `json:"amount"`
	Type              string   `json:"type"`
	Category          string   `json:"category"`
	Frequency         string   `json:"frequency"`
	StartDate         string   `json:"start_date"`
	EndDate           *string  `json:"end_date"`
	RemainingPayments *int     `json:"remaining_payments"`
}

type AICommandModelResponse struct {
	Intent              string                      `json:"intent"`
	Confidence          float64                     `json:"confidence"`
	NeedsConfirmation   bool                        `json:"needs_confirmation"`
	MissingFields       []string                    `json:"missing_fields"`
	Message             string                      `json:"message"`
	Data                *AICommandTransactionData   `json:"data"`
	Items               []AICommandTransactionData  `json:"items"`
	TransactionSelector *AICommandTransactionSelector `json:"transaction_selector"`
	Goal                *AICommandGoalData          `json:"goal"`
	Recurring           *AICommandRecurringData     `json:"recurring"`
}

type AICommandResponse struct {
	Status        string                 `json:"status"`
	Intent        string                 `json:"intent"`
	Message       string                 `json:"message"`
	MissingFields []string               `json:"missing_fields,omitempty"`
	Transaction   *TransactionResponse   `json:"transaction,omitempty"`
	Transactions  []*TransactionResponse `json:"transactions,omitempty"`
	Goal          *GoalResponse          `json:"goal,omitempty"`
	Recurring     *RecurringResponse     `json:"recurring,omitempty"`
}

type PendingAICommand struct {
	Intent               string                       `json:"intent"`
	Data                 AICommandTransactionData     `json:"data,omitempty"`
	TransactionSelector  *AICommandTransactionSelector `json:"transaction_selector,omitempty"`
	Goal                 *AICommandGoalData           `json:"goal,omitempty"`
	Recurring            *AICommandRecurringData      `json:"recurring,omitempty"`
	MissingFields        []string                     `json:"missing_fields,omitempty"`
	Context              string                       `json:"context,omitempty"`
	AwaitingConfirmation bool                         `json:"awaiting_confirmation,omitempty"`
}

// InsightRequest for generating AI-powered insights
type InsightRequest struct {
	Type string `json:"type" binding:"required"` // habits, tasks, budget, weekly
	Data string `json:"data" binding:"required"` // JSON string with user data
}

// InsightItem represents a single insight
type InsightItem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

// InsightResponse for insight generation
type InsightResponse struct {
	Insights []InsightItem `json:"insights"`
}

// WeeklyInsightResponse for weekly review
type WeeklyInsightResponse struct {
	Summary      string   `json:"summary"`
	Wins         []string `json:"wins"`
	Improvements []string `json:"improvements"`
	Tip          string   `json:"tip"`
}

// Chat handles AI chat requests
// POST /api/v1/ai/chat
func (h *AIHandler) Chat(c *gin.Context) {
	var req ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	// Validate agent type
	agentType := ai.AgentType(req.Agent)
	switch agentType {
	case ai.AgentHabitCoach, ai.AgentTaskAssistant, ai.AgentFinanceAdvisor, ai.AgentLifeCoach:
		// Valid agent
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_AGENT", "message": "Invalid agent type"}})
		return
	}

	// Get system prompt
	systemPrompt := ai.GetSystemPrompt(agentType, req.Context)

	// Call OpenAI
	response, err := h.client.Chat(c.Request.Context(), systemPrompt, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ChatMessageResponse{Response: response}})
}

// Command handles AI command execution for structured actions.
// POST /api/v1/ai/command
func (h *AIHandler) Command(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var req AICommandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondValidationError(c, err.Error())
		return
	}

	systemPrompt := ai.GetCommandPrompt()
	pending, hasPending, err := h.getPendingCommand(c.Request.Context(), userID)
	if err != nil {
		log.Printf("ai/command: failed to load pending command for user %s: %v", userID, err)
		respondInternalError(c, "Failed to load pending AI command")
		return
	}
	if hasPending && shouldResetPendingCommand(req.Message) {
		if err := h.clearPendingCommand(c.Request.Context(), userID); err != nil {
			log.Printf("ai/command: failed to reset pending command for user %s: %v", userID, err)
			respondInternalError(c, "Failed to reset pending AI command")
			return
		}
		pending = PendingAICommand{}
		hasPending = false
	}
	if hasPending && pending.AwaitingConfirmation {
		h.handlePendingConfirmation(c, userID, req.Message, pending)
		return
	}
	userMessage := h.buildCommandUserMessage(c, userID, req, pending, hasPending)

	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		log.Printf("ai/command: openai chat failed for user %s: %v", userID, err)
		respondInternalError(c, "Failed to process AI command")
		return
	}

	command, err := parseAICommandResponse(response)
	if err != nil {
		log.Printf("ai/command: failed to parse model response for user %s: %v; raw=%q", userID, err, response)
		respondInternalError(c, "Failed to parse AI command")
		return
	}

	if hasPending {
		command = mergePendingCommand(command, pending)
	}

	switch command.Intent {
	case "create_transaction":
		// handled below
	case "update_transaction":
		h.handleUpdateTransactionCommand(c, userID, command)
		return
	case "delete_transaction":
		h.handleDeleteTransactionCommand(c, userID, command)
		return
	case "create_goal":
		h.handleCreateGoalCommand(c, userID, command)
		return
	case "create_recurring_transaction":
		h.handleCreateRecurringTransactionCommand(c, userID, command)
		return
	default:
		if hasPending && isTypeHelpQuestion(req.Message, pending.MissingFields) {
			respondOK(c, AICommandResponse{
				Status:        "needs_clarification",
				Intent:        pending.Intent,
				Message:       clarificationMessage(appendUniqueMissingFields(pending.MissingFields, "type")),
				MissingFields: appendUniqueMissingFields(pending.MissingFields, "type"),
			})
			return
		}
		respondOK(c, AICommandResponse{
			Status:  "unsupported",
			Intent:  command.Intent,
			Message: "This AI action is not supported yet",
		})
		return
	}

	if len(command.Items) > 0 {
		missingFields := validateTransactionCommandItems(command.Items)
		if len(missingFields) > 0 || command.NeedsConfirmation || command.Confidence < 0.6 {
			message := command.Message
			if message == "" {
				message = "Need more details before creating the transactions"
			}
			respondOK(c, AICommandResponse{
				Status:        "needs_clarification",
				Intent:        "create_transaction",
				Message:       message,
				MissingFields: missingFields,
			})
			return
		}

		created := make([]*TransactionResponse, 0, len(command.Items))
		for _, item := range command.Items {
			itemCopy := item
			tx, err := itemCopy.toDomainTransaction(userID)
			if err != nil {
				respondValidationError(c, err.Error())
				return
			}
			if err := h.txRepo.Create(c.Request.Context(), tx); err != nil {
				respondInternalError(c, "Failed to create transactions from AI command")
				return
			}
			created = append(created, toTransactionResponse(tx))
		}
		if err := h.clearPendingCommand(c.Request.Context(), userID); err != nil {
			log.Printf("ai/command: failed to clear pending command after multi-create for user %s: %v", userID, err)
			respondInternalError(c, "Failed to clear pending AI command")
			return
		}

		message := command.Message
		if message == "" {
			message = "Transactions created"
		}

		respondCreated(c, AICommandResponse{
			Status:       "completed",
			Intent:       command.Intent,
			Message:      message,
			Transactions: created,
		})
		return
	}

	missingFields := validateTransactionCommand(command.Data)
	if len(missingFields) > 0 || command.NeedsConfirmation || command.Confidence < 0.6 {
		if err := h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
			Intent:        "create_transaction",
			Data:          derefTransactionCommandData(command.Data),
			MissingFields: missingFields,
			Context:       req.Context,
		}); err != nil {
			log.Printf("ai/command: failed to save create_transaction pending command for user %s: %v", userID, err)
			respondInternalError(c, "Failed to save pending AI command")
			return
		}

		message := clarificationMessage(missingFields)
		if command.Message != "" && !isTypeHelpQuestion(req.Message, missingFields) {
			message = command.Message
		}
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "create_transaction",
			Message:       message,
			MissingFields: missingFields,
		})
		return
	}

	tx, err := command.Data.toDomainTransaction(userID)
	if err != nil {
		respondValidationError(c, err.Error())
		return
	}

	if err := h.txRepo.Create(c.Request.Context(), tx); err != nil {
		log.Printf("ai/command: failed to create transaction for user %s: %v", userID, err)
		respondInternalError(c, "Failed to create transaction from AI command")
		return
	}
	if err := h.clearPendingCommand(c.Request.Context(), userID); err != nil {
		log.Printf("ai/command: failed to clear pending command after create for user %s: %v", userID, err)
		respondInternalError(c, "Failed to clear pending AI command")
		return
	}

	message := command.Message
	if message == "" {
		message = "Transaction created"
	}

	respondCreated(c, AICommandResponse{
		Status:      "completed",
		Intent:      command.Intent,
		Message:     message,
		Transaction: toTransactionResponse(tx),
	})
}

// GenerateInsight generates AI-powered insights based on user data
// POST /api/v1/ai/insights
func (h *AIHandler) GenerateInsight(c *gin.Context) {
	var req InsightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	// Validate insight type
	insightType := ai.InsightType(req.Type)
	switch insightType {
	case ai.InsightHabits, ai.InsightTasks, ai.InsightBudget, ai.InsightWeekly:
		// Valid type
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_TYPE", "message": "Invalid insight type"}})
		return
	}

	// Get system prompt for insights
	systemPrompt := ai.GetInsightPrompt(insightType)

	// Create user message with data
	userMessage := "Analyze the following data and generate insights:\n\n" + req.Data

	// Call OpenAI
	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	// Parse response based on type
	if insightType == ai.InsightWeekly {
		var weeklyResponse WeeklyInsightResponse
		if err := json.Unmarshal([]byte(response), &weeklyResponse); err != nil {
			// If parsing fails, return raw response
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": weeklyResponse})
	} else {
		var insights []InsightItem
		if err := json.Unmarshal([]byte(response), &insights); err != nil {
			// If parsing fails, return raw response
			c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": InsightResponse{Insights: insights}})
	}
}

func (h *AIHandler) buildCommandUserMessage(c *gin.Context, userID uuid.UUID, req AICommandRequest, pending PendingAICommand, hasPending bool) string {
	var builder strings.Builder
	builder.WriteString("Current date: ")
	builder.WriteString(time.Now().Format("2006-01-02"))
	builder.WriteString("\n")

	if req.Context != "" {
		builder.WriteString("Provided UI context:\n")
		builder.WriteString(req.Context)
		builder.WriteString("\n")
	}

	if hasPending {
		builder.WriteString("Pending command intent: ")
		builder.WriteString(pending.Intent)
		builder.WriteString("\n")
		builder.WriteString("Existing draft:\n")
		appendPendingDraft(&builder, pending)
		if len(pending.MissingFields) > 0 {
			builder.WriteString("Still missing fields: ")
			builder.WriteString(strings.Join(pending.MissingFields, ", "))
			builder.WriteString("\n")
		}
		if pending.AwaitingConfirmation {
			builder.WriteString("The previous step is waiting for explicit confirmation. If the user confirms, keep the same intent and selector.\n")
		} else {
			builder.WriteString("Interpret the new user reply as a continuation of the pending command. Preserve existing draft fields unless the user clearly changes them.\n")
		}
	}

	if h.txRepo != nil {
		if transactions, err := h.txRepo.GetByUserID(c.Request.Context(), userID); err == nil && len(transactions) > 0 {
			limit := len(transactions)
			if limit > 5 {
				limit = 5
			}
			builder.WriteString("Recent transactions:\n")
			for i := 0; i < limit; i++ {
				tx := transactions[i]
				builder.WriteString("- ")
				builder.WriteString(tx.Date)
				builder.WriteString(" | ")
				builder.WriteString(tx.Title)
				builder.WriteString(" | ")
				builder.WriteString(string(tx.Type))
				builder.WriteString(" | ")
				builder.WriteString(tx.Category)
				builder.WriteString("\n")
			}
		}
	}

	builder.WriteString("User request:\n")
	builder.WriteString(req.Message)
	return builder.String()
}

func (h *AIHandler) handleUpdateTransactionCommand(c *gin.Context, userID uuid.UUID, command *AICommandModelResponse) {
	tx, err := h.resolveTransactionSelector(c, userID, command.TransactionSelector)
	if err != nil {
		missingFields, message := selectorResolutionMessage(err, "update")
		_ = h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
			Intent:              "update_transaction",
			Data:                derefTransactionCommandData(command.Data),
			TransactionSelector: command.TransactionSelector,
			MissingFields:       missingFields,
		})
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "update_transaction",
			Message:       message,
			MissingFields: missingFields,
		})
		return
	}

	if command.Data == nil {
		_ = h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
			Intent:              "update_transaction",
			TransactionSelector: command.TransactionSelector,
			MissingFields:       []string{"data"},
		})
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "update_transaction",
			Message:       "Нужно уточнить, какие поля обновить",
			MissingFields: []string{"data"},
		})
		return
	}

	changed := false
	if strings.TrimSpace(command.Data.Title) != "" {
		tx.Title = strings.TrimSpace(command.Data.Title)
		changed = true
	}
	if command.Data.Amount != nil && *command.Data.Amount > 0 {
		tx.Amount = *command.Data.Amount
		changed = true
	}
	if normalizedType := normalizeTransactionType(command.Data.Type); normalizedType != "" {
		tx.Type = normalizedType
		changed = true
	}
	if strings.TrimSpace(command.Data.Category) != "" {
		tx.Category = normalizeCategoryForType(tx.Type, command.Data.Category)
		changed = true
	}
	if strings.TrimSpace(command.Data.Date) != "" {
		if _, err := time.Parse("2006-01-02", command.Data.Date); err != nil {
			respondValidationError(c, "Invalid transaction date")
			return
		}
		tx.Date = command.Data.Date
		changed = true
	}
	tx.Category = normalizeCategoryForType(tx.Type, tx.Category)

	if !changed {
		_ = h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
			Intent:              "update_transaction",
			TransactionSelector: command.TransactionSelector,
			Data:                derefTransactionCommandData(command.Data),
			MissingFields:       []string{"data"},
		})
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "update_transaction",
			Message:       "Нужно уточнить, что именно обновить",
			MissingFields: []string{"data"},
		})
		return
	}

	if err := h.txRepo.Update(c.Request.Context(), tx); err != nil {
		respondInternalError(c, "Failed to update transaction from AI command")
		return
	}
	_ = h.clearPendingCommand(c.Request.Context(), userID)

	message := command.Message
	if message == "" {
		message = "Transaction updated"
	}

	respondOK(c, AICommandResponse{
		Status:      "completed",
		Intent:      "update_transaction",
		Message:     message,
		Transaction: toTransactionResponse(tx),
	})
}

func (h *AIHandler) handleDeleteTransactionCommand(c *gin.Context, userID uuid.UUID, command *AICommandModelResponse) {
	tx, err := h.resolveTransactionSelector(c, userID, command.TransactionSelector)
	if err != nil {
		missingFields, message := selectorResolutionMessage(err, "delete")
		_ = h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
			Intent:              "delete_transaction",
			TransactionSelector: command.TransactionSelector,
			MissingFields:       missingFields,
		})
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "delete_transaction",
			Message:       message,
			MissingFields: missingFields,
		})
		return
	}

	respondOK(c, AICommandResponse{
		Status:  "needs_confirmation",
		Intent:  "delete_transaction",
		Message: confirmationMessageForTransaction(tx),
	})
	_ = h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
		Intent:               "delete_transaction",
		TransactionSelector:  &AICommandTransactionSelector{ID: tx.ID.String(), Title: tx.Title, Amount: &tx.Amount, Date: tx.Date},
		AwaitingConfirmation: true,
	})
}

func (h *AIHandler) handleCreateGoalCommand(c *gin.Context, userID uuid.UUID, command *AICommandModelResponse) {
	missingFields := validateGoalCommand(command.Goal)
	if len(missingFields) > 0 {
		_ = h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
			Intent:        "create_goal",
			Goal:          command.Goal,
			MissingFields: missingFields,
		})
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "create_goal",
			Message:       clarificationMessageForGoal(missingFields),
			MissingFields: missingFields,
		})
		return
	}

	icon := strings.TrimSpace(command.Goal.Icon)
	if icon == "" {
		icon = "🎯"
	}

	goal := &domain.Goal{
		UserID:      userID,
		Title:       strings.TrimSpace(command.Goal.Title),
		Icon:        icon,
		TargetValue: command.Goal.TargetValue,
		Unit:        command.Goal.Unit,
	}

	if command.Goal.Deadline != nil && strings.TrimSpace(*command.Goal.Deadline) != "" {
		deadline, err := parseFlexibleDateTime(*command.Goal.Deadline)
		if err != nil {
			respondValidationError(c, "Invalid goal deadline")
			return
		}
		goal.Deadline = deadline
	}

	if err := h.goalRepo.Create(c.Request.Context(), goal); err != nil {
		respondInternalError(c, "Failed to create goal from AI command")
		return
	}
	_ = h.clearPendingCommand(c.Request.Context(), userID)

	message := command.Message
	if message == "" {
		message = "Goal created"
	}

	respondCreated(c, AICommandResponse{
		Status:  "completed",
		Intent:  "create_goal",
		Message: message,
		Goal:    toGoalResponse(goal),
	})
}

func (h *AIHandler) handleCreateRecurringTransactionCommand(c *gin.Context, userID uuid.UUID, command *AICommandModelResponse) {
	if command.Recurring == nil {
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "create_recurring_transaction",
			Message:       "Нужно уточнить данные для регулярной операции",
			MissingFields: []string{"recurring"},
		})
		return
	}

	rt, missing, err := command.Recurring.toDomainRecurringTransaction(userID)
	if err != nil {
		respondValidationError(c, err.Error())
		return
	}
	if len(missing) > 0 {
		_ = h.setPendingCommand(c.Request.Context(), userID, PendingAICommand{
			Intent:        "create_recurring_transaction",
			Recurring:     command.Recurring,
			MissingFields: missing,
		})
		respondOK(c, AICommandResponse{
			Status:        "needs_clarification",
			Intent:        "create_recurring_transaction",
			Message:       clarificationMessageForRecurring(missing),
			MissingFields: missing,
		})
		return
	}

	if err := h.rtRepo.Create(c.Request.Context(), rt); err != nil {
		respondInternalError(c, "Failed to create recurring transaction from AI command")
		return
	}
	_ = h.clearPendingCommand(c.Request.Context(), userID)

	message := command.Message
	if message == "" {
		message = "Recurring transaction created"
	}

	respondCreated(c, AICommandResponse{
		Status:    "completed",
		Intent:    "create_recurring_transaction",
		Message:   message,
		Recurring: toRecurringResponse(rt),
	})
}

func (h *AIHandler) getPendingCommand(ctx context.Context, userID uuid.UUID) (PendingAICommand, bool, error) {
	if h.pendingRepo == nil {
		return PendingAICommand{}, false, nil
	}

	payload, err := h.pendingRepo.GetPayloadByUserID(ctx, userID)
	if err != nil {
		return PendingAICommand{}, false, err
	}
	if len(payload) == 0 {
		return PendingAICommand{}, false, nil
	}

	var command PendingAICommand
	if err := json.Unmarshal(payload, &command); err != nil {
		return PendingAICommand{}, false, err
	}

	return command, true, nil
}

func (h *AIHandler) handlePendingConfirmation(c *gin.Context, userID uuid.UUID, message string, pending PendingAICommand) {
	if !pending.AwaitingConfirmation {
		return
	}

	if isNegativeConfirmation(message) {
		_ = h.clearPendingCommand(c.Request.Context(), userID)
		respondOK(c, AICommandResponse{
			Status:  "cancelled",
			Intent:  pending.Intent,
			Message: "Действие отменено",
		})
		return
	}

	if !isPositiveConfirmation(message) {
		respondOK(c, AICommandResponse{
			Status:  "needs_confirmation",
			Intent:  pending.Intent,
			Message: "Подтвердите действие: ответьте да или нет.",
		})
		return
	}

	switch pending.Intent {
	case "delete_transaction":
		tx, err := h.resolveTransactionSelector(c, userID, pending.TransactionSelector)
		if err != nil {
			_ = h.clearPendingCommand(c.Request.Context(), userID)
			missingFields, clarification := selectorResolutionMessage(err, "delete")
			respondOK(c, AICommandResponse{
				Status:        "needs_clarification",
				Intent:        pending.Intent,
				Message:       clarification,
				MissingFields: missingFields,
			})
			return
		}
		if err := h.txRepo.Delete(c.Request.Context(), tx.ID); err != nil {
			respondInternalError(c, "Failed to delete transaction from AI command")
			return
		}
		_ = h.clearPendingCommand(c.Request.Context(), userID)
		respondOK(c, AICommandResponse{
			Status:  "completed",
			Intent:  pending.Intent,
			Message: "Transaction deleted",
		})
	default:
		_ = h.clearPendingCommand(c.Request.Context(), userID)
		respondOK(c, AICommandResponse{
			Status:  "unsupported",
			Intent:  pending.Intent,
			Message: "Pending confirmation is not supported for this action",
		})
	}
}

func (h *AIHandler) resolveTransactionSelector(c *gin.Context, userID uuid.UUID, selector *AICommandTransactionSelector) (*domain.Transaction, error) {
	if selector == nil {
		return nil, domain.ErrInvalidInput
	}

	if strings.TrimSpace(selector.ID) != "" {
		txID, err := uuid.Parse(strings.TrimSpace(selector.ID))
		if err != nil {
			return nil, err
		}
		owns, err := h.txRepo.VerifyOwnership(c.Request.Context(), txID, userID)
		if err != nil || !owns {
			return nil, domain.ErrNotFound
		}
		return h.txRepo.GetByID(c.Request.Context(), txID)
	}

	if strings.TrimSpace(selector.Title) == "" {
		return nil, domain.ErrInvalidInput
	}

	transactions, err := h.txRepo.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		return nil, err
	}

	title := strings.ToLower(strings.TrimSpace(selector.Title))
	var matches []*domain.Transaction
	for _, tx := range transactions {
		if strings.ToLower(strings.TrimSpace(tx.Title)) != title {
			continue
		}
		if selector.Amount != nil && tx.Amount != *selector.Amount {
			continue
		}
		if strings.TrimSpace(selector.Date) != "" && tx.Date != strings.TrimSpace(selector.Date) {
			continue
		}
		matches = append(matches, tx)
	}

	switch len(matches) {
	case 0:
		return nil, domain.ErrNotFound
	case 1:
		return matches[0], nil
	default:
		return nil, domain.ErrAlreadyExists
	}
}

func (h *AIHandler) setPendingCommand(ctx context.Context, userID uuid.UUID, command PendingAICommand) error {
	if h.pendingRepo == nil {
		return nil
	}

	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}

	return h.pendingRepo.UpsertPayload(ctx, userID, payload)
}

func (h *AIHandler) clearPendingCommand(ctx context.Context, userID uuid.UUID) error {
	if h.pendingRepo == nil {
		return nil
	}
	return h.pendingRepo.DeleteByUserID(ctx, userID)
}

func mergePendingCommand(command *AICommandModelResponse, pending PendingAICommand) *AICommandModelResponse {
	if command == nil {
		command = &AICommandModelResponse{}
	}

	currentIntent := strings.TrimSpace(command.Intent)
	if currentIntent == "" || currentIntent == "unsupported" {
		command.Intent = pending.Intent
		currentIntent = pending.Intent
	}
	if currentIntent != pending.Intent {
		return command
	}

	switch pending.Intent {
	case "create_transaction":
		command.Data = mergeTransactionCommandData(&pending.Data, command.Data)
	case "update_transaction":
		command = normalizeSelectorClarification(command)
		command.TransactionSelector = mergeTransactionSelector(pending.TransactionSelector, command.TransactionSelector)
		command.Data = mergeTransactionCommandData(&pending.Data, command.Data)
	case "delete_transaction":
		command = normalizeSelectorClarification(command)
		command.TransactionSelector = mergeTransactionSelector(pending.TransactionSelector, command.TransactionSelector)
	case "create_goal":
		command.Goal = mergeGoalCommandData(pending.Goal, command.Goal)
	case "create_recurring_transaction":
		command.Recurring = mergeRecurringCommandData(pending.Recurring, command.Recurring)
	}

	return command
}

func normalizeSelectorClarification(command *AICommandModelResponse) *AICommandModelResponse {
	if command == nil {
		return nil
	}

	if command.TransactionSelector == nil {
		command.TransactionSelector = &AICommandTransactionSelector{}
	}
	if command.Data == nil {
		return command
	}

	if command.TransactionSelector.Amount == nil && command.Data.Amount != nil && *command.Data.Amount > 0 {
		command.TransactionSelector.Amount = command.Data.Amount
	}
	if strings.TrimSpace(command.TransactionSelector.Date) == "" && strings.TrimSpace(command.Data.Date) != "" {
		command.TransactionSelector.Date = strings.TrimSpace(command.Data.Date)
	}
	if strings.TrimSpace(command.TransactionSelector.Title) == "" && strings.TrimSpace(command.Data.Title) != "" {
		command.TransactionSelector.Title = strings.TrimSpace(command.Data.Title)
	}

	return command
}

func parseAICommandResponse(raw string) (*AICommandModelResponse, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)

	var parsed AICommandModelResponse
	if err := json.Unmarshal([]byte(clean), &parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseFlexibleDateTime(value string) (*time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}

	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return &parsed, nil
	}
	if parsed, err := time.Parse("2006-01-02", trimmed); err == nil {
		return &parsed, nil
	}

	return nil, domain.ErrInvalidInput
}

func validateTransactionCommand(data *AICommandTransactionData) []string {
	if data == nil {
		return []string{"title", "amount", "type"}
	}

	var missing []string
	if strings.TrimSpace(data.Title) == "" {
		missing = append(missing, "title")
	}
	if data.Amount == nil || *data.Amount <= 0 {
		missing = append(missing, "amount")
	}

	txType := normalizeTransactionType(data.Type)
	if txType == "" {
		missing = append(missing, "type")
	}

	if txType != "" {
		normalizedCategory := normalizeCategoryForType(txType, data.Category)
		if strings.TrimSpace(data.Category) != "" && normalizedCategory == "other" && normalizeTransactionCategory(data.Category) != "other" {
			missing = append(missing, "category")
		}
	}

	if data.Date != "" {
		if _, err := time.Parse("2006-01-02", data.Date); err != nil {
			missing = append(missing, "date")
		}
	}

	return missing
}

func validateTransactionCommandItems(items []AICommandTransactionData) []string {
	var missing []string
	for index, item := range items {
		itemMissing := validateTransactionCommand(&item)
		for _, field := range itemMissing {
			missing = append(missing, "item_"+strconv.Itoa(index+1)+"."+field)
		}
	}
	return missing
}

func mergeTransactionCommandData(existing *AICommandTransactionData, incoming *AICommandTransactionData) *AICommandTransactionData {
	var merged AICommandTransactionData
	if existing != nil {
		merged = *existing
	}
	if incoming == nil {
		return &merged
	}
	if strings.TrimSpace(incoming.Title) != "" {
		merged.Title = strings.TrimSpace(incoming.Title)
	}
	if incoming.Amount != nil && *incoming.Amount > 0 {
		merged.Amount = incoming.Amount
	}
	if strings.TrimSpace(incoming.Type) != "" {
		merged.Type = strings.TrimSpace(incoming.Type)
	}
	if strings.TrimSpace(incoming.Category) != "" {
		merged.Category = strings.TrimSpace(incoming.Category)
	}
	if strings.TrimSpace(incoming.Date) != "" {
		merged.Date = strings.TrimSpace(incoming.Date)
	}
	return &merged
}

func mergeTransactionSelector(existing *AICommandTransactionSelector, incoming *AICommandTransactionSelector) *AICommandTransactionSelector {
	if existing == nil && incoming == nil {
		return nil
	}
	var merged AICommandTransactionSelector
	if existing != nil {
		merged = *existing
	}
	if incoming == nil {
		return &merged
	}
	if strings.TrimSpace(incoming.ID) != "" {
		merged.ID = strings.TrimSpace(incoming.ID)
	}
	if strings.TrimSpace(incoming.Title) != "" {
		merged.Title = strings.TrimSpace(incoming.Title)
	}
	if incoming.Amount != nil && *incoming.Amount > 0 {
		merged.Amount = incoming.Amount
	}
	if strings.TrimSpace(incoming.Date) != "" {
		merged.Date = strings.TrimSpace(incoming.Date)
	}
	return &merged
}

func mergeGoalCommandData(existing *AICommandGoalData, incoming *AICommandGoalData) *AICommandGoalData {
	if existing == nil && incoming == nil {
		return nil
	}
	var merged AICommandGoalData
	if existing != nil {
		merged = *existing
	}
	if incoming == nil {
		return &merged
	}
	if strings.TrimSpace(incoming.Title) != "" {
		merged.Title = strings.TrimSpace(incoming.Title)
	}
	if strings.TrimSpace(incoming.Icon) != "" {
		merged.Icon = strings.TrimSpace(incoming.Icon)
	}
	if incoming.TargetValue != nil && *incoming.TargetValue > 0 {
		merged.TargetValue = incoming.TargetValue
	}
	if incoming.Unit != nil && strings.TrimSpace(*incoming.Unit) != "" {
		unit := strings.TrimSpace(*incoming.Unit)
		merged.Unit = &unit
	}
	if incoming.Deadline != nil && strings.TrimSpace(*incoming.Deadline) != "" {
		deadline := strings.TrimSpace(*incoming.Deadline)
		merged.Deadline = &deadline
	}
	return &merged
}

func mergeRecurringCommandData(existing *AICommandRecurringData, incoming *AICommandRecurringData) *AICommandRecurringData {
	if existing == nil && incoming == nil {
		return nil
	}
	var merged AICommandRecurringData
	if existing != nil {
		merged = *existing
	}
	if incoming == nil {
		return &merged
	}
	if strings.TrimSpace(incoming.Title) != "" {
		merged.Title = strings.TrimSpace(incoming.Title)
	}
	if incoming.Amount != nil && *incoming.Amount > 0 {
		merged.Amount = incoming.Amount
	}
	if strings.TrimSpace(incoming.Type) != "" {
		merged.Type = strings.TrimSpace(incoming.Type)
	}
	if strings.TrimSpace(incoming.Category) != "" {
		merged.Category = strings.TrimSpace(incoming.Category)
	}
	if strings.TrimSpace(incoming.Frequency) != "" {
		merged.Frequency = strings.TrimSpace(incoming.Frequency)
	}
	if strings.TrimSpace(incoming.StartDate) != "" {
		merged.StartDate = strings.TrimSpace(incoming.StartDate)
	}
	if incoming.EndDate != nil && strings.TrimSpace(*incoming.EndDate) != "" {
		endDate := strings.TrimSpace(*incoming.EndDate)
		merged.EndDate = &endDate
	}
	if incoming.RemainingPayments != nil && *incoming.RemainingPayments > 0 {
		merged.RemainingPayments = incoming.RemainingPayments
	}
	return &merged
}

func derefTransactionCommandData(data *AICommandTransactionData) AICommandTransactionData {
	if data == nil {
		return AICommandTransactionData{}
	}
	return *data
}

func normalizeTransactionType(value string) domain.TransactionType {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "expense", "расход", "трата":
		return domain.TransactionTypeExpense
	case "income", "доход":
		return domain.TransactionTypeIncome
	default:
		return ""
	}
}

func clarificationMessage(missingFields []string) string {
	if len(missingFields) == 0 {
		return "Need more details before creating the transaction"
	}

	if len(missingFields) == 1 && missingFields[0] == "type" {
		return "Уточните тип операции: доход или расход."
	}

	if containsField(missingFields, "type") {
		return "Нужно уточнить недостающие поля. Для типа операции доступны только значения: доход или расход."
	}

	return "Нужно уточнить недостающие поля: " + strings.Join(missingFields, ", ")
}

func clarificationMessageForGoal(missingFields []string) string {
	return "Нужно уточнить данные цели: " + strings.Join(missingFields, ", ")
}

func clarificationMessageForRecurring(missingFields []string) string {
	return "Нужно уточнить данные регулярной операции: " + strings.Join(missingFields, ", ")
}

func confirmationMessageForTransaction(tx *domain.Transaction) string {
	if tx == nil {
		return "Подтвердите удаление транзакции"
	}
	return "Подтвердите удаление транзакции: " + tx.Title + " на " + strings.TrimRight(strings.TrimRight(formatAmount(tx.Amount), "0"), ".") + " от " + tx.Date + "."
}

func selectorResolutionMessage(err error, action string) ([]string, string) {
	switch err {
	case domain.ErrInvalidInput:
		return []string{"transaction_selector"}, "Нужно уточнить, какую транзакцию выбрать."
	case domain.ErrNotFound:
		if action == "delete" {
			return []string{"transaction_selector"}, "Не нашел транзакцию для удаления. Уточните название, сумму или дату."
		}
		return []string{"transaction_selector"}, "Не нашел транзакцию для обновления. Уточните название, сумму или дату."
	case domain.ErrAlreadyExists:
		return []string{"transaction_selector"}, "Нашел несколько подходящих транзакций. Уточните сумму или дату."
	default:
		return []string{"transaction_selector"}, "Нужно уточнить, какую транзакцию выбрать."
	}
}

func appendPendingDraft(builder *strings.Builder, pending PendingAICommand) {
	if builder == nil {
		return
	}

	if pending.TransactionSelector != nil {
		if pending.TransactionSelector.ID != "" {
			builder.WriteString("- selector.id: ")
			builder.WriteString(pending.TransactionSelector.ID)
			builder.WriteString("\n")
		}
		if pending.TransactionSelector.Title != "" {
			builder.WriteString("- selector.title: ")
			builder.WriteString(pending.TransactionSelector.Title)
			builder.WriteString("\n")
		}
		if pending.TransactionSelector.Amount != nil {
			builder.WriteString("- selector.amount: ")
			builder.WriteString(strings.TrimRight(strings.TrimRight(formatAmount(*pending.TransactionSelector.Amount), "0"), "."))
			builder.WriteString("\n")
		}
		if pending.TransactionSelector.Date != "" {
			builder.WriteString("- selector.date: ")
			builder.WriteString(pending.TransactionSelector.Date)
			builder.WriteString("\n")
		}
	}

	if pending.Data.Title != "" {
		builder.WriteString("- title: ")
		builder.WriteString(pending.Data.Title)
		builder.WriteString("\n")
	}
	if pending.Data.Amount != nil {
		builder.WriteString("- amount: ")
		builder.WriteString(strings.TrimRight(strings.TrimRight(formatAmount(*pending.Data.Amount), "0"), "."))
		builder.WriteString("\n")
	}
	if pending.Data.Type != "" {
		builder.WriteString("- type: ")
		builder.WriteString(pending.Data.Type)
		builder.WriteString("\n")
	}
	if pending.Data.Category != "" {
		builder.WriteString("- category: ")
		builder.WriteString(pending.Data.Category)
		builder.WriteString("\n")
	}
	if pending.Data.Date != "" {
		builder.WriteString("- date: ")
		builder.WriteString(pending.Data.Date)
		builder.WriteString("\n")
	}

	if pending.Goal != nil {
		if pending.Goal.Title != "" {
			builder.WriteString("- goal.title: ")
			builder.WriteString(pending.Goal.Title)
			builder.WriteString("\n")
		}
		if pending.Goal.TargetValue != nil {
			builder.WriteString("- goal.target_value: ")
			builder.WriteString(strconv.Itoa(*pending.Goal.TargetValue))
			builder.WriteString("\n")
		}
		if pending.Goal.Unit != nil {
			builder.WriteString("- goal.unit: ")
			builder.WriteString(*pending.Goal.Unit)
			builder.WriteString("\n")
		}
		if pending.Goal.Deadline != nil {
			builder.WriteString("- goal.deadline: ")
			builder.WriteString(*pending.Goal.Deadline)
			builder.WriteString("\n")
		}
	}

	if pending.Recurring != nil {
		if pending.Recurring.Title != "" {
			builder.WriteString("- recurring.title: ")
			builder.WriteString(pending.Recurring.Title)
			builder.WriteString("\n")
		}
		if pending.Recurring.Amount != nil {
			builder.WriteString("- recurring.amount: ")
			builder.WriteString(strings.TrimRight(strings.TrimRight(formatAmount(*pending.Recurring.Amount), "0"), "."))
			builder.WriteString("\n")
		}
		if pending.Recurring.Type != "" {
			builder.WriteString("- recurring.type: ")
			builder.WriteString(pending.Recurring.Type)
			builder.WriteString("\n")
		}
		if pending.Recurring.Category != "" {
			builder.WriteString("- recurring.category: ")
			builder.WriteString(pending.Recurring.Category)
			builder.WriteString("\n")
		}
		if pending.Recurring.Frequency != "" {
			builder.WriteString("- recurring.frequency: ")
			builder.WriteString(pending.Recurring.Frequency)
			builder.WriteString("\n")
		}
		if pending.Recurring.StartDate != "" {
			builder.WriteString("- recurring.start_date: ")
			builder.WriteString(pending.Recurring.StartDate)
			builder.WriteString("\n")
		}
	}
}

func containsField(fields []string, target string) bool {
	for _, field := range fields {
		if field == target {
			return true
		}
	}
	return false
}

func appendUniqueMissingFields(fields []string, extra string) []string {
	if containsField(fields, extra) {
		return fields
	}
	return append(fields, extra)
}

func validateGoalCommand(goal *AICommandGoalData) []string {
	if goal == nil {
		return []string{"goal.title", "goal.target_value"}
	}

	var missing []string
	if strings.TrimSpace(goal.Title) == "" {
		missing = append(missing, "goal.title")
	}
	if goal.TargetValue == nil || *goal.TargetValue <= 0 {
		missing = append(missing, "goal.target_value")
	}
	if goal.Deadline != nil && strings.TrimSpace(*goal.Deadline) != "" {
		if _, err := parseFlexibleDateTime(*goal.Deadline); err != nil {
			missing = append(missing, "goal.deadline")
		}
	}
	return missing
}

func isTypeHelpQuestion(message string, missingFields []string) bool {
	if !containsField(missingFields, "type") {
		return false
	}
	normalized := strings.ToLower(strings.TrimSpace(message))
	return strings.Contains(normalized, "какие тип") ||
		strings.Contains(normalized, "какой тип") ||
		strings.Contains(normalized, "what types") ||
		strings.Contains(normalized, "which type")
}

func isPositiveConfirmation(message string) bool {
	normalized := strings.ToLower(strings.TrimSpace(message))
	switch normalized {
	case "да", "ага", "ок", "okay", "ok", "yes", "y", "подтверждаю", "confirm", "confirmed":
		return true
	default:
		return normalized == "удали" || normalized == "delete"
	}
}

func isNegativeConfirmation(message string) bool {
	normalized := strings.ToLower(strings.TrimSpace(message))
	switch normalized {
	case "нет", "не", "отмена", "cancel", "no", "n", "stop":
		return true
	default:
		return false
	}
}

func formatAmount(amount float64) string {
	return strconv.FormatFloat(amount, 'f', 2, 64)
}

func shouldResetPendingCommand(message string) bool {
	normalized := strings.ToLower(strings.TrimSpace(message))
	if normalized == "" {
		return false
	}
	if isPositiveConfirmation(normalized) || isNegativeConfirmation(normalized) {
		return false
	}

	explicitPrefixes := []string{
		"add ", "create ", "delete ", "remove ", "update ", "set ",
		"добав", "созда", "удал", "обнов", "запиш", "постав",
	}
	for _, prefix := range explicitPrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}

	return false
}

func (d *AICommandRecurringData) toDomainRecurringTransaction(userID uuid.UUID) (*domain.RecurringTransaction, []string, error) {
	if d == nil {
		return nil, []string{"recurring.title", "recurring.amount", "recurring.type", "recurring.frequency"}, nil
	}

	var missing []string
	if strings.TrimSpace(d.Title) == "" {
		missing = append(missing, "recurring.title")
	}
	if d.Amount == nil || *d.Amount <= 0 {
		missing = append(missing, "recurring.amount")
	}

	txType := normalizeTransactionType(d.Type)
	if txType == "" {
		missing = append(missing, "recurring.type")
	}

	frequency := domain.RecurrenceFrequency(strings.TrimSpace(d.Frequency))
	switch frequency {
	case domain.FrequencyWeekly, domain.FrequencyBiweekly, domain.FrequencyMonthly, domain.FrequencyQuarterly, domain.FrequencyYearly:
	default:
		missing = append(missing, "recurring.frequency")
	}

	startDate := strings.TrimSpace(d.StartDate)
	if startDate == "" {
		startDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return nil, nil, err
	}

	if len(missing) > 0 {
		return nil, missing, nil
	}

	var endDate *string
	if d.EndDate != nil && strings.TrimSpace(*d.EndDate) != "" {
		trimmed := strings.TrimSpace(*d.EndDate)
		if _, err := time.Parse("2006-01-02", trimmed); err != nil {
			return nil, nil, err
		}
		endDate = &trimmed
	}

	return &domain.RecurringTransaction{
		UserID:            userID,
		Title:             strings.TrimSpace(d.Title),
		Amount:            *d.Amount,
		Type:              txType,
		Category:          normalizeCategoryForType(txType, d.Category),
		Frequency:         frequency,
		StartDate:         startDate,
		NextDate:          startDate,
		EndDate:           endDate,
		RemainingPayments: d.RemainingPayments,
		IsActive:          true,
	}, nil, nil
}

func (d *AICommandTransactionData) toDomainTransaction(userID uuid.UUID) (*domain.Transaction, error) {
	txType := normalizeTransactionType(d.Type)
	if txType == "" {
		return nil, domain.ErrInvalidInput
	}

	date := strings.TrimSpace(d.Date)
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return nil, err
	}

	return &domain.Transaction{
		UserID:   userID,
		Title:    strings.TrimSpace(d.Title),
		Amount:   *d.Amount,
		Type:     txType,
		Category: normalizeCategoryForType(txType, d.Category),
		Date:     date,
	}, nil
}

// ExpenseAnalysisRequest for detailed expense analysis
type ExpenseAnalysisRequest struct {
	Data string `json:"data" binding:"required"` // JSON string with detailed expense data
}

// ExpenseInsightItem represents a detailed expense insight
type ExpenseInsightItem struct {
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Amount   *float64 `json:"amount,omitempty"`
	Category *string  `json:"category,omitempty"`
	Priority *int     `json:"priority,omitempty"`
}

// QuestionableTransactionItem represents a flagged transaction
type QuestionableTransactionItem struct {
	TransactionID   string   `json:"transactionId"`
	Reason          string   `json:"reason"`
	Category        string   `json:"category"`
	PotentialSavings *float64 `json:"potentialSavings,omitempty"`
}

// SavingsSuggestionItem represents a savings suggestion
type SavingsSuggestionItem struct {
	Category         string  `json:"category"`
	CurrentSpending  float64 `json:"currentSpending"`
	SuggestedBudget  float64 `json:"suggestedBudget"`
	PotentialSavings float64 `json:"potentialSavings"`
	Reason           string  `json:"reason"`
	Difficulty       string  `json:"difficulty"`
}

// ExpenseAnalysisResponse for expense analysis
type ExpenseAnalysisResponse struct {
	Insights                 []ExpenseInsightItem          `json:"insights"`
	QuestionableTransactions []QuestionableTransactionItem `json:"questionableTransactions"`
	SavingsSuggestions       []SavingsSuggestionItem       `json:"savingsSuggestions"`
}

// GoalToHabitsRequest for converting goals to habits
type GoalToHabitsRequest struct {
	GoalTitle    string  `json:"goalTitle" binding:"required"`
	GoalDeadline *string `json:"goalDeadline,omitempty"`
	TargetValue  *string `json:"targetValue,omitempty"`
	Context      *string `json:"context,omitempty"` // Additional context like "available time per day"
}

// SuggestedHabitItem represents a suggested habit
type SuggestedHabitItem struct {
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
	Period string `json:"period"`
	Reason string `json:"reason"`
}

// GoalToHabitsResponse for habit suggestions
type GoalToHabitsResponse struct {
	Habits      []SuggestedHabitItem `json:"habits"`
	Explanation string               `json:"explanation"`
}

// GenerateHabitsFromGoal converts an outcome goal to process habits
// POST /api/v1/ai/goal-to-habits
func (h *AIHandler) GenerateHabitsFromGoal(c *gin.Context) {
	var req GoalToHabitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	// Get system prompt for goal-to-habits conversion
	systemPrompt := ai.GetInsightPrompt(ai.InsightGoalToHabits)

	// Build user message with goal details
	userMessage := "Convert this outcome goal into process habits:\n\n"
	userMessage += "Goal: " + req.GoalTitle + "\n"
	if req.GoalDeadline != nil && *req.GoalDeadline != "" {
		userMessage += "Deadline: " + *req.GoalDeadline + "\n"
	}
	if req.TargetValue != nil && *req.TargetValue != "" {
		userMessage += "Target: " + *req.TargetValue + "\n"
	}
	if req.Context != nil && *req.Context != "" {
		userMessage += "Context: " + *req.Context + "\n"
	}

	// Call OpenAI
	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	// Parse structured response
	var habitsResponse GoalToHabitsResponse
	if err := json.Unmarshal([]byte(response), &habitsResponse); err != nil {
		// If parsing fails, return raw response
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": habitsResponse})
}

// GoalClarifyRequest for generating clarifying questions
type GoalClarifyRequest struct {
	GoalTitle string `json:"goalTitle" binding:"required"`
}

// ClarifyQuestion represents a clarifying question
type ClarifyQuestion struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}

// GoalClarifyResponse for clarifying questions
type GoalClarifyResponse struct {
	Questions   []ClarifyQuestion `json:"questions"`
	ContextHint string            `json:"context_hint"`
}

// GenerateGoalQuestions generates clarifying questions for a goal
// POST /api/v1/ai/goal-clarify
func (h *AIHandler) GenerateGoalQuestions(c *gin.Context) {
	var req GoalClarifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	// Get system prompt for goal clarification
	systemPrompt := ai.GetInsightPrompt(ai.InsightGoalClarify)

	// Build user message
	userMessage := "Generate clarifying questions for this goal:\n\nGoal: " + req.GoalTitle

	// Call OpenAI
	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	// Parse structured response
	var clarifyResponse GoalClarifyResponse
	if err := json.Unmarshal([]byte(response), &clarifyResponse); err != nil {
		// If parsing fails, return raw response
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": clarifyResponse})
}

// GenerateExpenseAnalysis generates detailed AI expense analysis
// POST /api/v1/ai/expense-analysis
func (h *AIHandler) GenerateExpenseAnalysis(c *gin.Context) {
	var req ExpenseAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	// Get system prompt for expense analysis
	systemPrompt := ai.GetInsightPrompt(ai.InsightExpenseAnalysis)

	// Create user message with detailed data
	userMessage := "Analyze this spending data and identify patterns, questionable expenses, and savings opportunities:\n\n" + req.Data

	// Call OpenAI
	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	// Parse structured response
	var analysisResponse ExpenseAnalysisResponse
	if err := json.Unmarshal([]byte(response), &analysisResponse); err != nil {
		// If parsing fails, return raw response for client to handle
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": analysisResponse})
}
