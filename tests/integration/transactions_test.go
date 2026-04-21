//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestTransactionsCRUD(t *testing.T) {
	env := setupTestEnv(t)
	_, token := env.createTestUser(t)

	var txID string

	t.Run("create expense", func(t *testing.T) {
		body := `{"title":"Coffee","amount":5.50,"type":"expense","category":"food","date":"2026-04-21"}`
		w := env.doRequest("POST", "/api/v1/transactions", token, body)
		assertStatus(t, w, http.StatusCreated)

		data := assertData(t, w)
		txID = data["id"].(string)
		if data["title"] != "Coffee" {
			t.Fatalf("expected title 'Coffee', got %v", data["title"])
		}
		if data["type"] != "expense" {
			t.Fatalf("expected type 'expense', got %v", data["type"])
		}
		if data["amount"].(float64) != 5.50 {
			t.Fatalf("expected amount 5.50, got %v", data["amount"])
		}
	})

	t.Run("create income", func(t *testing.T) {
		body := `{"title":"Salary","amount":5000,"type":"income","category":"salary","date":"2026-04-21"}`
		w := env.doRequest("POST", "/api/v1/transactions", token, body)
		assertStatus(t, w, http.StatusCreated)
	})

	t.Run("get summary", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/transactions/summary?year=2026&month=4", token, "")
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		if data["income"].(float64) != 5000 {
			t.Fatalf("expected income 5000, got %v", data["income"])
		}
		if data["expense"].(float64) != 5.50 {
			t.Fatalf("expected expense 5.50, got %v", data["expense"])
		}
		if data["balance"].(float64) != 4994.50 {
			t.Fatalf("expected balance 4994.50, got %v", data["balance"])
		}
	})

	t.Run("list transactions by month", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/transactions?year=2026&month=4", token, "")
		assertStatus(t, w, http.StatusOK)

		data := assertDataList(t, w)
		if len(data) != 2 {
			t.Fatalf("expected 2 transactions, got %d", len(data))
		}
	})

	t.Run("update transaction", func(t *testing.T) {
		body := `{"title":"Fancy Coffee","amount":7.00}`
		w := env.doRequest("PUT", "/api/v1/transactions/"+txID, token, body)
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		if data["title"] != "Fancy Coffee" {
			t.Fatalf("expected updated title, got %v", data["title"])
		}
	})

	t.Run("delete transaction", func(t *testing.T) {
		w := env.doRequest("DELETE", "/api/v1/transactions/"+txID, token, "")
		assertStatus(t, w, http.StatusOK)

		w = env.doRequest("GET", "/api/v1/transactions/"+txID, token, "")
		assertStatus(t, w, http.StatusNotFound)
	})
}

func TestTransactionsValidation(t *testing.T) {
	env := setupTestEnv(t)
	_, token := env.createTestUser(t)

	t.Run("create without required fields returns 400", func(t *testing.T) {
		w := env.doRequest("POST", "/api/v1/transactions", token, `{"title":"test"}`)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("create with invalid type returns 400", func(t *testing.T) {
		body := `{"title":"test","amount":10,"type":"invalid"}`
		w := env.doRequest("POST", "/api/v1/transactions", token, body)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("create with negative amount returns 400", func(t *testing.T) {
		body := `{"title":"test","amount":-10,"type":"expense"}`
		w := env.doRequest("POST", "/api/v1/transactions", token, body)
		assertStatus(t, w, http.StatusBadRequest)
	})
}
