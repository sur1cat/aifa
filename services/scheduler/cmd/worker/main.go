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

	"github.com/sur1cat/aifa/scheduler-worker/internal/config"
	"github.com/sur1cat/aifa/scheduler-worker/internal/events"

	"github.com/nats-io/nats.go"
	"github.com/robfig/cron/v3"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(log)

	cfg := config.Load()

	nc, err := nats.Connect(cfg.NATSURL, nats.Name("scheduler-worker"), nats.MaxReconnects(-1))
	if err != nil {
		log.Error("nats connect failed", "err", err)
		os.Exit(1)
	}
	defer nc.Close()
	log.Info("nats connected", "url", cfg.NATSURL)

	pub := events.NewPublisher(nc)

	c := cron.New(cron.WithLocation(time.UTC))

	if cfg.RecurringSchedule != "" {
		if _, err := c.AddFunc(cfg.RecurringSchedule, func() { pub.Tick(events.SubjectRecurringTick) }); err != nil {
			log.Error("add recurring schedule failed", "err", err, "spec", cfg.RecurringSchedule)
			os.Exit(1)
		}
		log.Info("scheduled", "subject", events.SubjectRecurringTick, "spec", cfg.RecurringSchedule)
	}
	if cfg.ReminderSchedule != "" {
		if _, err := c.AddFunc(cfg.ReminderSchedule, func() { pub.Tick(events.SubjectReminderTick) }); err != nil {
			log.Error("add reminder schedule failed", "err", err, "spec", cfg.ReminderSchedule)
			os.Exit(1)
		}
		log.Info("scheduled", "subject", events.SubjectReminderTick, "spec", cfg.ReminderSchedule)
	}

	c.Start()
	defer c.Stop()

	srv := startHealthServer(cfg.HealthAddr, log)
	defer shutdown(srv, log)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("shutting down")
}

func startHealthServer(addr string, log *slog.Logger) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"scheduler"}`))
	})
	srv := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	go func() {
		log.Info("scheduler health listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("health server failed", "err", err)
		}
	}()
	return srv
}

func shutdown(srv *http.Server, log *slog.Logger) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("health shutdown failed", "err", err)
	}
}
