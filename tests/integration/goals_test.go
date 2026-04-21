//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestGoalsCRUD(t *testing.T) {
	env := setupTestEnv(t)
	_, token := env.createTestUser(t)

	var goalID string

	t.Run("create goal", func(t *testing.T) {
		body := `{"title":"Learn Go","icon":"🎯","target_value":100,"unit":"hours","deadline":"2026-12-31T00:00:00Z"}`
		w := env.doRequest("POST", "/api/v1/goals", token, body)
		assertStatus(t, w, http.StatusCreated)

		data := assertData(t, w)
		goalID = data["id"].(string)
		if data["title"] != "Learn Go" {
			t.Fatalf("expected title 'Learn Go', got %v", data["title"])
		}
	})

	t.Run("list goals", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/goals", token, "")
		assertStatus(t, w, http.StatusOK)

		data := assertDataList(t, w)
		if len(data) != 1 {
			t.Fatalf("expected 1 goal, got %d", len(data))
		}
	})

	t.Run("get goal", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/goals/"+goalID, token, "")
		assertStatus(t, w, http.StatusOK)
	})

	t.Run("update goal", func(t *testing.T) {
		w := env.doRequest("PUT", "/api/v1/goals/"+goalID, token, `{"title":"Master Go"}`)
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		if data["title"] != "Master Go" {
			t.Fatalf("expected updated title, got %v", data["title"])
		}
	})

	t.Run("delete goal", func(t *testing.T) {
		w := env.doRequest("DELETE", "/api/v1/goals/"+goalID, token, "")
		assertStatus(t, w, http.StatusOK)

		w = env.doRequest("GET", "/api/v1/goals/"+goalID, token, "")
		assertStatus(t, w, http.StatusNotFound)
	})
}

func TestGoalsOwnership(t *testing.T) {
	env := setupTestEnv(t)
	_, token1 := env.createTestUser(t)
	_, token2 := env.createTestUser(t)

	w := env.doRequest("POST", "/api/v1/goals", token1, `{"title":"User1 goal"}`)
	assertStatus(t, w, http.StatusCreated)
	goalID := assertData(t, w)["id"].(string)

	t.Run("user cannot access another user goal", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/goals/"+goalID, token2, "")
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("user cannot update another user goal", func(t *testing.T) {
		w := env.doRequest("PUT", "/api/v1/goals/"+goalID, token2, `{"title":"hacked"}`)
		assertStatus(t, w, http.StatusNotFound)
	})
}
