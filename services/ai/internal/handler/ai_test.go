package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sur1cat/aifa/ai-service/internal/localai"
	"github.com/sur1cat/aifa/ai-service/internal/openai"
)

type fakeOpenAIClient struct {
	chatFn           func(ctx context.Context, systemPrompt, userMessage string) (string, error)
	chatWithVisionFn func(ctx context.Context, systemPrompt, base64Image, mimeType string) (string, error)
	transcribeFn     func(ctx context.Context, audioData []byte, filename, language string) (string, error)
}

func (f *fakeOpenAIClient) Chat(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	if f.chatFn != nil {
		return f.chatFn(ctx, systemPrompt, userMessage)
	}
	return "", nil
}

func (f *fakeOpenAIClient) ChatWithVision(ctx context.Context, systemPrompt, base64Image, mimeType string) (string, error) {
	if f.chatWithVisionFn != nil {
		return f.chatWithVisionFn(ctx, systemPrompt, base64Image, mimeType)
	}
	return "", nil
}

func (f *fakeOpenAIClient) Transcribe(ctx context.Context, audioData []byte, filename, language string) (string, error) {
	if f.transcribeFn != nil {
		return f.transcribeFn(ctx, audioData, filename, language)
	}
	return "", nil
}

type fakeLocalAIClient struct {
	categorizeFn        func(ctx context.Context, text string) (*localai.CategoryResult, error)
	batchCategorizeFn   func(ctx context.Context, texts []string) ([]localai.CategoryResult, error)
	transcribeVoiceFn   func(ctx context.Context, audioData []byte, filename, language string) (*localai.VoiceTranscribeResult, error)
	scanReceiptFn       func(ctx context.Context, imageData []byte, filename string) (*localai.ReceiptScanResult, error)
	parseMessageFn      func(ctx context.Context, message string, debtsContext []map[string]any) (*localai.ParseMessageResponse, error)
	forecastFn          func(ctx context.Context, transactions []localai.ForecastTransaction, horizonDays int, refDate string) (*localai.ForecastResponse, error)
	detectAnomaliesFn   func(ctx context.Context, transactions []localai.ForecastTransaction, sensitivity string) (*localai.AnomalyResponse, error)
	spendingSummaryFn   func(ctx context.Context, transactions []localai.InsightTransaction, periodStart, periodEnd string) (*localai.SummaryResponse, error)
	budgetSuggestionsFn func(ctx context.Context, transactions []localai.InsightTransaction, lookbackDays int, percentile float64) (*localai.BudgetSuggestResponse, error)
}

func (f *fakeLocalAIClient) CategorizeExpense(ctx context.Context, text string) (*localai.CategoryResult, error) {
	return f.categorizeFn(ctx, text)
}

func (f *fakeLocalAIClient) BatchCategorizeExpenses(ctx context.Context, texts []string) ([]localai.CategoryResult, error) {
	if f.batchCategorizeFn != nil {
		return f.batchCategorizeFn(ctx, texts)
	}
	return nil, nil
}

func (f *fakeLocalAIClient) TranscribeVoice(ctx context.Context, audioData []byte, filename, language string) (*localai.VoiceTranscribeResult, error) {
	return f.transcribeVoiceFn(ctx, audioData, filename, language)
}

func (f *fakeLocalAIClient) ScanReceipt(ctx context.Context, imageData []byte, filename string) (*localai.ReceiptScanResult, error) {
	return f.scanReceiptFn(ctx, imageData, filename)
}

func (f *fakeLocalAIClient) ParseMessage(ctx context.Context, message string, debtsContext []map[string]any) (*localai.ParseMessageResponse, error) {
	if f.parseMessageFn != nil {
		return f.parseMessageFn(ctx, message, debtsContext)
	}
	return nil, nil
}

func (f *fakeLocalAIClient) Forecast(ctx context.Context, transactions []localai.ForecastTransaction, horizonDays int, refDate string) (*localai.ForecastResponse, error) {
	if f.forecastFn != nil {
		return f.forecastFn(ctx, transactions, horizonDays, refDate)
	}
	return nil, nil
}

