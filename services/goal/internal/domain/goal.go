package domain

import (
	"time"

	"github.com/google/uuid"
)

// GoalType classifies a financial goal:
//
//	savings    — accumulate money toward a target (emergency fund, trip)
//	debt       — pay down a balance; current_amount = paid so far
//	purchase   — discrete purchase when current_amount hits target
//	investment — track contributions into investments (not market value)
type GoalType string

const (
	GoalSavings    GoalType = "savings"
	GoalDebt       GoalType = "debt"
	GoalPurchase   GoalType = "purchase"
	GoalInvestment GoalType = "investment"
)

func (g GoalType) Valid() bool {
	switch g {
	case GoalSavings, GoalDebt, GoalPurchase, GoalInvestment:
		return true
	}
	return false
}

type Goal struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	Title         string     `json:"title"`
	Icon          string     `json:"icon"`
	GoalType      GoalType   `json:"goal_type"`
	TargetAmount  *float64   `json:"target_amount,omitempty"`
	CurrentAmount float64    `json:"current_amount"`
	Currency      string     `json:"currency"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	ArchivedAt    *time.Time `json:"archived_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Progress returns 0..1 once a target is set; 0 otherwise.
func (g *Goal) Progress() float64 {
	if g.TargetAmount == nil || *g.TargetAmount <= 0 {
		return 0
	}
	p := g.CurrentAmount / *g.TargetAmount
	switch {
	case p > 1:
		return 1
	case p < 0:
		return 0
	default:
		return p
	}
}
