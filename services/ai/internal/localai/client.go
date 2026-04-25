package localai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client обращается к ai-local-service (Python/FastAPI).
type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// CategoryResult — ответ от /categorize.
type CategoryResult struct {
	Text       string  `json:"text"`
	Category   string  `json:"category"`
	LabelRu    string  `json:"label_ru"`
	LabelKz    string  `json:"label_kz"`
	Confidence float64 `json:"confidence"`
	Confident  bool    `json:"confident"`
}

// CategorizeExpense классифицирует название транзакции.
// Если уверенность модели низкая (confident=false) — лучше использовать OpenAI.
func (c *Client) CategorizeExpense(ctx context.Context, text string) (*CategoryResult, error) {
	body, _ := json.Marshal(map[string]string{"text": text})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/categorize", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("localai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("localai: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("localai: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("localai: status %d: %s", resp.StatusCode, string(raw))
	}

	var result CategoryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("localai: decode response: %w", err)
	}
	return &result, nil
}

// BatchCategorizeExpenses классифицирует список транзакций за один запрос.
func (c *Client) BatchCategorizeExpenses(ctx context.Context, texts []string) ([]CategoryResult, error) {
	body, _ := json.Marshal(map[string][]string{"texts": texts})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/categorize/batch", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("localai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("localai: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("localai: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("localai: status %d: %s", resp.StatusCode, string(raw))
	}

	var response struct {
		Results []CategoryResult `json:"results"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return nil, fmt.Errorf("localai: decode response: %w", err)
	}
	return response.Results, nil
}

// ── Forecast ─────────────────────────────────────────────────────────────────

type ForecastTransaction struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

type ForecastPoint struct {
	Date      string  `json:"date"`
	Predicted float64 `json:"predicted"`
	Lower     float64 `json:"lower"`
	Upper     float64 `json:"upper"`
}

type CategoryForecast struct {
	Category       string          `json:"category"`
	LabelRu        string          `json:"label_ru"`
	LabelKz        string          `json:"label_kz"`
	HorizonDays    int             `json:"horizon_days"`
	TotalPredicted float64         `json:"total_predicted"`
	Daily          []ForecastPoint `json:"daily"`
	Method         string          `json:"method"`
	Confidence     float64         `json:"confidence"`
}

type ForecastResponse struct {
	Forecasts   []CategoryForecast `json:"forecasts"`
	HorizonDays int                `json:"horizon_days"`
	RefDate     string             `json:"ref_date"`
}

func (c *Client) Forecast(ctx context.Context, transactions []ForecastTransaction, horizonDays int, refDate string) (*ForecastResponse, error) {
	payload := map[string]any{
		"transactions": transactions,
		"horizon_days": horizonDays,
	}
	if refDate != "" {
		payload["ref_date"] = refDate
	}
	return doPost[ForecastResponse](ctx, c, "/forecast", payload)
}

// ── Anomalies ─────────────────────────────────────────────────────────────────

type AnomalyPoint struct {
	Date          string  `json:"date"`
	Category      string  `json:"category"`
	LabelRu       string  `json:"label_ru"`
	LabelKz       string  `json:"label_kz"`
	Amount        float64 `json:"amount"`
	Mean          float64 `json:"mean"`
	Std           float64 `json:"std"`
	ZScore        float64 `json:"z_score"`
	Severity      string  `json:"severity"`
	Source        string  `json:"source"`
	ExpectedLower float64 `json:"expected_lower"`
	ExpectedUpper float64 `json:"expected_upper"`
}

type AnomalyResponse struct {
	Anomalies      []AnomalyPoint     `json:"anomalies"`
	TotalAnomalies int                `json:"total_anomalies"`
	Sensitivity    string             `json:"sensitivity"`
	ZThreshold     float64            `json:"z_threshold"`
	Method         string             `json:"method"`
	Stats          map[string]any     `json:"stats"`
}

func (c *Client) DetectAnomalies(ctx context.Context, transactions []ForecastTransaction, sensitivity string) (*AnomalyResponse, error) {
	if sensitivity == "" {
		sensitivity = "medium"
	}
	payload := map[string]any{
		"transactions": transactions,
		"sensitivity":  sensitivity,
	}
	return doPost[AnomalyResponse](ctx, c, "/anomalies", payload)
}

// doPost — общий helper для JSON POST запросов к ai-local-service.
func doPost[T any](ctx context.Context, c *Client, path string, payload any) (*T, error) {
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("localai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("localai: request failed: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("localai: read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("localai: status %d: %s", resp.StatusCode, string(raw))
	}

	var result T
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("localai: decode response: %w", err)
	}
	return &result, nil
}

// ParsedTransaction — структурированная транзакция из свободного текста.
type ParsedTransaction struct {
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Title         string  `json:"title"`
	Category      string  `json:"category"`
	CategoryLabel string  `json:"category_label"`
	Date          string  `json:"date"`
}

// ParseDebtUpdate — данные частичного возврата долга.
type ParseDebtUpdate struct {
	Type     string  `json:"type"`      // "half" | "amount" | "not_found"
	ReduceBy float64 `json:"reduce_by"`
	DebtID   *string `json:"debt_id,omitempty"`
}

// ParseMessageResponse — ответ от /parse-message.
type ParseMessageResponse struct {
	Intent           string             `json:"intent"`
	Response         string             `json:"response"`
	Transaction      *ParsedTransaction `json:"transaction,omitempty"`
	Counterparty     *string            `json:"counterparty,omitempty"`
	DebtDirection    *string            `json:"debt_direction,omitempty"`
	DebtUpdate       *ParseDebtUpdate   `json:"debt_update,omitempty"`
	Frequency        *string            `json:"frequency,omitempty"`
	ClarifyQuestions []string           `json:"clarify_questions,omitempty"`
	// Task
	TaskTitle    *string  `json:"task_title,omitempty"`
	TaskTime     *string  `json:"task_time,omitempty"`  // "HH:MM"
	TaskDay      *string  `json:"task_day,omitempty"`   // today | tomorrow | day_after_tomorrow
	TaskKeywords []string `json:"task_keywords,omitempty"`
	// Habit
	HabitTitle        *string `json:"habit_title,omitempty"`
	HabitDurationDays *int    `json:"habit_duration_days,omitempty"`
	HabitTime         *string `json:"habit_time,omitempty"`
	// Finance ecosystem
	GoalTitle     *string  `json:"goal_title,omitempty"`
	SavingsAmount *float64 `json:"savings_amount,omitempty"`
	SavingsPeriod *string  `json:"savings_period,omitempty"`
	AlertLimit    *float64 `json:"alert_limit,omitempty"`
	AlertPeriod   *string  `json:"alert_period,omitempty"`
}

// ParseMessage парсит свободный текст пользователя в структурированный intent.
// debtsContext — список долгов пользователя для поиска при update_debt; может быть nil.
func (c *Client) ParseMessage(ctx context.Context, message string, debtsContext []map[string]any) (*ParseMessageResponse, error) {
	payload := map[string]any{"message": message}
	if debtsContext != nil {
		payload["debts_context"] = debtsContext
	}
	return doPost[ParseMessageResponse](ctx, c, "/parse-message", payload)
}

// ── Insights ─────────────────────────────────────────────────────────────────

type InsightTransaction struct {
	Date     string  `json:"date"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"`
	Category string  `json:"category"`
}

type CategoryStat struct {
	Category  string  `json:"category"`
	LabelRu   string  `json:"label_ru"`
	LabelKz   string  `json:"label_kz"`
	Amount    float64 `json:"amount"`
	Pct       float64 `json:"pct"`
	TxCount   int     `json:"tx_count"`
	AvgPerTx  float64 `json:"avg_per_tx"`
}

type SummaryResponse struct {
	PeriodStart      string         `json:"period_start"`
	PeriodEnd        string         `json:"period_end"`
	TotalIncome      float64        `json:"total_income"`
	TotalExpense     float64        `json:"total_expense"`
	SavingsRate      float64        `json:"savings_rate"`
	Net              float64        `json:"net"`
	AvgDailyExpense  float64        `json:"avg_daily_expense"`
	TopCategories    []CategoryStat `json:"top_categories"`
	ExpenseTrend     string         `json:"expense_trend"`
	ExpenseTrendPct  float64        `json:"expense_trend_pct"`
}

type BudgetSuggestion struct {
	Category           string  `json:"category"`
	LabelRu            string  `json:"label_ru"`
	LabelKz            string  `json:"label_kz"`
	CurrentMonthlyAvg  float64 `json:"current_monthly_avg"`
	SuggestedLimit     float64 `json:"suggested_limit"`
	OverspendMonths    int     `json:"overspend_months"`
	Reason             string  `json:"reason"`
	Priority           string  `json:"priority"`
}

type BudgetSuggestResponse struct {
	Suggestions  []BudgetSuggestion `json:"suggestions"`
	LookbackDays int                `json:"lookback_days"`
	Percentile   float64            `json:"percentile"`
}

func (c *Client) SpendingSummary(ctx context.Context, transactions []InsightTransaction, periodStart, periodEnd string) (*SummaryResponse, error) {
	payload := map[string]any{"transactions": transactions}
	if periodStart != "" {
		payload["period_start"] = periodStart
	}
	if periodEnd != "" {
		payload["period_end"] = periodEnd
	}
	return doPost[SummaryResponse](ctx, c, "/insights/summary", payload)
}

func (c *Client) BudgetSuggestions(ctx context.Context, transactions []InsightTransaction, lookbackDays int, percentile float64) (*BudgetSuggestResponse, error) {
	payload := map[string]any{
		"transactions":  transactions,
		"lookback_days": lookbackDays,
		"percentile":    percentile,
	}
	return doPost[BudgetSuggestResponse](ctx, c, "/insights/budget-suggest", payload)
}

// Healthy проверяет, доступен ли ai-local-service.
func (c *Client) Healthy(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return false
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
