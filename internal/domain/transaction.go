package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeIncome  TransactionType = "income"
	TransactionTypeExpense TransactionType = "expense"
)

type Transaction struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	Title     string          `json:"title"`
	Amount    float64         `json:"amount"`
	Type      TransactionType `json:"type"`
	Category  string          `json:"category"`
	Date      string          `json:"date"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}
