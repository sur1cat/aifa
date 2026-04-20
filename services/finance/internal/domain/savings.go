package domain

import (
	"time"

	"github.com/google/uuid"
)

type SavingsGoal struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	MonthlyTarget float64   `json:"monthly_target"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
