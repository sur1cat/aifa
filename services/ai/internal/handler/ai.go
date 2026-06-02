package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode"

	"github.com/sur1cat/aifa/ai-service/internal/finance"
	"github.com/sur1cat/aifa/ai-service/internal/localai"
	"github.com/sur1cat/aifa/ai-service/internal/openai"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	client      openAIClient
	localClient localAIClient
	finance     financeClient
}

type openAIClient interface {
	Chat(ctx context.Context, systemPrompt, userMessage string) (string, error)
	ChatWithVision(ctx context.Context, systemPrompt, base64Image, mimeType string) (string, error)
	Transcribe(ctx context.Context, audioData []byte, filename, language string) (string, error)
}

type localAIClient interface {
	CategorizeExpense(ctx context.Context, text string) (*localai.CategoryResult, error)
	BatchCategorizeExpenses(ctx context.Context, texts []string) ([]localai.CategoryResult, error)
	TranscribeVoice(ctx context.Context, audioData []byte, filename, language string) (*localai.VoiceTranscribeResult, error)
	ScanReceipt(ctx context.Context, imageData []byte, filename string) (*localai.ReceiptScanResult, error)
	ParseMessage(ctx context.Context, message string, debtsContext []map[string]any) (*localai.ParseMessageResponse, error)
	Forecast(ctx context.Context, transactions []localai.ForecastTransaction, horizonDays int, refDate string) (*localai.ForecastResponse, error)
	DetectAnomalies(ctx context.Context, transactions []localai.ForecastTransaction, sensitivity string) (*localai.AnomalyResponse, error)
	SpendingSummary(ctx context.Context, transactions []localai.InsightTransaction, periodStart, periodEnd string) (*localai.SummaryResponse, error)
	BudgetSuggestions(ctx context.Context, transactions []localai.InsightTransaction, lookbackDays int, percentile float64) (*localai.BudgetSuggestResponse, error)
}

type financeClient interface {
	CreateDebt(ctx context.Context, authHeader string, req finance.CreateDebtRequest) error
	ListDebts(ctx context.Context, authHeader string, settled *bool) ([]finance.Debt, error)
	PatchDebt(ctx context.Context, authHeader, debtID string, req finance.PatchDebtRequest) error
}

func NewAIHandler(c openAIClient, lc localAIClient, fc ...financeClient) *AIHandler {
	var financeClient financeClient
	if len(fc) > 0 {
		financeClient = fc[0]
	}
	return &AIHandler{client: c, localClient: lc, finance: financeClient}
}

func (h *AIHandler) chat(c *gin.Context, systemPrompt, userMessage string) (string, bool) {
	resp, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, openai.ErrNotConfigured) {
			status = http.StatusServiceUnavailable
		}
		slog.Error("openai chat", "err", err)
		respondError(c, status, codeAIError, err.Error())
		return "", false
	}
	return resp, true
}

// respondJSONOrRaw parses the model output into `target` and responds with it;
// if parsing fails, returns { raw: "<text>" } so clients can display the
// fallback. The model is instructed to emit JSON but occasionally produces
// conversational text — the client UI already handles the raw case.
func respondJSONOrRaw(c *gin.Context, raw string, target any) {
	if err := json.Unmarshal([]byte(raw), target); err != nil {
		respondOK(c, gin.H{"raw": raw})
		return
	}
	respondOK(c, target)
}

func authHeader(c *gin.Context) string {
	return strings.TrimSpace(c.GetHeader("Authorization"))
}

func sameCounterparty(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	if a == "" || b == "" {
		return false
	}
	return a == b || strings.Contains(a, b) || strings.Contains(b, a)
}

func (h *AIHandler) executeCreateDebt(ctx context.Context, auth string, debt *commandDebt) error {
	if h.finance == nil || debt == nil {
		return nil
	}
	return h.finance.CreateDebt(ctx, auth, finance.CreateDebtRequest{
		Counterparty: strings.TrimSpace(debt.Counterparty),
		Direction:    debt.Direction,
		Amount:       debt.Amount,
	})
}

func (h *AIHandler) executeSettleDebt(ctx context.Context, auth string, counterparty string) error {
	if h.finance == nil {
		return nil
	}
	settled := false
	debts, err := h.finance.ListDebts(ctx, auth, &settled)
	if err != nil {
		return err
	}
	for _, debt := range debts {
		if sameCounterparty(debt.Counterparty, counterparty) {
			return h.finance.PatchDebt(ctx, auth, debt.ID, finance.PatchDebtRequest{Settle: true})
		}
	}
	return fmt.Errorf("active debt not found for %q", counterparty)
}

func (h *AIHandler) applyCommandSideEffects(c *gin.Context, body *commandResponse) bool {
	if body == nil || body.Status != cmdStatusCompleted {
		return true
	}
	auth := authHeader(c)
	switch body.Intent {
	case "create_debt":
		if err := h.executeCreateDebt(c.Request.Context(), auth, body.Debt); err != nil {
			slog.Error("execute create_debt", "err", err)
			respondError(c, http.StatusBadGateway, codeAIError, "Failed to save debt")
			return false
		}
	case "settle_debt":
		if body.SettleDebt == nil {
			return true
		}
		if err := h.executeSettleDebt(c.Request.Context(), auth, body.SettleDebt.Counterparty); err != nil {
			slog.Error("execute settle_debt", "err", err)
			respondError(c, http.StatusBadGateway, codeAIError, "Failed to settle debt")
			return false
		}
	}
	return true
}