func (f *fakeLocalAIClient) DetectAnomalies(ctx context.Context, transactions []localai.ForecastTransaction, sensitivity string) (*localai.AnomalyResponse, error) {
	if f.detectAnomaliesFn != nil {
		return f.detectAnomaliesFn(ctx, transactions, sensitivity)
	}
	return nil, nil
}

func (f *fakeLocalAIClient) SpendingSummary(ctx context.Context, transactions []localai.InsightTransaction, periodStart, periodEnd string) (*localai.SummaryResponse, error) {
	if f.spendingSummaryFn != nil {
		return f.spendingSummaryFn(ctx, transactions, periodStart, periodEnd)
	}
	return nil, nil
}

func (f *fakeLocalAIClient) BudgetSuggestions(ctx context.Context, transactions []localai.InsightTransaction, lookbackDays int, percentile float64) (*localai.BudgetSuggestResponse, error) {
	if f.budgetSuggestionsFn != nil {
		return f.budgetSuggestionsFn(ctx, transactions, lookbackDays, percentile)
	}
	return nil, nil
}

func setupRouter(h *AIHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/categorize", h.CategorizeExpense)
	r.POST("/voice/transcribe", h.TranscribeVoice)
	r.POST("/receipt/scan", h.ScanReceipt)
	return r
}

func decodeData(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload
}

