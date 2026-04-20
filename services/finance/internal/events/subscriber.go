package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sur1cat/aifa/finance-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	SubjectUserDeleted = "user.deleted"
	handlerTimeout     = 5 * time.Second
)

type Subscriber struct {
	transactions *repository.TransactionRepository
	recurring    *repository.RecurringRepository
	savings      *repository.SavingsRepository
	sub          *nats.Subscription
}

func NewSubscriber(
	nc *nats.Conn,
	transactions *repository.TransactionRepository,
	recurring *repository.RecurringRepository,
	savings *repository.SavingsRepository,
) (*Subscriber, error) {
	s := &Subscriber{transactions: transactions, recurring: recurring, savings: savings}
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

	if err := s.transactions.DeleteByUser(ctx, uid); err != nil {
		slog.Error("delete transactions by user", "err", err, "user_id", uid)
	}
	if err := s.recurring.DeleteByUser(ctx, uid); err != nil {
		slog.Error("delete recurring by user", "err", err, "user_id", uid)
	}
	if err := s.savings.DeleteByUser(ctx, uid); err != nil {
		slog.Error("delete savings by user", "err", err, "user_id", uid)
	}
	slog.Info("finance data purged", "user_id", uid)
}
