package domain

import (
	"time"

	"github.com/google/uuid"
)

type Period string

const (
	PeriodDaily   Period = "daily"
	PeriodWeekly  Period = "weekly"
	PeriodMonthly Period = "monthly"
)

type Habit struct {
	ID             uuid.UUID      `json:"id"`
	UserID         uuid.UUID      `json:"user_id"`
	GoalID         *uuid.UUID     `json:"goal_id,omitempty"`
	Title          string         `json:"title"`
	Icon           string         `json:"icon"`
	Color          string         `json:"color"`
	Period         Period         `json:"period"`
	TargetValue    *int           `json:"target_value,omitempty"`
	Unit           *string        `json:"unit,omitempty"`
	CompletedDates []string       `json:"completed_dates"`
	ProgressValues map[string]int `json:"progress_values"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	ArchivedAt     *time.Time     `json:"archived_at,omitempty"`
}