func makeMultipartRequest(t *testing.T, path, field, filename string, content []byte, fields map[string]string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	for k, v := range fields {
		if err := w.WriteField(k, v); err != nil {
			t.Fatalf("write field: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, path, &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestCategorizeExpenseLocalConfident(t *testing.T) {
	h := NewAIHandler(&fakeOpenAIClient{}, &fakeLocalAIClient{
		categorizeFn: func(ctx context.Context, text string) (*localai.CategoryResult, error) {
			return &localai.CategoryResult{
				Text: "coffee", Category: "cafe", LabelRu: "Кафе", LabelKz: "Кафе", Confidence: 0.91, Confident: true,
			}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/categorize", bytes.NewBufferString(`{"text":"coffee"}`))
	req.Header.Set("Content-Type", "application/json")
	setupRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	payload := decodeData(t, rec)
	data := payload["data"].(map[string]any)
	if data["source"] != "local" {
		t.Fatalf("expected local source, got %v", data["source"])
	}
}

func TestCategorizeExpenseFallsBackToOpenAI(t *testing.T) {
	h := NewAIHandler(&fakeOpenAIClient{
		chatFn: func(ctx context.Context, systemPrompt, userMessage string) (string, error) {
			return `{"category":"shopping","label_ru":"Покупки","label_kz":"Сатып алулар"}`, nil
		},
	}, &fakeLocalAIClient{
		categorizeFn: func(ctx context.Context, text string) (*localai.CategoryResult, error) {
			return &localai.CategoryResult{
				Text: "weird store", Category: "other", LabelRu: "Другое", LabelKz: "Басқа", Confidence: 0.31, Confident: false,
			}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/categorize", bytes.NewBufferString(`{"text":"weird store"}`))
	req.Header.Set("Content-Type", "application/json")
	setupRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	payload := decodeData(t, rec)
	data := payload["data"].(map[string]any)
	if data["source"] != "gpt4" {
		t.Fatalf("expected gpt4 source, got %v", data["source"])
	}
}

func TestTranscribeVoiceUsesLocalWhenAvailable(t *testing.T) {
	h := NewAIHandler(&fakeOpenAIClient{
		transcribeFn: func(ctx context.Context, audioData []byte, filename, language string) (string, error) {
			t.Fatal("openai transcribe should not be called")
			return "", nil
		},
	}, &fakeLocalAIClient{
		transcribeVoiceFn: func(ctx context.Context, audioData []byte, filename, language string) (*localai.VoiceTranscribeResult, error) {
			return &localai.VoiceTranscribeResult{
				Transcript: "купил кофе", Category: "cafe", LabelRu: "Кафе", LabelKz: "Кафе", Confidence: 0.9, Language: "ru",
			}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := makeMultipartRequest(t, "/voice/transcribe", "audio", "voice.m4a", []byte("audio"), map[string]string{"language": "ru"})
	setupRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTranscribeVoiceFallsBackToOpenAI(t *testing.T) {
	h := NewAIHandler(&fakeOpenAIClient{
		transcribeFn: func(ctx context.Context, audioData []byte, filename, language string) (string, error) {
			return "купил кофе за 1500 тг", nil
		},
		chatFn: func(ctx context.Context, systemPrompt, userMessage string) (string, error) {
			return `{"amount":1500,"currency":"KZT","description":"кофе","category":"cafe","label_ru":"Кафе","label_kz":"Кафе","confidence":0.82}`, nil
		},
	}, &fakeLocalAIClient{
		transcribeVoiceFn: func(ctx context.Context, audioData []byte, filename, language string) (*localai.VoiceTranscribeResult, error) {
			return nil, errors.New("timeout")
		},
	})

	rec := httptest.NewRecorder()
	req := makeMultipartRequest(t, "/voice/transcribe", "audio", "voice.m4a", []byte("audio"), map[string]string{"language": "ru"})
	setupRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	payload := decodeData(t, rec)
	data := payload["data"].(map[string]any)
	if data["transcript"] != "купил кофе за 1500 тг" {
		t.Fatalf("unexpected transcript: %v", data["transcript"])
	}
}

func TestScanReceiptUsesLocalWhenAvailable(t *testing.T) {
	amount := 14.25
	date := "2026-05-24"
	h := NewAIHandler(&fakeOpenAIClient{}, &fakeLocalAIClient{
		scanReceiptFn: func(ctx context.Context, imageData []byte, filename string) (*localai.ReceiptScanResult, error) {
			return &localai.ReceiptScanResult{
				Amount: &amount, Currency: "USD", Date: &date, Merchant: "ALPHA STORE", Category: "shopping",
			}, nil
		},
	})

	rec := httptest.NewRecorder()
	req := makeMultipartRequest(t, "/receipt/scan", "image", "receipt.jpg", []byte("img"), nil)
	setupRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestScanReceiptFallsBackToOpenAIVision(t *testing.T) {
	h := NewAIHandler(&fakeOpenAIClient{
		chatWithVisionFn: func(ctx context.Context, systemPrompt, base64Image, mimeType string) (string, error) {
			return `{"amount":785.00,"currency":"RUB","date":"2025-07-29T16:36:00","merchant":"АВТОКАФЕ","category":"food","label_ru":"Еда","label_kz":"Тамақ","items":["морс"],"confidence":0.9,"raw_total":"785.00"}`, nil
		},
	}, &fakeLocalAIClient{
		scanReceiptFn: func(ctx context.Context, imageData []byte, filename string) (*localai.ReceiptScanResult, error) {
			return nil, errors.New("local ocr timeout")
		},
	})

	rec := httptest.NewRecorder()
	req := makeMultipartRequest(t, "/receipt/scan", "image", "receipt.jpg", []byte("img"), nil)
	setupRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	payload := decodeData(t, rec)
	data := payload["data"].(map[string]any)
	if data["merchant"] != "АВТОКАФЕ" {
		t.Fatalf("unexpected merchant: %v", data["merchant"])
	}
}

func TestScanReceiptReturns503WhenFallbackNotConfigured(t *testing.T) {
	h := NewAIHandler(&fakeOpenAIClient{
		chatWithVisionFn: func(ctx context.Context, systemPrompt, base64Image, mimeType string) (string, error) {
			return "", openai.ErrNotConfigured
		},
	}, &fakeLocalAIClient{
		scanReceiptFn: func(ctx context.Context, imageData []byte, filename string) (*localai.ReceiptScanResult, error) {
			return nil, errors.New("local ocr timeout")
		},
	})

	rec := httptest.NewRecorder()
	req := makeMultipartRequest(t, "/receipt/scan", "image", "receipt.jpg", []byte("img"), nil)
	setupRouter(h).ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", rec.Code, rec.Body.String())
	}
}
