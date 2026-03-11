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

	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Connected to database")

	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	goalRepo := repository.NewGoalRepository(pool)
	habitRepo := repository.NewHabitRepository(pool)
	taskRepo := repository.NewTaskRepository(pool)
	txRepo := repository.NewTransactionRepository(pool)
	rtRepo := repository.NewRecurringTransactionRepository(pool)
	deviceTokenRepo := repository.NewDeviceTokenRepository(pool)
	savingsGoalRepo := repository.NewSavingsGoalRepository(pool)

	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	googleVerifier := auth.NewGoogleTokenVerifier()
	appleVerifier := auth.NewAppleTokenVerifier()
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	aiClient := ai.NewClient(cfg.OpenAI.APIKey, cfg.OpenAI.Model)

	authHandler := handler.NewAuthHandler(jwtManager, googleVerifier, appleVerifier, userRepo, tokenRepo)
	goalHandler := handler.NewGoalHandler(goalRepo)
	habitHandler := handler.NewHabitHandler(habitRepo)
	taskHandler := handler.NewTaskHandler(taskRepo)
	txHandler := handler.NewTransactionHandler(txRepo)
	rtHandler := handler.NewRecurringTransactionHandler(rtRepo, txRepo)
	aiHandler := handler.NewAIHandler(aiClient)
	pushHandler := handler.NewPushHandler(deviceTokenRepo)
	savingsGoalHandler := handler.NewSavingsGoalHandler(savingsGoalRepo)

	r := gin.Default()

	allowedOrigins := map[string]bool{
		"https://aifa.app":           true,
		"https://www.aifa.app":       true,
	}

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

	r.GET("/health", handler.Health)

	generalRateLimit := middleware.RateLimit(100, time.Minute)
	strictRateLimit := middleware.StrictRateLimit(10, time.Minute)

	v1 := r.Group("/api/v1")
	v1.Use(generalRateLimit)
	{
		v1.GET("/health", handler.Health)

		authGroup := v1.Group("/auth")
		authGroup.Use(strictRateLimit)
		{

			authGroup.POST("/google", authHandler.GoogleSignIn)
			authGroup.POST("/apple", authHandler.AppleSignIn)

			authGroup.POST("/refresh", authHandler.RefreshToken)
		}

		protected := v1.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			protected.GET("/auth/me", authHandler.GetCurrentUser)
			protected.POST("/auth/logout", authHandler.Logout)
			protected.DELETE("/auth/account", authHandler.DeleteAccount)

			protected.GET("/goals", goalHandler.ListGoals)
			protected.POST("/goals", goalHandler.CreateGoal)
			protected.GET("/goals/:id", goalHandler.GetGoal)
			protected.PUT("/goals/:id", goalHandler.UpdateGoal)
			protected.DELETE("/goals/:id", goalHandler.DeleteGoal)

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

			protected.GET("/recurring-transactions", rtHandler.ListRecurringTransactions)
			protected.POST("/recurring-transactions", rtHandler.CreateRecurringTransaction)
			protected.GET("/recurring-transactions/projection", rtHandler.GetProjection)
			protected.POST("/recurring-transactions/process", rtHandler.ProcessRecurringTransactions)
			protected.GET("/recurring-transactions/:id", rtHandler.GetRecurringTransaction)
			protected.PUT("/recurring-transactions/:id", rtHandler.UpdateRecurringTransaction)
			protected.DELETE("/recurring-transactions/:id", rtHandler.DeleteRecurringTransaction)

			protected.GET("/savings-goal", savingsGoalHandler.GetSavingsGoal)
			protected.POST("/savings-goal", savingsGoalHandler.SetSavingsGoal)
			protected.DELETE("/savings-goal", savingsGoalHandler.DeleteSavingsGoal)

			protected.POST("/ai/chat", aiHandler.Chat)
			protected.POST("/ai/insights", aiHandler.GenerateInsight)
			protected.POST("/ai/expense-analysis", aiHandler.GenerateExpenseAnalysis)
			protected.POST("/ai/goal-to-habits", aiHandler.GenerateHabitsFromGoal)
			protected.POST("/ai/goal-clarify", aiHandler.GenerateGoalQuestions)

			protected.POST("/push/register", pushHandler.RegisterToken)
			protected.POST("/push/unregister", pushHandler.UnregisterToken)
		}
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Starting HabitFlow API on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
