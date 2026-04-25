package events

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

const SubjectBudgetExceeded = "budget.exceeded"

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(nc *nats.Conn) *Publisher {
	return &Publisher{nc: nc}
}

type BudgetExceededPayload struct {
	UserID       string  `json:"user_id"`
	Category     string  `json:"category"`
	LabelRu      string  `json:"label_ru"`
	MonthlyLimit float64 `json:"monthly_limit"`
	Spent        float64 `json:"spent"`
}

func (p *Publisher) Publish(subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", subject, err)
	}
	return p.nc.Publish(subject, data)
}

func (p *Publisher) PublishBudgetExceeded(payload BudgetExceededPayload) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal budget.exceeded: %w", err)
	}
	return p.nc.Publish(SubjectBudgetExceeded, data)
}
