package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"habitflow/internal/handler"
	"habitflow/internal/middleware"
	"habitflow/internal/repository"
	"habitflow/pkg/ai"
	"habitflow/pkg/auth"
	"habitflow/pkg/config"
	"habitflow/pkg/database"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database connection
	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database")

	// Initialize repositories
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	goalRepo := repository.NewGoalRepository(pool)
	habitRepo := repository.NewHabitRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)
	txRepo := repository.NewTransactionRepository(pool)
	rtRepo := repository.NewRecurringTransactionRepository(pool)
	deviceTokenRepo := repository.NewDeviceTokenRepository(pool)
	savingsGoalRepo := repository.NewSavingsGoalRepository(pool)
	aiPendingRepo := repository.NewAIPendingCommandRepository(pool)

	// Initialize auth
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	googleVerifier := auth.NewGoogleTokenVerifier()
	appleVerifier := auth.NewAppleTokenVerifier()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize AI client
	aiClient := ai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(jwtManager, googleVerifier, appleVerifier, userRepo, tokenRepo)
	goalHandler := handler.NewGoalHandler(goalRepo)
	habitHandler := handler.NewHabitHandler(habitRepo)
	taskHandler := handler.NewTaskHandler(taskRepo)
	txHandler := handler.NewTransactionHandler(txRepo)
	rtHandler := handler.NewRecurringTransactionHandler(rtRepo, txRepo)
	aiHandler := handler.NewAIHandler(aiClient, txRepo, goalRepo, rtRepo, aiPendingRepo)
	pushHandler := handler.NewPushHandler(deviceTokenRepo)
	savingsGoalHandler := handler.NewSavingsGoalHandler(savingsGoalRepo)

	// Router
	r := gin.Default()

	// CORS middleware
	allowedOrigins := map[string]bool{
		"https://atoma.app":           true,
		"https://www.atoma.app":       true,
		"https://api.azamatbigali.online": true,
	}
	// Allow localhost in debug mode
	if cfg.Debug {
		allowedOrigins["http://localhost:3000"] = true
		allowedOrigins["http://localhost:8080"] = true
		allowedOrigins["http://127.0.0.1:3000"] = true
		allowedOrigins["http://127.0.0.1:8080"] = true
	}

	r.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", handler.Health)

	// Rate limiters
	generalRateLimit := middleware.RateLimit(100, time.Minute)       // 100 requests per minute per user
	strictRateLimit := middleware.StrictRateLimit(10, time.Minute)   // 10 requests per minute per IP (for auth)

	// API v1 routes
	v1 := r.Group("/api/v1")
	v1.Use(generalRateLimit)
	{
		v1.GET("/health", handler.Health)

		// Auth routes (public) with strict rate limiting
		authGroup := v1.Group("/auth")
		authGroup.Use(strictRateLimit)
		{
			// Social auth (Google & Apple only)
			authGroup.POST("/google", authHandler.GoogleSignIn)
			authGroup.POST("/apple", authHandler.AppleSignIn)

			// Token
			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.DELETE("/auth/account", authHandler.DeleteAccount)

			// Goals
			protected.GET("/goals", goalHandler.ListGoals)
			protected.POST("/goals", goalHandler.CreateGoal)
			protected.GET("/goals/:id", goalHandler.GetGoal)
			protected.PUT("/goals/:id", goalHandler.UpdateGoal)
			protected.DELETE("/goals/:id", goalHandler.DeleteGoal)

			// Habits
			protected.GET("/habits", habitHandler.ListHabits)
			protected.POST("/habits", habitHandler.CreateHabit)
			protected.GET("/habits/:id", habitHandler.GetHabit)
			protected.PUT("/habits/:id", habitHandler.UpdateHabit)
			protected.DELETE("/habits/:id", habitHandler.DeleteHabit)
			protected.POST("/habits/:id/toggle", habitHandler.ToggleCompletion)

			// Tasks
			protected.GET("/tasks", taskHandler.ListTasks)
			protected.POST("/tasks", taskHandler.CreateTask)
			protected.GET("/tasks/:id", taskHandler.GetTask)
			protected.PUT("/tasks/:id", taskHandler.UpdateTask)
			protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
			protected.POST("/tasks/:id/toggle", taskHandler.ToggleTask)

			// Transactions (Budget)
			protected.GET("/transactions", txHandler.ListTransactions)
			protected.POST("/transactions", txHandler.CreateTransaction)
			protected.GET("/transactions/summary", txHandler.GetSummary)
			protected.GET("/transactions/:id", txHandler.GetTransaction)
			protected.PUT("/transactions/:id", txHandler.UpdateTransaction)
			protected.DELETE("/transactions/:id", txHandler.DeleteTransaction)

			// Recurring Transactions
			protected.GET("/recurring-transactions", rtHandler.ListRecurringTransactions)
			protected.POST("/recurring-transactions", rtHandler.CreateRecurringTransaction)
			protected.GET("/recurring-transactions/projection", rtHandler.GetProjection)
			protected.POST("/recurring-transactions/process", rtHandler.ProcessRecurringTransactions)
			protected.GET("/recurring-transactions/:id", rtHandler.GetRecurringTransaction)
			protected.PUT("/recurring-transactions/:id", rtHandler.UpdateRecurringTransaction)
			protected.DELETE("/recurring-transactions/:id", rtHandler.DeleteRecurringTransaction)

			// Savings Goal
			protected.GET("/savings-goal", savingsGoalHandler.GetSavingsGoal)
			protected.POST("/savings-goal", savingsGoalHandler.SetSavingsGoal)
			protected.DELETE("/savings-goal", savingsGoalHandler.DeleteSavingsGoal)

			// AI
			protected.POST("/ai/chat", aiHandler.Chat)
			protected.POST("/ai/command", aiHandler.Command)
			protected.POST("/ai/insights", aiHandler.GenerateInsight)
			protected.POST("/ai/expense-analysis", aiHandler.GenerateExpenseAnalysis)
			protected.POST("/ai/goal-to-habits", aiHandler.GenerateHabitsFromGoal)
			protected.POST("/ai/goal-clarify", aiHandler.GenerateGoalQuestions)

			// Push notifications
			protected.POST("/push/register", pushHandler.RegisterToken)
			protected.POST("/push/unregister", pushHandler.UnregisterToken)
		}
	}

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting HabitFlow API on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
