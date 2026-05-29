package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sur1cat/aifa/finance-service/internal/domain"
	"github.com/sur1cat/aifa/finance-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	SubjectUserDeleted    = "user.deleted"
	SubjectTransactionNew = "transaction.created"
	handlerTimeout        = 5 * time.Second
)

type Subscriber struct {
	transactions *repository.TransactionRepository
	recurring    *repository.RecurringRepository
	savings      *repository.SavingsRepository
	rules        *repository.SavingsRuleRepository
	pub          *Publisher
	subs         []*nats.Subscription
}

func NewSubscriber(
	nc *nats.Conn,
	transactions *repository.TransactionRepository,
	recurring *repository.RecurringRepository,
	savings *repository.SavingsRepository,
	rules *repository.SavingsRuleRepository,
	pub *Publisher,
) (*Subscriber, error) {
	s := &Subscriber{
		transactions: transactions,
		recurring:    recurring,
		savings:      savings,
		rules:        rules,
		pub:          pub,
	}

	pairs := []struct {
		subj    string
		handler nats.MsgHandler
	}{
		{SubjectUserDeleted, s.onUserDeleted},
		{SubjectTransactionNew, s.onTransactionCreated},
	}
	for _, p := range pairs {
		sub, err := nc.Subscribe(p.subj, p.handler)
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

// ── user.deleted ──────────────────────────────────────────────────────────────

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

// ── transaction.created ───────────────────────────────────────────────────────

type transactionCreatedPayload struct {
	UserID   string  `json:"user_id"`
	TxID     string  `json:"tx_id"`
	Type     string  `json:"type"`
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
}

func (s *Subscriber) onTransactionCreated(msg *nats.Msg) {
	var p transactionCreatedPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		slog.Error("decode transaction.created", "err", err)
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	// Правило: on_income_savings — откладываем X с каждого дохода
	if p.Type == "income" {
		rules, err := s.rules.ListActiveByKind(ctx, uid, domain.KindOnIncomeSavings)
		if err != nil {
			slog.Error("list on_income rules", "err", err, "user_id", uid)
		}
		for _, rule := range rules {
			savingsTx := &domain.Transaction{
				UserID:   uid,
				Title:    "Автонакопление с дохода",
				Amount:   rule.Amount,
				Type:     domain.TypeExpense,
				Category: "savings",
				Date:     time.Now().Format("2006-01-02"),
			}
			if err := s.transactions.Create(ctx, savingsTx); err != nil {
				slog.Error("create auto savings tx", "err", err, "rule_id", rule.ID)
				continue
			}
			slog.Info("auto savings applied", "user_id", uid, "amount", rule.Amount)
		}
	}

	// Правило: spending_alert — дневной лимит расходов
	if p.Type == "expense" {
		alerts, err := s.rules.ListActiveByKind(ctx, uid, domain.KindSpendingAlert)
		if err != nil {
			slog.Error("list spending_alert rules", "err", err)
			return
		}
		if len(alerts) == 0 {
			return
		}
		dailySpent, err := s.rules.DailySpent(ctx, uid)
		if err != nil {
			slog.Error("daily spent check", "err", err)
			return
		}
		for _, alert := range alerts {
			if dailySpent >= alert.Amount {
				slog.Info("spending alert triggered",
					"user_id", uid, "daily_spent", dailySpent, "limit", alert.Amount)
				// Публикуем событие — notification-service отправит push
				_ = s.pub.Publish("spending.alert", map[string]any{
					"user_id":     uid.String(),
					"daily_spent": dailySpent,
					"limit":       alert.Amount,
				})
			}
		}
	}
}
