package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sur1cat/aifa/notification-service/internal/apns"
	"github.com/sur1cat/aifa/notification-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	SubjectUserDeleted = "user.deleted"
	SubjectReminderDue = "reminder.due"
	handlerTimeout     = 10 * time.Second
)

// Subscriber wires NATS subjects to handler methods. Push delivery is
// best-effort: per-token failures are logged but don't propagate.
type Subscriber struct {
	tokens *repository.DeviceTokenRepository
	apns   *apns.Client
	subs   []*nats.Subscription
}

func NewSubscriber(nc *nats.Conn, tokens *repository.DeviceTokenRepository, client *apns.Client) (*Subscriber, error) {
	s := &Subscriber{tokens: tokens, apns: client}

	for _, b := range []struct {
		subject string
		handler nats.MsgHandler
	}{
		{SubjectUserDeleted, s.onUserDeleted},
		{SubjectReminderDue, s.onReminderDue},
	} {
		sub, err := nc.Subscribe(b.subject, b.handler)
		if err != nil {
			s.Unsubscribe()
			return nil, err
		}
		s.subs = append(s.subs, sub)
	}
	return s, nil
}

func (s *Subscriber) Unsubscribe() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
}

type userDeletedPayload struct {
	UserID string `json:"user_id"`
}

func (s *Subscriber) onUserDeleted(msg *nats.Msg) {
	var p userDeletedPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		slog.Error("decode user.deleted", "err", err)
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		slog.Error("parse user id", "err", err, "raw", p.UserID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	if err := s.tokens.DeleteByUser(ctx, uid); err != nil {
		slog.Error("delete tokens by user", "err", err, "user_id", uid)
		return
	}
	slog.Info("device tokens purged", "user_id", uid)
}

// reminderDuePayload is what scheduler-worker publishes when something is due
// (e.g. recurring transaction landed, habit window starting). Title and body
// are pre-localized by the publisher — notification-service only ships them.
type reminderDuePayload struct {
	UserID string         `json:"user_id"`
	Title  string         `json:"title"`
	Body   string         `json:"body"`
	Data   map[string]any `json:"data,omitempty"`
}

func (s *Subscriber) onReminderDue(msg *nats.Msg) {
	if s.apns == nil {
		// APNS not configured — drop silently in dev. Scheduler still ticks.
		return
	}
	var p reminderDuePayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		slog.Error("decode reminder.due", "err", err)
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		slog.Error("parse user id", "err", err, "raw", p.UserID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	tokens, err := s.tokens.ListByUser(ctx, uid)
	if err != nil {
		slog.Error("list tokens for reminder", "err", err, "user_id", uid)
		return
	}
	for _, t := range tokens {
		if err := s.apns.Send(ctx, &apns.Notification{
			DeviceToken: t.Token,
			Title:       p.Title,
			Body:        p.Body,
			Data:        p.Data,
		}); err != nil {
			slog.Warn("apns send failed", "err", err, "user_id", uid, "platform", t.Platform)
			continue
		}
	}
	slog.Info("reminder delivered", "user_id", uid, "tokens", len(tokens))
}
