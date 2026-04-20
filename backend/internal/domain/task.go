package domain

import (
	"time"

	"github.com/google/uuid"
)

type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

type Task struct {
	ID          uuid.UUID    `json:"id"`
	UserID      uuid.UUID    `json:"user_id"`
	Title       string       `json:"title"`
	IsCompleted bool         `json:"is_completed"`
	Priority    TaskPriority `json:"priority"`
	DueDate     string       `json:"due_date"` // "YYYY-MM-DD" format
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
