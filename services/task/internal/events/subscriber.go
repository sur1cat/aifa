package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/sur1cat/aifa/task-service/internal/domain"
	"github.com/sur1cat/aifa/task-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	SubjectUserDeleted    = "user.deleted"
	subjectBudgetExceeded = "budget.exceeded"
	handlerTimeout        = 5 * time.Second
)

type Subscriber struct {
	tasks *repository.TaskRepository
	subs  []*nats.Subscription
}

func NewSubscriber(nc *nats.Conn, tasks *repository.TaskRepository) (*Subscriber, error) {
	s := &Subscriber{tasks: tasks}

	for _, binding := range []struct {
		subject string
		handler nats.MsgHandler
	}{
		{SubjectUserDeleted, s.onUserDeleted},
		{subjectBudgetExceeded, s.onBudgetExceeded},
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

type budgetExceededPayload struct {
	UserID       string  `json:"user_id"`
	Category     string  `json:"category"`
	LabelRu      string  `json:"label_ru"`
	MonthlyLimit float64 `json:"monthly_limit"`
	Spent        float64 `json:"spent"`
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

	if err := s.tasks.DeleteByUser(ctx, uid); err != nil {
		slog.Error("delete tasks by user", "err", err, "user_id", uid)
		return
	}
	slog.Info("tasks purged", "user_id", uid)
}

// onBudgetExceeded auto-creates a high-priority task so the user sees a
// concrete action item when their spending goes over the monthly limit.
func (s *Subscriber) onBudgetExceeded(msg *nats.Msg) {
	var p budgetExceededPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		slog.Error("decode budget.exceeded", "err", err)
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		slog.Error("parse user id in budget.exceeded", "err", err)
		return
	}

	label := p.LabelRu
	if label == "" {
		label = p.Category
	}
	title := fmt.Sprintf("Проверь расходы: %s (%.0f₸ из %.0f₸)", label, p.Spent, p.MonthlyLimit)
	dueDate := time.Now().Format("2006-01-02")
	category := p.Category

	t := &domain.Task{
		UserID:   uid,
		Title:    title,
		Priority: domain.PriorityHigh,
		DueDate:  dueDate,
		Kind:     domain.KindTodo,
		Category: &category,
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	if err := s.tasks.Create(ctx, t); err != nil {
		slog.Error("create budget task", "err", err, "user_id", uid, "category", p.Category)
		return
	}
	slog.Info("budget task created", "user_id", uid, "category", p.Category, "task_id", t.ID)
}
