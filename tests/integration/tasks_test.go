//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestTasksCRUD(t *testing.T) {
	env := setupTestEnv(t)
	_, token := env.createTestUser(t)

	var taskID string

	t.Run("create task", func(t *testing.T) {
		body := `{"title":"Buy groceries","priority":"high","due_date":"2026-04-21"}`
		w := env.doRequest("POST", "/api/v1/tasks", token, body)
		assertStatus(t, w, http.StatusCreated)

		data := assertData(t, w)
		taskID = data["id"].(string)
		if data["title"] != "Buy groceries" {
			t.Fatalf("expected title 'Buy groceries', got %v", data["title"])
		}
		if data["priority"] != "high" {
			t.Fatalf("expected priority 'high', got %v", data["priority"])
		}
	})

	t.Run("list tasks", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/tasks", token, "")
		assertStatus(t, w, http.StatusOK)

		data := assertDataList(t, w)
		if len(data) != 1 {
			t.Fatalf("expected 1 task, got %d", len(data))
		}
	})

	t.Run("toggle task completion", func(t *testing.T) {
		w := env.doRequest("POST", "/api/v1/tasks/"+taskID+"/toggle", token, "")
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		if completed, ok := data["is_completed"].(bool); !ok || !completed {
			t.Fatalf("expected is_completed true after toggle, got %v", data["is_completed"])
		}
	})

	t.Run("update task", func(t *testing.T) {
		body := `{"title":"Buy organic groceries"}`
		w := env.doRequest("PUT", "/api/v1/tasks/"+taskID, token, body)
		assertStatus(t, w, http.StatusOK)

		data := assertData(t, w)
		if data["title"] != "Buy organic groceries" {
			t.Fatalf("expected updated title, got %v", data["title"])
		}
	})

	t.Run("delete task", func(t *testing.T) {
		w := env.doRequest("DELETE", "/api/v1/tasks/"+taskID, token, "")
		assertStatus(t, w, http.StatusOK)

		w = env.doRequest("GET", "/api/v1/tasks/"+taskID, token, "")
		assertStatus(t, w, http.StatusNotFound)
	})
}

func TestTasksValidation(t *testing.T) {
	env := setupTestEnv(t)
	_, token := env.createTestUser(t)

	t.Run("create task without title returns 400", func(t *testing.T) {
		w := env.doRequest("POST", "/api/v1/tasks", token, `{"priority":"high"}`)
		assertStatus(t, w, http.StatusBadRequest)
	})

	t.Run("get nonexistent task returns 404", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/tasks/00000000-0000-0000-0000-000000000000", token, "")
		assertStatus(t, w, http.StatusNotFound)
	})

	t.Run("invalid UUID returns 400", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/tasks/not-a-uuid", token, "")
		assertStatus(t, w, http.StatusBadRequest)
	})
}