func (h *AIHandler) liveDebtsContext(ctx context.Context, auth string) []map[string]any {
	if h.finance == nil {
		return nil
	}
	settled := false
	debts, err := h.finance.ListDebts(ctx, auth, &settled)
	if err != nil {
		slog.Error("load active debts", "err", err)
		return nil
	}
	out := make([]map[string]any, 0, len(debts))
	for _, debt := range debts {
		out = append(out, map[string]any{
			"id":           debt.ID,
			"counterparty": debt.Counterparty,
			"amount":       debt.Amount,
			"direction":    debt.Direction,
			"settled":      debt.Settled,
		})
	}
	return out
}

func (h *AIHandler) maybeRefreshDebtContext(ctx context.Context, auth string, reqCtx []map[string]any, parsed *localai.ParseMessageResponse) ([]map[string]any, bool) {
	if parsed == nil || parsed.Intent != "update_debt" {
		return reqCtx, false
	}
	if parsed.DebtUpdate == nil || parsed.DebtUpdate.Type != "not_found" {
		return reqCtx, false
	}
	live := h.liveDebtsContext(ctx, auth)
	if len(live) == 0 {
		return reqCtx, false
	}
	return live, true
}

func (h *AIHandler) applyParsedDebtSideEffects(c *gin.Context, parsed *localai.ParseMessageResponse) bool {
	if parsed == nil || h.finance == nil {
		return true
	}
	auth := authHeader(c)
	switch parsed.Intent {
	case "create_debt":
		if parsed.Counterparty == nil || parsed.DebtDirection == nil || parsed.Amount == nil {
			return true
		}
		if *parsed.Amount <= 0 {
			return true
		}
		if err := h.finance.CreateDebt(c.Request.Context(), auth, finance.CreateDebtRequest{
			Counterparty: strings.TrimSpace(*parsed.Counterparty),
			Direction:    *parsed.DebtDirection,
			Amount:       *parsed.Amount,
		}); err != nil {
			slog.Error("execute parsed create_debt", "err", err)
			respondError(c, http.StatusBadGateway, codeAIError, "Failed to save debt")
			return false
		}
	case "update_debt":
		if parsed.Counterparty == nil || parsed.DebtUpdate == nil {
			return true
		}
		if parsed.DebtUpdate.Type == "not_found" {
			return true
		}
		debtID := ""
		if parsed.DebtUpdate.DebtID != nil {
			debtID = strings.TrimSpace(*parsed.DebtUpdate.DebtID)
		}
		if debtID == "" {
			return true
		}
		req := finance.PatchDebtRequest{}
		if parsed.DebtUpdate.ReduceBy > 0 {
			reduceBy := parsed.DebtUpdate.ReduceBy
			req.ReduceBy = &reduceBy
		} else {
			req.Settle = true
		}
		if err := h.finance.PatchDebt(c.Request.Context(), auth, debtID, req); err != nil {
			slog.Error("execute parsed update_debt", "err", err)
			respondError(c, http.StatusBadGateway, codeAIError, "Failed to update debt")
			return false
		}
	}
	return true
}

var supportedDomainKeywords = []string{
	// finance
	"деньг", "финанс", "бюджет", "расход", "доход", "зарплат", "зп", "потрат", "купил", "купила",
	"заплат", "плач", "получа", "перевод", "долг", "долж", "одолж", "взайм", "взаймы", "накоп", "сбереж", "кредит", "инвест", "ежемесяч", "в месяц", "каждый месяц", "expense", "income", "budget",
	"spent", "spend", "salary", "debt", "save", "saving", "savings", "transaction", "finance", "money",
	// habits / goals
	"привыч", "каждый день", "ежеднев", "стрик", "цель", "бег", "читать", "медит", "тренир",
	"просып", "похуд", "англий", "study", "habit", "daily", "routine", "streak", "goal", "run", "running", "read", "meditat", "wake up", "workout", "exercise",
	// tasks / planning
	"задач", "таск", "напом", "сделать", "выполн", "дедлайн", "план", "todo", "task", "remind", "reminder",
	"complete", "deadline", "plan", "planning",
}

func containsCyrillic(s string) bool {
	for _, r := range s {
		if unicode.In(r, unicode.Cyrillic) {
			return true
		}
	}
	return false
}

func unsupportedDomainMessage(message string) string {
	if containsCyrillic(message) {
		return "Я могу помогать только с финансами, привычками и задачами в AIFA. Сформулируй запрос в одной из этих тем."
	}
	return "I can only help with finances, habits, and tasks in AIFA. Please rephrase your request within one of those areas."
}

