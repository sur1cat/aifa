package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sur1cat/aifa/ai-service/internal/openai"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	client *openai.Client
}

func NewAIHandler(c *openai.Client) *AIHandler {
	return &AIHandler{client: c}
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

type goalToHabitsRequest struct {
	GoalTitle    string  `json:"goalTitle" binding:"required"`
	GoalDeadline *string `json:"goalDeadline,omitempty"`
	TargetValue  *string `json:"targetValue,omitempty"`
	Context      *string `json:"context,omitempty"`
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

	msg := "Convert this outcome goal into process habits:\n\nGoal: " + req.GoalTitle + "\n"
	if req.GoalDeadline != nil && *req.GoalDeadline != "" {
		msg += "Deadline: " + *req.GoalDeadline + "\n"
	}
	if req.TargetValue != nil && *req.TargetValue != "" {
		msg += "Target: " + *req.TargetValue + "\n"
	}
	if req.Context != nil && *req.Context != "" {
		msg += "Context: " + *req.Context + "\n"
	}

	raw, ok := h.chat(c, openai.InsightPrompt(openai.InsightGoalToHabits), msg)
	if !ok {
		return
	}
	var body goalToHabitsBody
	respondJSONOrRaw(c, raw, &body)
}

// ---------------- goal clarify ----------------

type goalClarifyRequest struct {
	GoalTitle string `json:"goalTitle" binding:"required"`
}

type clarifyQuestion struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
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
	raw, ok := h.chat(c,
		openai.InsightPrompt(openai.InsightGoalClarify),
		"Generate clarifying questions for this goal:\n\nGoal: "+req.GoalTitle,
	)
	if !ok {
		return
	}
	var body goalClarifyBody
	respondJSONOrRaw(c, raw, &body)
}
