//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHabitsCRUD(t *testing.T) {
	env := setupTestEnv(t)
	_, token := env.createTestUser(t)

	var habitID string

	t.Run("create habit", func(t *testing.T) {
		body := `{"title":"Read 30 pages","icon":"📖","color":"blue","period":"daily"}`
		w := env.doRequest("POST", "/api/v1/habits", token, body)
		assertStatus(t, w, http.StatusCreated)

		data := assertData(t, w)
		habitID = data["id"].(string)
		if data["title"] != "Read 30 pages" {
			t.Fatalf("expected title 'Read 30 pages', got %v", data["title"])
		}
		if data["period"] != "daily" {
			t.Fatalf("expected period 'daily', got %v", data["period"])
		}
	})

	t.Run("list habits", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/habits", token, "")
		assertStatus(t, w, http.StatusOK)

		data := assertDataList(t, w)
		if len(data) != 1 {
			t.Fatalf("expected 1 habit, got %d", len(data))
		}
	})

	t.Run("get habit by ID", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/habits/"+habitID, token, "")
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		if data["id"] != habitID {
			t.Fatalf("expected id %s, got %v", habitID, data["id"])
		}
	})

	t.Run("update habit", func(t *testing.T) {
		body := `{"title":"Read 50 pages","color":"green"}`
		w := env.doRequest("PUT", "/api/v1/habits/"+habitID, token, body)
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		if data["title"] != "Read 50 pages" {
			t.Fatalf("expected title 'Read 50 pages', got %v", data["title"])
		}
	})

	t.Run("toggle habit completion", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		body := fmt.Sprintf(`{"date":"%s"}`, today)
		w := env.doRequest("POST", "/api/v1/habits/"+habitID+"/toggle", token, body)
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		dates := data["completed_dates"].([]interface{})
		if len(dates) != 1 {
			t.Fatalf("expected 1 completed date, got %d", len(dates))
		}
	})

	t.Run("toggle habit completion again (untoggle)", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		body := fmt.Sprintf(`{"date":"%s"}`, today)
		w := env.doRequest("POST", "/api/v1/habits/"+habitID+"/toggle", token, body)
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		dates := data["completed_dates"].([]interface{})
		if len(dates) != 0 {
			t.Fatalf("expected 0 completed dates after untoggle, got %d", len(dates))
		}
	})

	t.Run("delete habit", func(t *testing.T) {
		w := env.doRequest("DELETE", "/api/v1/habits/"+habitID, token, "")
		assertStatus(t, w, http.StatusOK)

		w = env.doRequest("GET", "/api/v1/habits/"+habitID, token, "")
		assertStatus(t, w, http.StatusNotFound)
	})
}

func TestHabitsOwnership(t *testing.T) {
	env := setupTestEnv(t)
	_, token1 := env.createTestUser(t)
	_, token2 := env.createTestUser(t)

	w := env.doRequest("POST", "/api/v1/habits", token1, `{"title":"User1 habit","period":"daily"}`)
	assertStatus(t, w, http.StatusCreated)
	habitID := assertData(t, w)["id"].(string)

	t.Run("user cannot access another user habit", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/habits/"+habitID, token2, "")
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("user cannot delete another user habit", func(t *testing.T) {
		w := env.doRequest("DELETE", "/api/v1/habits/"+habitID, token2, "")
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("user list shows only own habits", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/habits", token2, "")
		assertStatus(t, w, http.StatusOK)
		data := assertDataList(t, w)
		if len(data) != 0 {
			t.Fatalf("expected 0 habits for user2, got %d", len(data))
		}
	})
}
