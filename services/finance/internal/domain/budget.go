package domain

import (
	"time"

	"github.com/google/uuid"
)

type Budget struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Category     string
	MonthlyLimit float64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
