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

	"github.com/sur1cat/aifa/auth-service/internal/config"
	"github.com/sur1cat/aifa/auth-service/internal/events"
	"github.com/sur1cat/aifa/auth-service/internal/handler"
	"github.com/sur1cat/aifa/auth-service/internal/jwt"
	"github.com/sur1cat/aifa/auth-service/internal/middleware"
	"github.com/sur1cat/aifa/auth-service/internal/migrate"
	"github.com/sur1cat/aifa/auth-service/internal/oauth"
	"github.com/sur1cat/aifa/auth-service/internal/otp"
	"github.com/sur1cat/aifa/auth-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
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

	pool, rdb, pub, err := connectInfra(ctx, cfg)
	if err != nil {
		log.Error("infrastructure connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	defer rdb.Close()
	defer pub.Close()

	if err := migrate.Run(ctx, pool, "migrations", "auth"); err != nil {
		log.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	jwtMgr := jwt.NewManager(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)
	userRepo := repository.NewUserRepository(pool)
	blacklist := repository.NewBlacklist(rdb)

	authMW := middleware.NewAuth(jwtMgr, blacklist)
	authHandler := handler.NewAuthHandler(
		jwtMgr,
		oauth.NewGoogleVerifier(cfg.GoogleClientID),
		oauth.NewAppleVerifier(cfg.AppleClientID),
		userRepo,
		blacklist,
		pub,
	)

	otpStore := otp.NewStore(rdb)
	otpHandler := handler.NewOTPHandler(authHandler, otpStore, cfg.Debug)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(log))

	r.GET("/health", handler.Health)
	r.POST("/auth/google", authHandler.GoogleSignIn)
	r.POST("/auth/apple", authHandler.AppleSignIn)
	r.POST("/auth/otp/send", otpHandler.SendOTP)
	r.POST("/auth/otp/verify", otpHandler.VerifyOTP)
	r.POST("/auth/refresh", authHandler.RefreshToken)

	protected := r.Group("", authMW.RequireAuth())
	protected.GET("/auth/me", authHandler.GetCurrentUser)
	protected.POST("/auth/logout", authHandler.Logout)
	protected.DELETE("/auth/account", authHandler.DeleteAccount)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("auth-service listening", "port", cfg.Port)
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

// connectInfra establishes Postgres, Redis, and NATS connections in parallel.
// On failure any successfully-opened connections are closed before returning.
func connectInfra(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, *redis.Client, *events.Publisher, error) {
	var (
		pool *pgxpool.Pool
		rdb  *redis.Client
		pub  *events.Publisher
	)
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		p, err := newPgPool(gCtx, cfg.DatabaseURL)
		if err != nil {
			return err
		}
		pool = p
		return nil
	})

	g.Go(func() error {
		c := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr, Password: cfg.RedisPassword, DB: cfg.RedisDB})
		if err := c.Ping(gCtx).Err(); err != nil {
			c.Close()
			return err
		}
		rdb = c
		return nil
	})

	g.Go(func() error {
		p, err := events.NewPublisher(cfg.NATSURL)
		if err != nil {
			return err
		}
		pub = p
		return nil
	})

	if err := g.Wait(); err != nil {
		if pool != nil {
			pool.Close()
		}
		if rdb != nil {
			rdb.Close()
		}
		if pub != nil {
			pub.Close()
		}
		return nil, nil, nil, err
	}
	return pool, rdb, pub, nil
}

func newPgPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}
