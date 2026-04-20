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
	GoalID         *uuid.UUID     `json:"goal_id"`          // Optional goal reference
	Title          string         `json:"title"`
	Icon           string         `json:"icon"`
	Color          string         `json:"color"`
	Period         HabitPeriod    `json:"period"`
	CompletedDates []string       `json:"completed_dates"`  // List of dates in "YYYY-MM-DD" format
	TargetValue    *int           `json:"target_value"`     // Goal value (e.g., 100 pushups)
	Unit           *string        `json:"unit"`             // Unit of measurement (e.g., "reps", "pages")
	ProgressValues map[string]int `json:"progress_values"`  // Progress per date {"2024-01-15": 50}
	CreatedAt      time.Time      `json:"created_at"`
	ArchivedAt     *time.Time     `json:"archived_at"`      // When habit was archived (nil = active)
	UpdatedAt      time.Time      `json:"updated_at"`
}

type HabitCompletion struct {
	ID            uuid.UUID `json:"id"`
	HabitID       uuid.UUID `json:"habit_id"`
	CompletedDate string    `json:"completed_date"` // "YYYY-MM-DD" format
	CreatedAt     time.Time `json:"created_at"`
}
