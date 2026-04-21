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

// Kind classifies a habit's relationship to the user's money:
//
//	tracking  — improves financial awareness; no direct money impact
//	saving    — completion is a deposit toward savings (expected_amount > 0)
//	spending  — completion means the user kept under a budget for an
//	            expense category (expected_amount = the cap)
type Kind string

const (
	KindTracking Kind = "tracking"
	KindSaving   Kind = "saving"
	KindSpending Kind = "spending"
)

func (k Kind) Valid() bool {
	switch k {
	case KindTracking, KindSaving, KindSpending:
		return true
	}
	return false
}

type Habit struct {
	ID                uuid.UUID      `json:"id"`
	UserID            uuid.UUID      `json:"user_id"`
	GoalID            *uuid.UUID     `json:"goal_id,omitempty"`
	Title             string         `json:"title"`
	Icon              string         `json:"icon"`
	Color             string         `json:"color"`
	Period            Period         `json:"period"`
	Kind              Kind           `json:"kind"`
	Currency          string         `json:"currency"`
	FinancialCategory *string        `json:"financial_category,omitempty"`
	ExpectedAmount    *float64       `json:"expected_amount,omitempty"`
	TargetValue       *int           `json:"target_value,omitempty"`
	Unit              *string        `json:"unit,omitempty"`
	CompletedDates    []string       `json:"completed_dates"`
	ProgressValues    map[string]int `json:"progress_values"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	ArchivedAt        *time.Time     `json:"archived_at,omitempty"`
}
