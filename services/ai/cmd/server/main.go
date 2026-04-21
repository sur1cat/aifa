package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sur1cat/aifa/ai-service/internal/config"
	"github.com/sur1cat/aifa/ai-service/internal/handler"
	"github.com/sur1cat/aifa/ai-service/internal/jwt"
	"github.com/sur1cat/aifa/ai-service/internal/middleware"
	"github.com/sur1cat/aifa/ai-service/internal/openai"
	"github.com/sur1cat/aifa/ai-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg, err := config.Load()
	if err != nil {
		log.Error("config load failed", "err", err)
		os.Exit(1)
	}

	if !cfg.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error("redis ping failed", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	validator := jwt.NewValidator(cfg.JWTSecret)
	blacklist := repository.NewBlacklist(rdb)
	openaiClient := openai.NewClient(cfg.OpenAIAPIKey, cfg.OpenAIModel)
	if !openaiClient.Configured() {
		log.Warn("OPENAI_API_KEY is empty — /ai/* will return 503")
	}

	authMW := middleware.NewAuth(validator, blacklist)
	aiHandler := handler.NewAIHandler(openaiClient)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(log))
	r.GET("/health", handler.Health)

	p := r.Group("", authMW.RequireAuth())
	p.POST("/ai/chat", aiHandler.Chat)
	p.POST("/ai/insights", aiHandler.GenerateInsight)
	p.POST("/ai/expense-analysis", aiHandler.GenerateExpenseAnalysis)
	p.POST("/ai/goal-to-habits", aiHandler.GenerateHabitsFromGoal)
	p.POST("/ai/goal-clarify", aiHandler.GenerateGoalQuestions)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  90 * time.Second,  // OpenAI can take a while
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("ai-service listening", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "err", err)
	}
}
