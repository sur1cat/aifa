package handler

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sur1cat/aifa/ai-service/internal/localai"
	"github.com/sur1cat/aifa/ai-service/internal/openai"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	client      *openai.Client
	localClient *localai.Client
}

func NewAIHandler(c *openai.Client, lc *localai.Client) *AIHandler {
	return &AIHandler{client: c, localClient: lc}
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

// ---------------- chat ----------------

type chatRequest struct {
	Agent   string `json:"agent" binding:"required"`
	Message string `json:"message" binding:"required"`
	Context string `json:"context,omitempty"`
}

type chatResponseBody struct {
	Response string `json:"response"`
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
	resp, ok := h.chat(c, openai.SystemPrompt(agent, req.Context), req.Message)
	if !ok {
		return
	}
	respondOK(c, chatResponseBody{Response: resp})
}

// ---------------- insights ----------------

type insightRequest struct {
	Type string `json:"type" binding:"required"`
	Data string `json:"data" binding:"required"`
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

	raw, ok := h.chat(c, openai.InsightPrompt(insight), "Analyze the following data and generate insights:\n\n"+req.Data)
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
	Message string `json:"message" binding:"required"`
	Context string `json:"context,omitempty"`
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
	Title        string  `json:"title"`
	TargetAmount *float64 `json:"target_amount,omitempty"`
	Deadline     *string  `json:"deadline,omitempty"`
	Description  string  `json:"description,omitempty"`
}

type commandPlan struct {
	Goal   commandGoal    `json:"goal"`
	Habits []commandHabit `json:"habits"`
	Tasks  []commandTask  `json:"tasks"`
}

type commandTransaction struct {
	Type          string  `json:"type"`           // "expense" | "income"
	Amount        float64 `json:"amount"`
	Title         string  `json:"title"`
	Category      string  `json:"category"`
	CategoryLabel string  `json:"category_label"`
	Date          string  `json:"date"`
}

type commandResponse struct {
	Intent      string              `json:"intent"`
	Response    string              `json:"response"`
	Transaction *commandTransaction `json:"transaction,omitempty"`
	Habit       *commandHabit       `json:"habit,omitempty"`
	Task        *commandTask        `json:"task,omitempty"`
	Tasks       []commandTask       `json:"tasks,omitempty"`
	Plan        *commandPlan        `json:"plan,omitempty"`
	Advice      string              `json:"advice,omitempty"`
}

func (h *AIHandler) Command(c *gin.Context) {
	var req commandRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, codeValidation, err.Error())
		return
	}

	systemPrompt := openai.CommandPrompt()
	if req.Context != "" {
		systemPrompt += "\n\n## User Context:\n" + req.Context
	}

	raw, ok := h.chat(c, systemPrompt, req.Message)
	if !ok {
		return
	}

	var body commandResponse
	respondJSONOrRaw(c, raw, &body)
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

	result, err := h.localClient.ParseMessage(c.Request.Context(), req.Message, req.DebtsContext)
	if err != nil {
		slog.Error("localai parse-message", "err", err)
		respondError(c, http.StatusServiceUnavailable, "AI_UNAVAILABLE", "AI local service unavailable")
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
