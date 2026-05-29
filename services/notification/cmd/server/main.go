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

	"github.com/sur1cat/aifa/notification-service/internal/apns"
	"github.com/sur1cat/aifa/notification-service/internal/config"
	"github.com/sur1cat/aifa/notification-service/internal/events"
	"github.com/sur1cat/aifa/notification-service/internal/handler"
	"github.com/sur1cat/aifa/notification-service/internal/jwt"
	"github.com/sur1cat/aifa/notification-service/internal/middleware"
	"github.com/sur1cat/aifa/notification-service/internal/migrate"
	"github.com/sur1cat/aifa/notification-service/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
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

	pool, rdb, nc, err := connectInfra(ctx, cfg)
	if err != nil {
		log.Error("infrastructure connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	defer rdb.Close()
	defer nc.Close()

	if err := migrate.Run(ctx, pool, "migrations", "notifications"); err != nil {
		log.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	apnsClient, err := apns.NewClient(
		cfg.APNS.KeyPath, cfg.APNS.KeyID, cfg.APNS.TeamID, cfg.APNS.BundleID, cfg.APNS.Production,
	)
	if err != nil {
		if errors.Is(err, apns.ErrNotConfigured) {
			log.Warn("APNS is not configured — push delivery disabled, /push/* endpoints still register tokens")
			apnsClient = nil
		} else {
			log.Error("APNS init failed", "err", err)
			os.Exit(1)
		}
	}

	validator := jwt.NewValidator(cfg.JWTSecret)
	blacklist := repository.NewBlacklist(rdb)
	tokenRepo := repository.NewDeviceTokenRepository(pool)

	sub, err := events.NewSubscriber(nc, tokenRepo, apnsClient)
	if err != nil {
		log.Error("events subscribe failed", "err", err)
		os.Exit(1)
	}
	defer sub.Unsubscribe()

	authMW := middleware.NewAuth(validator, blacklist)
	pushHandler := handler.NewPushHandler(tokenRepo)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestLogger(log))
	r.GET("/health", handler.Health)

	p := r.Group("", authMW.RequireAuth())
	p.POST("/push/register", pushHandler.Register)
	p.POST("/push/unregister", pushHandler.Unregister)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("notification-service listening", "port", cfg.Port)
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

func connectInfra(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, *redis.Client, *nats.Conn, error) {
	var (
		pool *pgxpool.Pool
		rdb  *redis.Client
		nc   *nats.Conn
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
		c, err := nats.Connect(cfg.NATSURL, nats.Name("notification-service"), nats.MaxReconnects(-1))
		if err != nil {
			return err
		}
		nc = c
		return nil
	})

	if err := g.Wait(); err != nil {
		if pool != nil {
			pool.Close()
		}
		if rdb != nil {
			rdb.Close()
		}
		if nc != nil {
			nc.Close()
		}
		return nil, nil, nil, err
	}
	return pool, rdb, nc, nil
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
