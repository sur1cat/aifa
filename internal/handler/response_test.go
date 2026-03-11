package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRespondValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondValidationError(c, "invalid input")

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	errorObj, ok := response["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}

	if errorObj["code"] != ErrCodeValidation {
		t.Errorf("expected code %s, got %v", ErrCodeValidation, errorObj["code"])
	}

	if errorObj["message"] != "invalid input" {
		t.Errorf("expected message 'invalid input', got %v", errorObj["message"])
	}
}

func TestRespondUnauthorized(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondUnauthorized(c, "not logged in")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestRespondNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondNotFound(c, "resource not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestRespondOK(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	data := map[string]string{"name": "test"}
	respondOK(c, data)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	dataObj, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	if dataObj["name"] != "test" {
		t.Errorf("expected name 'test', got %v", dataObj["name"])
	}
}

func TestRespondCreated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondCreated(c, map[string]int{"id": 1})

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}
}

func TestRespondPaginated(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	items := []string{"a", "b", "c"}
	meta := PaginationMeta{Limit: 10, Offset: 0, Total: 3}
	respondPaginated(c, items, meta)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if _, ok := response["data"]; !ok {
		t.Error("expected data in response")
	}

	metaObj, ok := response["meta"].(map[string]interface{})
	if !ok {
		t.Fatal("expected meta object in response")
	}

	if metaObj["total"].(float64) != 3 {
		t.Errorf("expected total 3, got %v", metaObj["total"])
	}
}

func TestRespondRateLimited(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	respondRateLimited(c)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}