func isSupportedDomainMessage(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	for _, kw := range supportedDomainKeywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// ---------------- chat ----------------

type chatRequest struct {
	Agent   string           `json:"agent" binding:"required"`
	Message string           `json:"message" binding:"required"`
	Context string           `json:"context,omitempty"`
	History []historyMessage `json:"history,omitempty"`
}

type chatResponseBody struct {
	Response string `json:"response"`
}

type historyMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func formatHistory(history []historyMessage) string {
	if len(history) == 0 {
		return ""
	}
	var b strings.Builder
	for _, item := range history {
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		role := strings.TrimSpace(item.Role)
		if role == "" {
			role = "unknown"
		}
		b.WriteString("- ")
		b.WriteString(role)
		b.WriteString(": ")
		b.WriteString(content)
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String())
}

func historyContainsSupportedDomain(history []historyMessage) bool {
	for _, item := range history {
		if isSupportedDomainMessage(item.Content) {
			return true
		}
	}
	return false
}

func isContextualFollowUp(message string, history []historyMessage) bool {
	if !historyContainsSupportedDomain(history) {
		return false
	}
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	followUpSignals := []string{
		"он", "она", "оно", "это", "стоит", "стоил", "стоила", "стоило",
		"amount", "сумма", "цена", "цено", "was", "cost", "it", "that",
	}
	for _, kw := range followUpSignals {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	agent := openai.AgentType(req.Agent)
	if !openai.KnownAgent(agent) {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid agent type")
		return
	}
	if !isSupportedDomainMessage(req.Message) && !isContextualFollowUp(req.Message, req.History) {
		respondOK(c, chatResponseBody{Response: unsupportedDomainMessage(req.Message)})
		return
	}
	context := strings.TrimSpace(req.Context)
	if history := formatHistory(req.History); history != "" {
		if context != "" {
			context += "\n\nConversation History:\n" + history
		} else {
			context = "Conversation History:\n" + history
		}
	}
	resp, ok := h.chat(c, openai.SystemPrompt(agent, context), req.Message)
	if !ok {
		return
	}
	respondOK(c, chatResponseBody{Response: resp})
}

// ---------------- insights ----------------

type insightRequest struct {
	Type   string `json:"type" binding:"required"`
	Data   string `json:"data" binding:"required"`
	Locale string `json:"locale,omitempty"`
}

type insightItem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type weeklyInsightBody struct {
	Summary      string   `json:"summary"`
	Wins         []string `json:"wins"`
	Improvements []string `json:"improvements"`
	Tip          string   `json:"tip"`
}

func normalizeInsightLocale(locale string) string {
	l := strings.ToLower(strings.TrimSpace(locale))
	switch {
	case strings.HasPrefix(l, "ru"):
		return "ru"
	case strings.HasPrefix(l, "kk"), strings.HasPrefix(l, "kz"):
		return "kk"
	case strings.HasPrefix(l, "en"):
		return "en"
	default:
		return ""
	}
}

func insightLocaleInstruction(locale string) string {
	switch normalizeInsightLocale(locale) {
	case "ru":
		return "CRITICAL: Respond ENTIRELY in Russian regardless of the language of titles or raw data."
	case "kk":
		return "CRITICAL: Respond ENTIRELY in Kazakh regardless of the language of titles or raw data."
	case "en":
		return "CRITICAL: Respond ENTIRELY in English regardless of the language of titles or raw data."
	default:
		return ""
	}
}

func (h *AIHandler) GenerateInsight(c *gin.Context) {
	var req insightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	insight := openai.InsightType(req.Type)
	if !openai.KnownInsight(insight) {
		respondError(c, http.StatusBadRequest, codeBadRequest, "Invalid insight type")
		return
	}

	userMessage := "Analyze the following data and generate insights:\n\n" + req.Data
	if instruction := insightLocaleInstruction(req.Locale); instruction != "" {
		userMessage = instruction + "\n\n" + userMessage
	}

	raw, ok := h.chat(c, openai.InsightPrompt(insight), userMessage)
	if !ok {
		return
	}

	if insight == openai.InsightWeekly {
		var body weeklyInsightBody
		respondJSONOrRaw(c, raw, &body)
		return
	}
	var items []insightItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		respondOK(c, gin.H{"raw": raw})
		return
	}
	respondOK(c, gin.H{"insights": items})
}

// ---------------- expense analysis ----------------

type expenseRequest struct {
	Data string `json:"data" binding:"required"`
}

type expenseInsightItem struct {
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Amount   *float64 `json:"amount,omitempty"`
	Category *string  `json:"category,omitempty"`
	Priority *int     `json:"priority,omitempty"`
}

type questionableTx struct {
	TransactionID    string   `json:"transactionId"`
	Reason           string   `json:"reason"`
	Category         string   `json:"category"`
	PotentialSavings *float64 `json:"potentialSavings,omitempty"`
}

type savingsSuggestion struct {
	Category         string  `json:"category"`
	CurrentSpending  float64 `json:"currentSpending"`
	SuggestedBudget  float64 `json:"suggestedBudget"`
	PotentialSavings float64 `json:"potentialSavings"`
	Reason           string  `json:"reason"`
	Difficulty       string  `json:"difficulty"`
}

type expenseResponseBody struct {
	Insights                 []expenseInsightItem `json:"insights"`
	QuestionableTransactions []questionableTx     `json:"questionableTransactions"`
	SavingsSuggestions       []savingsSuggestion  `json:"savingsSuggestions"`
}

func (h *AIHandler) GenerateExpenseAnalysis(c *gin.Context) {
	var req expenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	raw, ok := h.chat(c,
		openai.InsightPrompt(openai.InsightExpenseAnalysis),
		"Analyze this spending data and identify patterns, questionable expenses, and savings opportunities:\n\n"+req.Data,
	)
	if !ok {
		return
	}
	var body expenseResponseBody
	respondJSONOrRaw(c, raw, &body)
}

// ---------------- goal → habits ----------------

type goalAnswerPair struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// goalToHabitsRequest accepts both the new Flutter contract (`goal` + `answers`)
// and the legacy fields (`goalTitle`, `goalDeadline`, `targetValue`, `context`).
type goalToHabitsRequest struct {
	Goal         string           `json:"goal,omitempty"`
	GoalTitle    string           `json:"goalTitle,omitempty"`
	GoalDeadline *string          `json:"goalDeadline,omitempty"`
	TargetValue  *string          `json:"targetValue,omitempty"`
	Context      *string          `json:"context,omitempty"`
	Answers      []goalAnswerPair `json:"answers,omitempty"`
}

type suggestedHabit struct {
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
	Period string `json:"period"`
	Reason string `json:"reason"`
}

type goalToHabitsBody struct {
	Habits      []suggestedHabit `json:"habits"`
	Explanation string           `json:"explanation"`
}

func (h *AIHandler) GenerateHabitsFromGoal(c *gin.Context) {
	var req goalToHabitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	title := req.Goal
	if title == "" {
		title = req.GoalTitle
	}
	title = strings.TrimSpace(title)
	if title == "" {
		respondError(c, http.StatusBadRequest, codeValidation, "goal is required")
		return
	}

	msg := "Convert this outcome goal into process habits:\n\nGoal: " + title + "\n"
	if req.GoalDeadline != nil && *req.GoalDeadline != "" {
		msg += "Deadline: " + *req.GoalDeadline + "\n"
	}
	if req.TargetValue != nil && *req.TargetValue != "" {
		msg += "Target: " + *req.TargetValue + "\n"
	}
	if req.Context != nil && *req.Context != "" {
		msg += "Context: " + *req.Context + "\n"
	}
	if len(req.Answers) > 0 {
		msg += "\nClarifying answers:\n"
		for _, a := range req.Answers {
			q := strings.TrimSpace(a.Question)
			ans := strings.TrimSpace(a.Answer)
			if q == "" || ans == "" {
				continue
			}
			msg += "- " + q + " → " + ans + "\n"
		}
	}

	raw, ok := h.chat(c, openai.InsightPrompt(openai.InsightGoalToHabits), msg)
	if !ok {
		return
	}
	var body goalToHabitsBody
	respondJSONOrRaw(c, raw, &body)
}

// ---------------- goal clarify ----------------

// goalClarifyRequest accepts both the new Flutter contract (`goal`) and the
// legacy `goalTitle` field.
type goalClarifyRequest struct {
	Goal      string `json:"goal,omitempty"`
	GoalTitle string `json:"goalTitle,omitempty"`
}

type clarifyQuestion struct {
	ID          string   `json:"id"`
	Question    string   `json:"question"`
	Placeholder string   `json:"placeholder"`
	Type        string   `json:"type"`
	Options     []string `json:"options,omitempty"`
}

type goalClarifyBody struct {
	Questions   []clarifyQuestion `json:"questions"`
	ContextHint string            `json:"context_hint"`
}

func (h *AIHandler) GenerateGoalQuestions(c *gin.Context) {
	var req goalClarifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	title := strings.TrimSpace(req.Goal)
	if title == "" {
		title = strings.TrimSpace(req.GoalTitle)
	}
	if title == "" {
		respondError(c, http.StatusBadRequest, codeValidation, "goal is required")
		return
	}
	raw, ok := h.chat(c,
		openai.InsightPrompt(openai.InsightGoalClarify),
		"Generate clarifying questions for this goal:\n\nGoal: "+title,
	)
	if !ok {
		return
	}
	var body goalClarifyBody
	respondJSONOrRaw(c, raw, &body)
}

// ---------------- universal command ----------------

type commandRequest struct {
	Message string           `json:"message" binding:"required"`
	Context string           `json:"context,omitempty"`
	History []historyMessage `json:"history,omitempty"`
}

type commandHabit struct {
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
	Period string `json:"period"`
	Reason string `json:"reason"`
}

type commandTask struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Priority    string `json:"priority"`
}

