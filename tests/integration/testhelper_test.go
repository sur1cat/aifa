//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"habitflow/internal/handler"
	"habitflow/internal/middleware"
	"habitflow/internal/repository"
	"habitflow/pkg/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://habitflow:habitflow@localhost:5432/habitflow_test?sslmode=disable"
	}

	ctx := context.Background()
	var err error
	testPool, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to test database: %v\n", err)
		os.Exit(1)
	}
	defer testPool.Close()

	if err := testPool.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to ping test database: %v\n", err)
		os.Exit(1)
	}

	if err := applyMigrations(ctx, testPool); err != nil {
		fmt.Fprintf(os.Stderr, "failed to apply migrations: %v\n", err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

type testEnv struct {
	pool       *pgxpool.Pool
	router     *gin.Engine
	jwtManager *auth.JWTManager
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	cleanupDB(t, testPool)

	jwtManager := auth.NewJWTManager("test-secret-minimum-32-characters-long", 24*time.Hour, 7*24*time.Hour)
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	habitRepo := repository.NewHabitRepository(testPool)
	taskRepo := repository.NewTaskRepository(testPool)
	txRepo := repository.NewTransactionRepository(testPool)
	goalRepo := repository.NewGoalRepository(testPool)

	habitHandler := handler.NewHabitHandler(habitRepo)
	taskHandler := handler.NewTaskHandler(taskRepo)
	txHandler := handler.NewTransactionHandler(txRepo)
	goalHandler := handler.NewGoalHandler(goalRepo)

	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/health", handler.HealthWithDB(testPool))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", handler.HealthWithDB(testPool))

		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/habits", habitHandler.ListHabits)
			protected.POST("/habits", habitHandler.CreateHabit)
			protected.GET("/habits/:id", habitHandler.GetHabit)
			protected.PUT("/habits/:id", habitHandler.UpdateHabit)
			protected.DELETE("/habits/:id", habitHandler.DeleteHabit)
			protected.POST("/habits/:id/toggle", habitHandler.ToggleCompletion)

			protected.GET("/tasks", taskHandler.ListTasks)
			protected.POST("/tasks", taskHandler.CreateTask)
			protected.GET("/tasks/:id", taskHandler.GetTask)
			protected.PUT("/tasks/:id", taskHandler.UpdateTask)
			protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
			protected.POST("/tasks/:id/toggle", taskHandler.ToggleTask)

			protected.GET("/transactions", txHandler.ListTransactions)
			protected.POST("/transactions", txHandler.CreateTransaction)
			protected.GET("/transactions/summary", txHandler.GetSummary)
			protected.GET("/transactions/:id", txHandler.GetTransaction)
			protected.PUT("/transactions/:id", txHandler.UpdateTransaction)
			protected.DELETE("/transactions/:id", txHandler.DeleteTransaction)

			protected.GET("/goals", goalHandler.ListGoals)
			protected.POST("/goals", goalHandler.CreateGoal)
			protected.GET("/goals/:id", goalHandler.GetGoal)
			protected.PUT("/goals/:id", goalHandler.UpdateGoal)
			protected.DELETE("/goals/:id", goalHandler.DeleteGoal)
		}
	}

	return &testEnv{
		pool:       testPool,
		router:     r,
		jwtManager: jwtManager,
	}
}

func applyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// Drop all tables for a clean slate
	_, _ = pool.Exec(ctx, `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)

	migrationsDir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var upFiles []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			upFiles = append(upFiles, e.Name())
		}
	}
	sort.Strings(upFiles)

	for _, f := range upFiles {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("apply %s: %w", f, err)
		}
	}
	return nil
}

func cleanupDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	tables := []string{
		"invalidated_tokens",
		"habit_completions",
		"habits",
		"tasks",
		"transactions",
		"recurring_transactions",
		"savings_goals",
		"device_tokens",
		"goals",
		"users",
	}
	for _, table := range tables {
		_, _ = pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
	}
}

func (e *testEnv) createTestUser(t *testing.T) (uuid.UUID, string) {
	t.Helper()

	userID := uuid.New()
	ctx := context.Background()
	_, err := e.pool.Exec(ctx,
		`INSERT INTO users (id, email, name, auth_provider, provider_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
		userID, "test@example.com", "Test User", "google", "google_"+userID.String(),
	)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	tokens, err := e.jwtManager.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	return userID, tokens.AccessToken
}

func (e *testEnv) doRequest(method, path, token string, body string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	e.router.ServeHTTP(w, req)
	return w
}

func assertStatus(t *testing.T, w *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if w.Code != expected {
		t.Fatalf("expected status %d, got %d: %s", expected, w.Code, w.Body.String())
	}
}

func assertData(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	resp := parseJSON(t, w.Body.Bytes())
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data object in response, got %v", resp)
	}
	return data
}

func assertDataList(t *testing.T, w *httptest.ResponseRecorder) []interface{} {
	t.Helper()
	resp := parseJSON(t, w.Body.Bytes())
	data, ok := resp["data"].([]interface{})
	if !ok {
		t.Fatalf("expected data array in response, got %v", resp)
	}
	return data
}

func parseJSON(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("failed to parse JSON: %v\nbody: %s", err, string(body))
	}
	return result
}
