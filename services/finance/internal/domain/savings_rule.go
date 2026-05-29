package domain

import (
	"time"

	"github.com/google/uuid"
)

type SavingsRuleKind string

const (
	KindMonthlySavings  SavingsRuleKind = "monthly_savings"
	KindOnIncomeSavings SavingsRuleKind = "on_income_savings"
	KindSpendingAlert   SavingsRuleKind = "spending_alert"
)

type SavingsRule struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Kind      SavingsRuleKind `json:"kind"`
	Amount    float64         `json:"amount"`
	Period    *string         `json:"period,omitempty"`
	GoalTitle *string         `json:"goal_title,omitempty"`
	Active    bool            `json:"active"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
