package domain

import (
	"time"

	"github.com/google/uuid"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

// Kind classifies a task's relationship to money:
//
//	todo   — generic action item; amount/category may be set for context
//	bill   — money the user OWES (rent, electricity, subscription)
//	income — money the user is OWED (invoice, payday)
type Kind string

const (
	KindTodo   Kind = "todo"
	KindBill   Kind = "bill"
	KindIncome Kind = "income"
)

func (k Kind) Valid() bool {
	switch k {
	case KindTodo, KindBill, KindIncome:
		return true
	}
	return false
}

type Task struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Title       string    `json:"title"`
	IsCompleted bool      `json:"is_completed"`
	Priority    Priority  `json:"priority"`
	DueDate     string    `json:"due_date"`
	Kind        Kind      `json:"kind"`
	Amount      *float64  `json:"amount,omitempty"`
	Currency    string    `json:"currency"`
	Category    *string   `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
