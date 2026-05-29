package events

import (
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
)

const (
	SubjectGoalDeleted  = "goal.deleted"
	SubjectGoalProgress = "goal.progress"
)

type GoalDeleted struct {
	GoalID string `json:"goal_id"`
	UserID string `json:"user_id"`
}

type GoalProgress struct {
	GoalID        string  `json:"goal_id"`
	UserID        string  `json:"user_id"`
	CurrentAmount float64 `json:"current_amount"`
	TargetAmount  float64 `json:"target_amount,omitempty"`
	Completed     bool    `json:"completed"`
}

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) PublishGoalDeleted(evt GoalDeleted) {
	p.publish(SubjectGoalDeleted, evt)
}

func (p *Publisher) PublishGoalProgress(evt GoalProgress) {
	p.publish(SubjectGoalProgress, evt)
}

func (p *Publisher) publish(subject string, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		slog.Error("marshal event", "subject", subject, "err", err)
		return
	}
	if err := p.nc.Publish(subject, data); err != nil {
		slog.Error("publish event", "subject", subject, "err", err)
	}
}