type commandGoal struct {
	Title        string   `json:"title"`
	TargetAmount *float64 `json:"target_amount,omitempty"`
	Deadline     *string  `json:"deadline,omitempty"`
	Description  string   `json:"description,omitempty"`
}

type commandPlan struct {
	Goal   commandGoal    `json:"goal"`
	Habits []commandHabit `json:"habits"`
	Tasks  []commandTask  `json:"tasks"`
}

type commandTransaction struct {
	Type          string  `json:"type"` // "expense" | "income"
	Amount        float64 `json:"amount"`
	Title         string  `json:"title"`
	Category      string  `json:"category"`
	CategoryLabel string  `json:"category_label"`
	Date          string  `json:"date"`
}

type commandDebt struct {
	Counterparty string  `json:"counterparty"`
	Direction    string  `json:"direction"` // "i_owe" | "they_owe"
	Amount       float64 `json:"amount"`
	Note         string  `json:"note,omitempty"`
}

type commandSettleDebt struct {
	Counterparty string `json:"counterparty"`
}

type commandRecurring struct {
	Title     string  `json:"title"`
	Amount    float64 `json:"amount"`
	Type      string  `json:"type"`      // "income" | "expense"
	Frequency string  `json:"frequency"` // "daily" | "weekly" | "monthly" | "yearly"
	Category  string  `json:"category"`
}

