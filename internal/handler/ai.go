package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"habitflow/pkg/ai"
)

type AIHandler struct {
	client *ai.Client
}

func NewAIHandler(client *ai.Client) *AIHandler {
	return &AIHandler{client: client}
}

type ChatMessageRequest struct {
	Agent   string `json:"agent" binding:"required"`
	Message string `json:"message" binding:"required"`
	Context string `json:"context,omitempty"`
}

type ChatMessageResponse struct {
	Response string `json:"response"`
}

type InsightRequest struct {
	Type string `json:"type" binding:"required"`
	Data string `json:"data" binding:"required"`
}

type InsightItem struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

type InsightResponse struct {
	Insights []InsightItem `json:"insights"`
}

type WeeklyInsightResponse struct {
	Summary      string   `json:"summary"`
	Wins         []string `json:"wins"`
	Improvements []string `json:"improvements"`
	Tip          string   `json:"tip"`
}

func (h *AIHandler) Chat(c *gin.Context) {
	var req ChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	agentType := ai.AgentType(req.Agent)
	switch agentType {
	case ai.AgentHabitCoach, ai.AgentTaskAssistant, ai.AgentFinanceAdvisor, ai.AgentLifeCoach:

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_AGENT", "message": "Invalid agent type"}})
		return
	}

	systemPrompt := ai.GetSystemPrompt(agentType, req.Context)

	response, err := h.client.Chat(c.Request.Context(), systemPrompt, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": ChatMessageResponse{Response: response}})
}

func (h *AIHandler) GenerateInsight(c *gin.Context) {
	var req InsightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	insightType := ai.InsightType(req.Type)
	switch insightType {
	case ai.InsightHabits, ai.InsightTasks, ai.InsightBudget, ai.InsightWeekly:

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_TYPE", "message": "Invalid insight type"}})
		return
	}

	systemPrompt := ai.GetInsightPrompt(insightType)

	userMessage := "Analyze the following data and generate insights:\n\n" + req.Data

	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	if insightType == ai.InsightWeekly {
		var weeklyResponse WeeklyInsightResponse
		if err := json.Unmarshal([]byte(response), &weeklyResponse); err != nil {

			c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": weeklyResponse})
	} else {
		var insights []InsightItem
		if err := json.Unmarshal([]byte(response), &insights); err != nil {

			c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": InsightResponse{Insights: insights}})
	}
}

type ExpenseAnalysisRequest struct {
	Data string `json:"data" binding:"required"`
}

type ExpenseInsightItem struct {
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Message  string   `json:"message"`
	Amount   *float64 `json:"amount,omitempty"`
	Category *string  `json:"category,omitempty"`
	Priority *int     `json:"priority,omitempty"`
}

type QuestionableTransactionItem struct {
	TransactionID   string   `json:"transactionId"`
	Reason          string   `json:"reason"`
	Category        string   `json:"category"`
	PotentialSavings *float64 `json:"potentialSavings,omitempty"`
}

type SavingsSuggestionItem struct {
	Category         string  `json:"category"`
	CurrentSpending  float64 `json:"currentSpending"`
	SuggestedBudget  float64 `json:"suggestedBudget"`
	PotentialSavings float64 `json:"potentialSavings"`
	Reason           string  `json:"reason"`
	Difficulty       string  `json:"difficulty"`
}

type ExpenseAnalysisResponse struct {
	Insights                 []ExpenseInsightItem          `json:"insights"`
	QuestionableTransactions []QuestionableTransactionItem `json:"questionableTransactions"`
	SavingsSuggestions       []SavingsSuggestionItem       `json:"savingsSuggestions"`
}

type GoalToHabitsRequest struct {
	GoalTitle    string  `json:"goalTitle" binding:"required"`
	GoalDeadline *string `json:"goalDeadline,omitempty"`
	TargetValue  *string `json:"targetValue,omitempty"`
	Context      *string `json:"context,omitempty"`
}

type SuggestedHabitItem struct {
	Title  string `json:"title"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
	Period string `json:"period"`
	Reason string `json:"reason"`
}

type GoalToHabitsResponse struct {
	Habits      []SuggestedHabitItem `json:"habits"`
	Explanation string               `json:"explanation"`
}

func (h *AIHandler) GenerateHabitsFromGoal(c *gin.Context) {
	var req GoalToHabitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	systemPrompt := ai.GetInsightPrompt(ai.InsightGoalToHabits)

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

	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	var habitsResponse GoalToHabitsResponse
	if err := json.Unmarshal([]byte(response), &habitsResponse); err != nil {

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": habitsResponse})
}

type GoalClarifyRequest struct {
	GoalTitle string `json:"goalTitle" binding:"required"`
}

type ClarifyQuestion struct {
	ID          string `json:"id"`
	Question    string `json:"question"`
	Placeholder string `json:"placeholder"`
	Type        string `json:"type"`
}

type GoalClarifyResponse struct {
	Questions   []ClarifyQuestion `json:"questions"`
	ContextHint string            `json:"context_hint"`
}

func (h *AIHandler) GenerateGoalQuestions(c *gin.Context) {
	var req GoalClarifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	systemPrompt := ai.GetInsightPrompt(ai.InsightGoalClarify)

	userMessage := "Generate clarifying questions for this goal:\n\nGoal: " + req.GoalTitle

	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	var clarifyResponse GoalClarifyResponse
	if err := json.Unmarshal([]byte(response), &clarifyResponse); err != nil {

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": clarifyResponse})
}

func (h *AIHandler) GenerateExpenseAnalysis(c *gin.Context) {
	var req ExpenseAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "INVALID_REQUEST", "message": err.Error()}})
		return
	}

	systemPrompt := ai.GetInsightPrompt(ai.InsightExpenseAnalysis)

	userMessage := "Analyze this spending data and identify patterns, questionable expenses, and savings opportunities:\n\n" + req.Data

	response, err := h.client.Chat(c.Request.Context(), systemPrompt, userMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "AI_ERROR", "message": err.Error()}})
		return
	}

	var analysisResponse ExpenseAnalysisResponse
	if err := json.Unmarshal([]byte(response), &analysisResponse); err != nil {

		c.JSON(http.StatusOK, gin.H{"data": gin.H{"raw": response}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": analysisResponse})
}
