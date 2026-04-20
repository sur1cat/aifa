package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sur1cat/aifa/habit-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	subjectUserDeleted = "user.deleted"
	subjectGoalDeleted = "goal.deleted"
	handlerTimeout     = 5 * time.Second
)

type Subscriber struct {
	habits *repository.HabitRepository
	subs   []*nats.Subscription
}

func NewSubscriber(nc *nats.Conn, habits *repository.HabitRepository) (*Subscriber, error) {
	s := &Subscriber{habits: habits}

	for _, binding := range []struct {
		subject string
		handler nats.MsgHandler
	}{
		{subjectUserDeleted, s.onUserDeleted},
		{subjectGoalDeleted, s.onGoalDeleted},
	} {
		sub, err := nc.Subscribe(binding.subject, binding.handler)
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

type goalDeletedPayload struct {
	GoalID string `json:"goal_id"`
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

	if err := s.habits.DeleteByUser(ctx, uid); err != nil {
		slog.Error("delete habits by user", "err", err, "user_id", uid)
		return
	}
	slog.Info("habits purged", "user_id", uid)
}

func (s *Subscriber) onGoalDeleted(msg *nats.Msg) {
	var p goalDeletedPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		slog.Error("decode goal.deleted", "err", err)
		return
	}
	gid, err := uuid.Parse(p.GoalID)
	if err != nil {
		slog.Error("parse goal id", "err", err, "raw", p.GoalID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	if err := s.habits.ClearGoalRef(ctx, gid); err != nil {
		slog.Error("clear goal ref", "err", err, "goal_id", gid)
		return
	}
	slog.Info("habits detached from goal", "goal_id", gid)
}