// commandResponse is what we send back to the Flutter client.
//
// `status` and `message` are added on top of the OpenAI-emitted shape so the
// client can route the response without re-parsing the LLM intent string.
// `message` mirrors `response` for backward-compatibility with older clients.
type commandResponse struct {
	Status        string              `json:"status"`
	Message       string              `json:"message"`
	MissingFields []string            `json:"missing_fields,omitempty"`
	Intent        string              `json:"intent"`
	Response      string              `json:"response"`
	Transaction   *commandTransaction `json:"transaction,omitempty"`
	Habit         *commandHabit       `json:"habit,omitempty"`
	Task          *commandTask        `json:"task,omitempty"`
	Tasks         []commandTask       `json:"tasks,omitempty"`
	Plan          *commandPlan        `json:"plan,omitempty"`
	Debt          *commandDebt        `json:"debt,omitempty"`
	SettleDebt    *commandSettleDebt  `json:"settle_debt,omitempty"`
	Recurring     *commandRecurring   `json:"recurring,omitempty"`
	Advice        string              `json:"advice,omitempty"`
}

const (
	cmdStatusCompleted          = "completed"
	cmdStatusNeedsClarification = "needs_clarification"
	cmdStatusNeedsConfirmation  = "needs_confirmation"
	cmdStatusUnsupported        = "unsupported"
)

// deriveCommandStatus turns the LLM-emitted `intent` + payload completeness
// into the four-state status the Flutter chat controller understands. Any
// intent that doesn't actually act on user data ("chat", "advice",
// "unsupported", or anything unknown) gets `unsupported` so the client can
// render `message` as a free-form reply without a second OpenAI call.
func deriveCommandStatus(body *commandResponse) []string {
	missing := []string{}
	switch body.Intent {
	case "create_transaction":
		if body.Transaction == nil {
			body.Status = cmdStatusNeedsClarification
			missing = append(missing, "transaction")
			return missing
		}
		if body.Transaction.Amount <= 0 {
			missing = append(missing, "amount")
		}
		if strings.TrimSpace(body.Transaction.Title) == "" {
			missing = append(missing, "title")
		}
		if len(missing) > 0 {
			body.Status = cmdStatusNeedsClarification
			return missing
		}
		body.Status = cmdStatusCompleted
	case "create_habit":
		if body.Habit == nil || strings.TrimSpace(body.Habit.Title) == "" {
			body.Status = cmdStatusNeedsClarification
			missing = append(missing, "habit")
			return missing
		}
		body.Status = cmdStatusNeedsConfirmation
	case "create_task":
		if body.Task == nil || strings.TrimSpace(body.Task.Title) == "" {
			body.Status = cmdStatusNeedsClarification
			missing = append(missing, "task")
			return missing
		}
		body.Status = cmdStatusNeedsConfirmation
	case "create_plan":
		if body.Plan == nil || strings.TrimSpace(body.Plan.Goal.Title) == "" {
			body.Status = cmdStatusNeedsClarification
			missing = append(missing, "plan")
			return missing
		}
		body.Status = cmdStatusNeedsConfirmation
	case "create_debt":
		if body.Debt == nil {
			body.Status = cmdStatusNeedsClarification
			missing = append(missing, "debt")
			return missing
		}
		if strings.TrimSpace(body.Debt.Counterparty) == "" {
			missing = append(missing, "counterparty")
		}
		if body.Debt.Amount <= 0 {
			missing = append(missing, "amount")
		}
		if body.Debt.Direction != "i_owe" && body.Debt.Direction != "they_owe" {
			missing = append(missing, "direction")
		}
		if len(missing) > 0 {
			body.Status = cmdStatusNeedsClarification
			return missing
		}
		body.Status = cmdStatusCompleted
	case "settle_debt":
		if body.SettleDebt == nil || strings.TrimSpace(body.SettleDebt.Counterparty) == "" {
			body.Status = cmdStatusNeedsClarification
			missing = append(missing, "settle_debt")
			return missing
		}
		body.Status = cmdStatusCompleted
	case "create_recurring":
		if body.Recurring == nil {
			body.Status = cmdStatusNeedsClarification
			missing = append(missing, "recurring")
			return missing
		}
		if strings.TrimSpace(body.Recurring.Title) == "" {
			missing = append(missing, "title")
		}
		if body.Recurring.Amount <= 0 {
			missing = append(missing, "amount")
		}
		if body.Recurring.Type != "income" && body.Recurring.Type != "expense" {
			missing = append(missing, "type")
		}
		if strings.TrimSpace(body.Recurring.Frequency) == "" {
			missing = append(missing, "frequency")
		}
		if strings.TrimSpace(body.Recurring.Category) == "" {
			missing = append(missing, "category")
		}
		if len(missing) > 0 {
			body.Status = cmdStatusNeedsClarification
			return missing
		}
		body.Status = cmdStatusCompleted
	default:
		// "chat", "advice", "unsupported", or unknown — surface the
		// model's `response` text but let the client know there's no
		// structured action to apply.
		body.Status = cmdStatusUnsupported
	}
	return missing
}

