//go:build integration

package integration

import (
	"net/http"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	env := setupTestEnv(t)

	t.Run("protected endpoint without token returns 401", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/habits", "", "")
		assertStatus(t, w, http.StatusUnauthorized)

		resp := parseJSON(t, w.Body.Bytes())
		errObj := resp["error"].(map[string]interface{})
		if errObj["code"] != "MISSING_TOKEN" {
			t.Fatalf("expected MISSING_TOKEN, got %v", errObj["code"])
		}
	})

	t.Run("protected endpoint with invalid token returns 401", func(t *testing.T) {
		w := env.doRequest("GET", "/api/v1/habits", "invalid-token", "")
		assertStatus(t, w, http.StatusUnauthorized)
	})

	t.Run("protected endpoint with valid token returns 200", func(t *testing.T) {
		_, token := env.createTestUser(t)
		w := env.doRequest("GET", "/api/v1/habits", token, "")
		assertStatus(t, w, http.StatusOK)
	})
}
