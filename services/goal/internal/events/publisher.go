package events

import (
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
)

const SubjectGoalDeleted = "goal.deleted"

type GoalDeleted struct {
	GoalID string `json:"goal_id"`
	UserID string `json:"user_id"`
}

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

func (p *Publisher) PublishGoalDeleted(evt GoalDeleted) {
	data, err := json.Marshal(evt)
	if err != nil {
		slog.Error("marshal goal.deleted", "err", err)
		return
	}
	if err := p.nc.Publish(SubjectGoalDeleted, data); err != nil {
		slog.Error("publish goal.deleted", "err", err)
	}
}