func (h *AIHandler) Command(c *gin.Context) {
	var req commandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if !isSupportedDomainMessage(req.Message) && !isContextualFollowUp(req.Message, req.History) {
		msg := unsupportedDomainMessage(req.Message)
		respondOK(c, commandResponse{
			Status:   cmdStatusUnsupported,
			Message:  msg,
			Intent:   "unsupported",
			Response: msg,
		})
		return
	}

	systemPrompt := openai.CommandPrompt()
	var promptContext strings.Builder
	if strings.TrimSpace(req.Context) != "" {
		promptContext.WriteString(strings.TrimSpace(req.Context))
	}
	if history := formatHistory(req.History); history != "" {
		if promptContext.Len() > 0 {
			promptContext.WriteString("\n\n")
		}
		promptContext.WriteString("Conversation History:\n")
		promptContext.WriteString(history)
	}
	if promptContext.Len() > 0 {
		systemPrompt += "\n\n## User Context:\n" + promptContext.String()
	}

	raw, ok := h.chat(c, systemPrompt, req.Message)
	if !ok {
		return
	}

	var body commandResponse
	if err := json.Unmarshal([]byte(raw), &body); err != nil {
		// Model didn't produce JSON — degrade gracefully so the client
		// can still display the text.
		respondOK(c, commandResponse{
			Status:  cmdStatusUnsupported,
			Message: strings.TrimSpace(raw),
			Intent:  "chat",
		})
		return
	}
	body.MissingFields = deriveCommandStatus(&body)
	if body.Message == "" {
		body.Message = body.Response
	}
	if body.Advice != "" && body.Message == "" {
		body.Message = body.Advice
	}
	if !h.applyCommandSideEffects(c, &body) {
		return
	}
	respondOK(c, body)
}

// ---------------- local AI: message parser ----------------

type parseMessageRequest struct {
	Message      string           `json:"message" binding:"required"`
	DebtsContext []map[string]any `json:"debts_context,omitempty"`
}

