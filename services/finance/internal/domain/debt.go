package domain

import (
	"time"

	"github.com/google/uuid"
)

type DebtDirection string

const (
	DirectionIOwe    DebtDirection = "i_owe"
	DirectionTheyOwe DebtDirection = "they_owe"
)

type Debt struct {
	ID             uuid.UUID     `json:"id"`
	UserID         uuid.UUID     `json:"user_id"`
	Counterparty   string        `json:"counterparty"`
	Direction      DebtDirection `json:"direction"`
	Amount         float64       `json:"amount"`
	OriginalAmount float64       `json:"original_amount"`
	Note           *string       `json:"note,omitempty"`
	Settled        bool          `json:"settled"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}
