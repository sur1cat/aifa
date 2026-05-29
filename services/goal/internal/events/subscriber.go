package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sur1cat/aifa/goal-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	SubjectUserDeleted = "user.deleted"
	handlerTimeout     = 5 * time.Second
)

type Subscriber struct {
	goals *repository.GoalRepository
	sub   *nats.Subscription
}

func NewSubscriber(nc *nats.Conn, goals *repository.GoalRepository) (*Subscriber, error) {
	s := &Subscriber{goals: goals}
	sub, err := nc.Subscribe(SubjectUserDeleted, s.onUserDeleted)
	if err != nil {
		return nil, err
	}
	s.sub = sub
	return s, nil
}

func (s *Subscriber) Unsubscribe() {
	if s.sub != nil {
		_ = s.sub.Unsubscribe()
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

	if err := s.goals.DeleteByUser(ctx, uid); err != nil {
		slog.Error("delete goals by user", "err", err, "user_id", uid)
		return
	}
	slog.Info("goals purged", "user_id", uid)
}