// ParseMessage парсит свободный текст в транзакцию без OpenAI.
// Пример: "Я потратил 7000 на обед" → intent=create_transaction, amount=7000, category=food
func (h *AIHandler) ParseMessage(c *gin.Context) {
	var req parseMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	auth := authHeader(c)
	debtsContext := req.DebtsContext
	if len(debtsContext) == 0 {
		debtsContext = h.liveDebtsContext(c.Request.Context(), auth)
	}

	result, err := h.localClient.ParseMessage(c.Request.Context(), req.Message, debtsContext)
	if err != nil {
		slog.Error("localai parse-message", "err", err)
		respondError(c, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI local service unavailable")
		return
	}
	refreshed, shouldRetry := h.maybeRefreshDebtContext(c.Request.Context(), auth, debtsContext, result)
	if shouldRetry {
		result, err = h.localClient.ParseMessage(c.Request.Context(), req.Message, refreshed)
		if err != nil {
			slog.Error("localai parse-message retry", "err", err)
			respondError(c, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI local service unavailable")
			return
		}
	}
	if !h.applyParsedDebtSideEffects(c, result) {
		return
	}
	respondOK(c, result)
}

// ---------------- local AI: expense categorization ----------------

type categorizeRequest struct {
	Text string `json:"text" binding:"required"`
}

// categorizeResult расширяет ответ ai-local полем source для трассировки.
type categorizeResult struct {
	Text       string  `json:"text"`
	Category   string  `json:"category"`
	LabelRu    string  `json:"label_ru"`
	LabelKz    string  `json:"label_kz"`
	Confidence float64 `json:"confidence"`
	Confident  bool    `json:"confident"`
	Source     string  `json:"source"` // "local" | "gpt4"
}

// gpt4FallbackCategory вызывает GPT-4 когда ai-local не уверена.
func (h *AIHandler) gpt4FallbackCategory(c *gin.Context, text string) *categorizeResult {
	type gptCategory struct {
		Category string `json:"category"`
		LabelRu  string `json:"label_ru"`
		LabelKz  string `json:"label_kz"`
	}

	raw, ok := h.chat(c, openai.CategorizeFallbackPrompt(), "Classify this transaction: "+text)
	if !ok {
		return nil
	}

	var parsed gptCategory
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		slog.Warn("gpt4 categorize: failed to parse", "raw", raw)
		return nil
	}

	return &categorizeResult{
		Text:       text,
		Category:   parsed.Category,
		LabelRu:    parsed.LabelRu,
		LabelKz:    parsed.LabelKz,
		Confidence: 1.0,
		Confident:  true,
		Source:     "gpt4",
	}
}

func (h *AIHandler) CategorizeExpense(c *gin.Context) {
	var req categorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	local, err := h.localClient.CategorizeExpense(c.Request.Context(), req.Text)
	if err != nil {
		slog.Error("localai categorize", "err", err)
		respondError(c, http.StatusServiceUnavailable, codeAIError, "local AI service unavailable")
		return
	}

	// Уверенный результат — возвращаем сразу.
	if local.Confident {
		respondOK(c, categorizeResult{
			Text: local.Text, Category: local.Category,
			LabelRu: local.LabelRu, LabelKz: local.LabelKz,
			Confidence: local.Confidence, Confident: true,
			Source: "local",
		})
		return
	}

	// Низкая уверенность — fallback на GPT-4.
	slog.Info("localai low confidence, fallback to gpt4",
		"text", req.Text, "confidence", local.Confidence)

	if result := h.gpt4FallbackCategory(c, req.Text); result != nil {
		respondOK(c, result)
		return
	}

	// GPT-4 тоже не смог — возвращаем лучшее что есть от ai-local.
	respondOK(c, categorizeResult{
		Text: local.Text, Category: local.Category,
		LabelRu: local.LabelRu, LabelKz: local.LabelKz,
		Confidence: local.Confidence, Confident: false,
		Source: "local",
	})
}

type batchCategorizeRequest struct {
	Texts []string `json:"texts" binding:"required"`
}

func (h *AIHandler) BatchCategorizeExpenses(c *gin.Context) {
	var req batchCategorizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	locals, err := h.localClient.BatchCategorizeExpenses(c.Request.Context(), req.Texts)
	if err != nil {
		slog.Error("localai batch categorize", "err", err)
		respondError(c, http.StatusServiceUnavailable, codeAIError, "local AI service unavailable")
		return
	}

	results := make([]categorizeResult, 0, len(locals))
	for _, l := range locals {
		if l.Confident {
			results = append(results, categorizeResult{
				Text: l.Text, Category: l.Category,
				LabelRu: l.LabelRu, LabelKz: l.LabelKz,
				Confidence: l.Confidence, Confident: true,
				Source: "local",
			})
			continue
		}

		// Fallback на GPT-4 для неуверенных результатов.
		// В batch-режиме ошибка GPT-4 не прерывает обработку — возвращаем local результат.
		slog.Info("localai batch: low confidence, fallback to gpt4",
			"text", l.Text, "confidence", l.Confidence)

		gptResult, _ := h.client.Chat(c.Request.Context(),
			openai.CategorizeFallbackPrompt(), "Classify this transaction: "+l.Text)

		parsed := struct {
			Category string `json:"category"`
			LabelRu  string `json:"label_ru"`
			LabelKz  string `json:"label_kz"`
		}{}
		if gptResult != "" && json.Unmarshal([]byte(gptResult), &parsed) == nil && parsed.Category != "" {
			results = append(results, categorizeResult{
				Text: l.Text, Category: parsed.Category,
				LabelRu: parsed.LabelRu, LabelKz: parsed.LabelKz,
				Confidence: 1.0, Confident: true, Source: "gpt4",
			})
		} else {
			results = append(results, categorizeResult{
				Text: l.Text, Category: l.Category,
				LabelRu: l.LabelRu, LabelKz: l.LabelKz,
				Confidence: l.Confidence, Confident: false,
				Source: "local",
			})
		}
	}

	respondOK(c, gin.H{"results": results})
}

// ---------------- voice transcription ----------------

type voiceResult struct {
	Transcript  string   `json:"transcript"`
	Amount      *float64 `json:"amount"`
	Currency    string   `json:"currency"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	LabelRu     string   `json:"label_ru"`
	LabelKz     string   `json:"label_kz"`
	Confidence  float64  `json:"confidence"`
}

func (h *AIHandler) TranscribeVoice(c *gin.Context) {
	file, header, err := c.Request.FormFile("audio")
	if err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, "audio file is required")
		return
	}
	defer file.Close()

	const maxSize = 25 << 20 // 25 MB — лимит Whisper API
	if header.Size > maxSize {
		respondError(c, http.StatusBadRequest, codeValidation, "audio too large (max 25 MB)")
		return
	}

	audioBytes, err := io.ReadAll(file)
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeAIError, "failed to read audio")
		return
	}

	// Flutter sends the hint as a multipart `language` field; older clients
	// pass it as `?lang=`. Accept both.
	lang := c.PostForm("language")
	if lang == "" {
		lang = c.PostForm("lang")
	}
	if lang == "" {
		lang = c.Query("lang")
	}

	if h.localClient != nil {
		local, err := h.localClient.TranscribeVoice(c.Request.Context(), audioBytes, header.Filename, lang)
		if err == nil {
			respondOK(c, local)
			return
		}
		slog.Warn("local whisper failed, trying OpenAI fallback", "err", err)
	}

	transcript, err := h.client.Transcribe(c.Request.Context(), audioBytes, header.Filename, lang)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, openai.ErrNotConfigured) {
			status = http.StatusServiceUnavailable
		}
		slog.Error("whisper transcribe", "err", err)
		respondError(c, status, codeAIError, err.Error())
		return
	}

	transcript = strings.TrimSpace(transcript)

	// Парсим транскрипцию в структуру транзакции через GPT-4
	raw, err := h.client.Chat(c.Request.Context(), openai.VoiceParsePrompt(), transcript)
	if err != nil {
		// Если GPT-4 недоступен — возвращаем хотя бы транскрипцию
		slog.Warn("voice parse: gpt4 unavailable, returning transcript only", "err", err)
		respondOK(c, voiceResult{Transcript: transcript, Confidence: 0})
		return
	}

	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}

	var result voiceResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		slog.Warn("voice parse: failed to parse json", "raw", raw)
		respondOK(c, voiceResult{Transcript: transcript, Confidence: 0})
		return
	}
	result.Transcript = transcript
	respondOK(c, result)
}

// ---------------- receipt OCR ----------------

type receiptResult struct {
	Amount     *float64 `json:"amount"`
	Currency   string   `json:"currency"`
	Date       *string  `json:"date"`
	Merchant   string   `json:"merchant"`
	Category   string   `json:"category"`
	LabelRu    string   `json:"label_ru"`
	LabelKz    string   `json:"label_kz"`
	Items      []string `json:"items"`
	Confidence float64  `json:"confidence"`
	RawTotal   string   `json:"raw_total"`
}

func (h *AIHandler) ScanReceipt(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, "image file is required")
		return
	}
	defer file.Close()

	const maxSize = 10 << 20 // 10 MB
	if header.Size > maxSize {
		respondError(c, http.StatusBadRequest, codeValidation, "image too large (max 10 MB)")
		return
	}

	imgBytes, err := io.ReadAll(file)
	if err != nil {
		respondError(c, http.StatusInternalServerError, codeAIError, "failed to read image")
		return
	}

	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" || !strings.HasPrefix(mimeType, "image/") {
		mimeType = "image/jpeg"
	}

	if h.localClient != nil {
		local, err := h.localClient.ScanReceipt(c.Request.Context(), imgBytes, header.Filename)
		if err == nil {
			respondOK(c, local)
			return
		}
		slog.Warn("local OCR failed, trying OpenAI fallback", "err", err)
	}

	b64 := base64.StdEncoding.EncodeToString(imgBytes)

	raw, err := h.client.ChatWithVision(c.Request.Context(), openai.ReceiptScanPrompt(), b64, mimeType)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, openai.ErrNotConfigured) {
			status = http.StatusServiceUnavailable
		}
		slog.Error("receipt scan vision", "err", err)
		respondError(c, status, codeAIError, err.Error())
		return
	}

	// Снимаем markdown-обёртку если модель добавила ```json
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		raw = strings.TrimPrefix(raw, "```json")
		raw = strings.TrimPrefix(raw, "```")
		raw = strings.TrimSuffix(raw, "```")
		raw = strings.TrimSpace(raw)
	}

	var result receiptResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		slog.Warn("receipt scan: failed to parse json", "raw", raw)
		respondOK(c, gin.H{"raw": raw})
		return
	}
	respondOK(c, result)
}

