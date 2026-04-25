package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"
)

const (
	endpoint         = "https://api.openai.com/v1/chat/completions"
	whisperEndpoint  = "https://api.openai.com/v1/audio/transcriptions"
)

var ErrNotConfigured = errors.New("OpenAI API key is not configured")

type Client struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (c *Client) Configured() bool { return c.apiKey != "" }

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// visionMessage используется для Vision API — content может быть массивом.
type visionMessage struct {
	Role    string        `json:"role"`
	Content []visionPart `json:"content"`
}

type visionPart struct {
	Type     string            `json:"type"`
	Text     string            `json:"text,omitempty"`
	ImageURL *visionImageURL   `json:"image_url,omitempty"`
}

type visionImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail"` // "low" | "high" | "auto"
}

type chatRequest struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	MaxCompletionTokens int       `json:"max_completion_tokens,omitempty"`
	Temperature         float64   `json:"temperature,omitempty"`
}

type visionRequest struct {
	Model               string          `json:"model"`
	Messages            []visionMessage `json:"messages"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"`
	Temperature         float64         `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type errorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Chat sends a system+user prompt and returns the first choice's text.
func (c *Client) Chat(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	if !c.Configured() {
		return "", ErrNotConfigured
	}

	body, err := json.Marshal(chatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		MaxCompletionTokens: 1000,
		Temperature:         0.7,
	})
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var e errorResponse
		if err := json.Unmarshal(raw, &e); err == nil && e.Error.Message != "" {
			return "", fmt.Errorf("openai: %s", e.Error.Message)
		}
		return "", fmt.Errorf("openai status %d", resp.StatusCode)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("openai returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

// ChatWithVision отправляет системный промпт + изображение (base64) в Vision API.
func (c *Client) ChatWithVision(ctx context.Context, systemPrompt, base64Image, mimeType string) (string, error) {
	if !c.Configured() {
		return "", ErrNotConfigured
	}

	body, err := json.Marshal(visionRequest{
		Model: c.model,
		Messages: []visionMessage{
			{
				Role: "system",
				Content: []visionPart{
					{Type: "text", Text: systemPrompt},
				},
			},
			{
				Role: "user",
				Content: []visionPart{
					{
						Type: "image_url",
						ImageURL: &visionImageURL{
							URL:    "data:" + mimeType + ";base64," + base64Image,
							Detail: "high",
						},
					},
				},
			},
		},
		MaxCompletionTokens: 500,
		Temperature:         0.1,
	})
	if err != nil {
		return "", fmt.Errorf("marshal vision request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build vision request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send vision request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read vision response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var e errorResponse
		if err := json.Unmarshal(raw, &e); err == nil && e.Error.Message != "" {
			return "", fmt.Errorf("openai vision: %s", e.Error.Message)
		}
		return "", fmt.Errorf("openai vision status %d", resp.StatusCode)
	}

	var parsed chatResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", fmt.Errorf("decode vision response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", errors.New("openai vision returned no choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

// Transcribe отправляет аудиофайл в Whisper API и возвращает транскрипцию.
// filename нужен Whisper для определения формата (например "audio.m4a").
func (c *Client) Transcribe(ctx context.Context, audioData []byte, filename, language string) (string, error) {
	if !c.Configured() {
		return "", ErrNotConfigured
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err = fw.Write(audioData); err != nil {
		return "", fmt.Errorf("write audio data: %w", err)
	}
	_ = w.WriteField("model", "whisper-1")
	_ = w.WriteField("response_format", "text")
	if language != "" {
		_ = w.WriteField("language", language)
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, whisperEndpoint, &buf)
	if err != nil {
		return "", fmt.Errorf("build whisper request: %w", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("send whisper request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read whisper response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		var e errorResponse
		if err := json.Unmarshal(raw, &e); err == nil && e.Error.Message != "" {
			return "", fmt.Errorf("whisper: %s", e.Error.Message)
		}
		return "", fmt.Errorf("whisper status %d: %s", resp.StatusCode, raw)
	}

	return string(raw), nil
}
