package finance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type Debt struct {
	ID           string  `json:"id"`
	Counterparty string  `json:"counterparty"`
	Direction    string  `json:"direction"`
	Amount       float64 `json:"amount"`
	Settled      bool    `json:"settled"`
}

type CreateDebtRequest struct {
	Counterparty string  `json:"counterparty"`
	Direction    string  `json:"direction"`
	Amount       float64 `json:"amount"`
}

type PatchDebtRequest struct {
	ReduceBy *float64 `json:"reduce_by,omitempty"`
	Settle   bool     `json:"settle,omitempty"`
}

func (c *Client) CreateDebt(ctx context.Context, authHeader string, req CreateDebtRequest) error {
	return c.doJSON(ctx, http.MethodPost, "/debts", authHeader, req, nil)
}

func (c *Client) ListDebts(ctx context.Context, authHeader string, settled *bool) ([]Debt, error) {
	path := "/debts"
	if settled != nil {
		q := url.Values{}
		if *settled {
			q.Set("settled", "true")
		} else {
			q.Set("settled", "false")
		}
		path += "?" + q.Encode()
	}

	var debts []Debt
	if err := c.doJSON(ctx, http.MethodGet, path, authHeader, nil, &debts); err != nil {
		return nil, err
	}
	return debts, nil
}

func (c *Client) PatchDebt(ctx context.Context, authHeader, debtID string, req PatchDebtRequest) error {
	return c.doJSON(ctx, http.MethodPatch, "/debts/"+debtID, authHeader, req, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path, authHeader string, payload any, out any) error {
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if len(raw) == 0 {
			return fmt.Errorf("finance service returned %d", resp.StatusCode)
		}
		return fmt.Errorf("finance service returned %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	if out == nil || len(raw) == 0 {
		return nil
	}

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err == nil && len(envelope.Data) > 0 {
		return json.Unmarshal(envelope.Data, out)
	}
	return json.Unmarshal(raw, out)
}
