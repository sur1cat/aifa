//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("root health returns 200", func(t *testing.T) {
		w := env.doRequest("GET", "/health", "", "")
		assertStatus(t, w, http.StatusOK)

		data := parseJSON(t, w.Body.Bytes())
		if data["status"] != "ok" {
			t.Fatalf("expected status ok, got %v", data["status"])
		}
		if data["database"] != "ok" {
			t.Fatalf("expected database ok, got %v", data["database"])
		}
	})

	t.Run("v1 health returns 200", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/health", "", "")
		assertStatus(t, w, http.StatusOK)
	})
}
