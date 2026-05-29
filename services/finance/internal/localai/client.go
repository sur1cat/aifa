package localai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 3 * time.Second},
	}
}

type CategoryResult struct {
	Category   string  `json:"category"`
	LabelRu    string  `json:"label_ru"`
	LabelKz    string  `json:"label_kz"`
	Confidence float64 `json:"confidence"`
	Confident  bool    `json:"confident"`
}

// CategorizeExpense возвращает категорию транзакции.
// При ошибке или низкой уверенности — возвращает ("", false).
func (c *Client) CategorizeExpense(ctx context.Context, text string) (category string, confident bool) {
	body, _ := json.Marshal(map[string]string{"text": text})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/categorize", bytes.NewReader(body))
	if err != nil {
		return "", false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return "", false
	}

	var result CategoryResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", false
	}
	if !result.Confident {
		return "", false
	}
	return result.Category, true
}
