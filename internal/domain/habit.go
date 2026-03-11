package domain

import (
	"time"

	"github.com/google/uuid"
)

type HabitPeriod string

const (
	HabitPeriodDaily   HabitPeriod = "daily"
	HabitPeriodWeekly  HabitPeriod = "weekly"
	HabitPeriodMonthly HabitPeriod = "monthly"
)

type Habit struct {
	ID             uuid.UUID      `json:"id"`
	UserID         uuid.UUID      `json:"user_id"`
	GoalID         *uuid.UUID     `json:"goal_id"`
	Title          string         `json:"title"`
	Icon           string         `json:"icon"`
	Color          string         `json:"color"`
	Period         HabitPeriod    `json:"period"`
	CompletedDates []string       `json:"completed_dates"`
	TargetValue    *int           `json:"target_value"`
	Unit           *string        `json:"unit"`
	ProgressValues map[string]int `json:"progress_values"`
	CreatedAt      time.Time      `json:"created_at"`
	ArchivedAt     *time.Time     `json:"archived_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type HabitCompletion struct {
	ID            uuid.UUID `json:"id"`
	HabitID       uuid.UUID `json:"habit_id"`
	CompletedDate string    `json:"completed_date"`
	CreatedAt     time.Time `json:"created_at"`
}
