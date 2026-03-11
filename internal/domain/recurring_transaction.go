package domain

import (
	"time"

	"github.com/google/uuid"
)

type RecurrenceFrequency string

const (
	FrequencyWeekly    RecurrenceFrequency = "weekly"
	FrequencyBiweekly  RecurrenceFrequency = "biweekly"
	FrequencyMonthly   RecurrenceFrequency = "monthly"
	FrequencyQuarterly RecurrenceFrequency = "quarterly"
	FrequencyYearly    RecurrenceFrequency = "yearly"
)

type RecurringTransaction struct {
	ID                uuid.UUID           `json:"id"`
	UserID            uuid.UUID           `json:"user_id"`
	Title             string              `json:"title"`
	Amount            float64             `json:"amount"`
	Type              TransactionType     `json:"type"`
	Category          string              `json:"category"`
	Frequency         RecurrenceFrequency `json:"frequency"`
	StartDate         string              `json:"start_date"`
	NextDate          string              `json:"next_date"`
	EndDate           *string             `json:"end_date"`
	RemainingPayments *int                `json:"remaining_payments"`
	IsActive          bool                `json:"is_active"`
	CreatedAt         time.Time           `json:"created_at"`
	UpdatedAt         time.Time           `json:"updated_at"`
}

func (r *RecurringTransaction) CalculateNextDate() string {
	layout := "2006-01-02"
	current, err := time.Parse(layout, r.NextDate)
	if err != nil {
		return r.NextDate
	}

	var next time.Time
	switch r.Frequency {
	case FrequencyWeekly:
		next = current.AddDate(0, 0, 7)
	case FrequencyBiweekly:
		next = current.AddDate(0, 0, 14)
	case FrequencyMonthly:
		next = current.AddDate(0, 1, 0)
	case FrequencyQuarterly:
		next = current.AddDate(0, 3, 0)
	case FrequencyYearly:
		next = current.AddDate(1, 0, 0)
	default:
		next = current.AddDate(0, 1, 0)
	}

	return next.Format(layout)
}
