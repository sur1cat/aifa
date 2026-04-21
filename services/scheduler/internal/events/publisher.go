package events

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	// SubjectRecurringTick wakes finance-service to scan and advance every
	// user's due recurring rows. Replaces the iOS-client-triggered
	// /recurring-transactions/process endpoint as the canonical source.
	SubjectRecurringTick = "cron.recurring.tick"

	// SubjectReminderTick is a daily heartbeat for services that want to
	// fan out per-user reminder.due events. Today nobody subscribes —
	// the subject is published so consumers can opt in later without a
	// scheduler change.
	SubjectReminderTick = "cron.reminder.tick"
)

type TickPayload struct {
	FiredAt time.Time `json:"fired_at"`
	Source  string    `json:"source"`
}

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) Tick(subject string) {
	data, err := json.Marshal(TickPayload{FiredAt: time.Now().UTC(), Source: "scheduler-worker"})
	if err != nil {
		slog.Error("marshal tick", "subject", subject, "err", err)
		return
	}
	if err := p.nc.Publish(subject, data); err != nil {
		slog.Error("publish tick", "subject", subject, "err", err)
		return
	}
	slog.Info("tick published", "subject", subject)
}
