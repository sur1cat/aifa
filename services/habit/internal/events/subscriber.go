package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/sur1cat/aifa/habit-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	subjectUserDeleted  = "user.deleted"
	subjectGoalDeleted  = "goal.deleted"
	subjectReminderTick = "cron.reminder.tick"
	subjectReminderDue  = "reminder.due"
	handlerTimeout      = 5 * time.Second
)

type Subscriber struct {
	habits *repository.HabitRepository
	nc     *nats.Conn
	subs   []*nats.Subscription
}

func NewSubscriber(nc *nats.Conn, habits *repository.HabitRepository) (*Subscriber, error) {
	s := &Subscriber{habits: habits, nc: nc}

	for _, binding := range []struct {
		subject string
		handler nats.MsgHandler
	}{
		{subjectUserDeleted, s.onUserDeleted},
		{subjectGoalDeleted, s.onGoalDeleted},
		{subjectReminderTick, s.onReminderTick},
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

type reminderDuePayload struct {
	UserID string         `json:"user_id"`
	Title  string         `json:"title"`
	Body   string         `json:"body"`
	Data   map[string]any `json:"data,omitempty"`
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

// onReminderTick handles cron.reminder.tick: fans out a reminder.due event
// for every user who has at least one active (non-archived) habit today.
func (s *Subscriber) onReminderTick(_ *nats.Msg) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	userIDs, err := s.habits.ListActiveUserIDs(ctx)
	if err != nil {
		slog.Error("reminder tick: list active users", "err", err)
		return
	}

	today := time.Now().Format("2006-01-02")
	sent := 0
	for _, uid := range userIDs {
		habits, err := s.habits.ListByUser(ctx, uid)
		if err != nil {
			slog.Warn("reminder tick: list habits", "err", err, "user_id", uid)
			continue
		}

		// Only send for habits that are active today and not yet completed.
		var pending []string
		for _, h := range habits {
			if h.ArchivedAt != nil {
				continue
			}
			completed := false
			for _, d := range h.CompletedDates {
				if d == today {
					completed = true
					break
				}
			}
			if !completed {
				pending = append(pending, h.Title)
			}
		}
		if len(pending) == 0 {
			continue
		}

		body := fmt.Sprintf("%s", strings.Join(pending, ", "))
		if len(pending) > 3 {
			body = fmt.Sprintf("%s и ещё %d", strings.Join(pending[:3], ", "), len(pending)-3)
		}

		payload := reminderDuePayload{
			UserID: uid.String(),
			Title:  "Привычки на сегодня",
			Body:   body,
			Data:   map[string]any{"type": "habit_reminder", "date": today},
		}
		data, _ := json.Marshal(payload)
		if err := s.nc.Publish(subjectReminderDue, data); err != nil {
			slog.Warn("publish reminder.due", "err", err, "user_id", uid)
			continue
		}
		sent++
	}
	slog.Info("habit reminders dispatched", "users_notified", sent, "total_users", len(userIDs))
}