// ── Insights ──────────────────────────────────────────────────────────────────

type insightTxItem struct {
	Date     string  `json:"date" binding:"required"`
	Amount   float64 `json:"amount" binding:"required,gt=0"`
	Type     string  `json:"type" binding:"required,oneof=income expense"`
	Category string  `json:"category"`
}

type summaryRequest struct {
	Transactions []insightTxItem `json:"transactions" binding:"required,min=1"`
	PeriodStart  string          `json:"period_start"`
	PeriodEnd    string          `json:"period_end"`
}

func toInsightTxs(items []insightTxItem) []localai.InsightTransaction {
	out := make([]localai.InsightTransaction, len(items))
	for i, t := range items {
		out[i] = localai.InsightTransaction{Date: t.Date, Amount: t.Amount, Type: t.Type, Category: t.Category}
	}
	return out
}

func (h *AIHandler) SpendingSummary(c *gin.Context) {
	var req summaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if h.localClient == nil {
		respondError(c, http.StatusServiceUnavailable, codeAIError, "ai-local-service is not configured")
		return
	}
	result, err := h.localClient.SpendingSummary(c.Request.Context(), toInsightTxs(req.Transactions), req.PeriodStart, req.PeriodEnd)
	if err != nil {
		slog.Error("localai spending summary", "err", err)
		respondError(c, http.StatusServiceUnavailable, codeAIError, err.Error())
		return
	}
	respondOK(c, result)
}

type budgetSuggestRequest struct {
	Transactions []insightTxItem `json:"transactions" binding:"required,min=1"`
	LookbackDays int             `json:"lookback_days"`
	Percentile   float64         `json:"percentile"`
}

func (h *AIHandler) BudgetSuggestions(c *gin.Context) {
	var req budgetSuggestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if h.localClient == nil {
		respondError(c, http.StatusServiceUnavailable, codeAIError, "ai-local-service is not configured")
		return
	}
	if req.LookbackDays <= 0 {
		req.LookbackDays = 90
	}
	if req.Percentile <= 0 {
		req.Percentile = 75
	}
	result, err := h.localClient.BudgetSuggestions(c.Request.Context(), toInsightTxs(req.Transactions), req.LookbackDays, req.Percentile)
	if err != nil {
		slog.Error("localai budget suggestions", "err", err)
		respondError(c, http.StatusServiceUnavailable, codeAIError, err.Error())
		return
	}
	respondOK(c, result)
}

// ── Forecast ──────────────────────────────────────────────────────────────────

type forecastRequest struct {
	Transactions []localai.ForecastTransaction `json:"transactions" binding:"required,min=1"`
	HorizonDays  int                           `json:"horizon_days"`
	RefDate      string                        `json:"ref_date"`
}

func (h *AIHandler) Forecast(c *gin.Context) {
	var req forecastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if h.localClient == nil {
		respondError(c, http.StatusServiceUnavailable, codeAIError, "ai-local-service is not configured")
		return
	}
	if req.HorizonDays <= 0 {
		req.HorizonDays = 30
	}

	result, err := h.localClient.Forecast(c.Request.Context(), req.Transactions, req.HorizonDays, req.RefDate)
	if err != nil {
		slog.Error("localai forecast", "err", err)
		respondError(c, http.StatusServiceUnavailable, codeAIError, err.Error())
		return
	}
	respondOK(c, result)
}

// ── Anomalies ─────────────────────────────────────────────────────────────────

type anomalyRequest struct {
	Transactions []localai.ForecastTransaction `json:"transactions" binding:"required,min=1"`
	Sensitivity  string                        `json:"sensitivity"`
}

func (h *AIHandler) DetectAnomalies(c *gin.Context) {
	var req anomalyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}
	if h.localClient == nil {
		respondError(c, http.StatusServiceUnavailable, codeAIError, "ai-local-service is not configured")
		return
	}

	result, err := h.localClient.DetectAnomalies(c.Request.Context(), req.Transactions, req.Sensitivity)
	if err != nil {
		slog.Error("localai anomalies", "err", err)
		respondError(c, http.StatusServiceUnavailable, codeAIError, err.Error())
		return
	}
	respondOK(c, result)
}
