package domain

import (
	"time"

	"github.com/google/uuid"
)

// SavingsGoal represents a user's monthly savings target
type SavingsGoal struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	MonthlyTarget float64   `json:"monthly_target"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SavingsGoalWithProgress includes calculated savings progress
type SavingsGoalWithProgress struct {
	ID             uuid.UUID `json:"id"`
	MonthlyTarget  float64   `json:"monthlyTarget"`
	CurrentSavings float64   `json:"currentSavings"`
	MonthlyIncome  float64   `json:"monthlyIncome"`
	MonthlyExpenses float64  `json:"monthlyExpenses"`
	Progress       float64   `json:"progress"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
