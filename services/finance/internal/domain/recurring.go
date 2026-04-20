package domain

import (
	"time"

	"github.com/google/uuid"
)

type Frequency string

const (
	FreqWeekly    Frequency = "weekly"
	FreqBiweekly  Frequency = "biweekly"
	FreqMonthly   Frequency = "monthly"
	FreqQuarterly Frequency = "quarterly"
	FreqYearly    Frequency = "yearly"
)

type Recurring struct {
	ID                uuid.UUID       `json:"id"`
	UserID            uuid.UUID       `json:"user_id"`
	Title             string          `json:"title"`
	Amount            float64         `json:"amount"`
	Type              TransactionType `json:"type"`
	Category          string          `json:"category"`
	Frequency         Frequency       `json:"frequency"`
	StartDate         string          `json:"start_date"`
	NextDate          string          `json:"next_date"`
	EndDate           *string         `json:"end_date,omitempty"`
	RemainingPayments *int            `json:"remaining_payments,omitempty"`
	IsActive          bool            `json:"is_active"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// NextDateFrom returns the next occurrence date in YYYY-MM-DD layout after
// `current`, applying the recurrence frequency. Falls back to +1 month on
// unknown frequency.
func (r *Recurring) NextDateFrom(current string) string {
	const layout = "2006-01-02"
	t, err := time.Parse(layout, current)
	if err != nil {
		return current
	}
	switch r.Frequency {
	case FreqWeekly:
		return t.AddDate(0, 0, 7).Format(layout)
	case FreqBiweekly:
		return t.AddDate(0, 0, 14).Format(layout)
	case FreqQuarterly:
		return t.AddDate(0, 3, 0).Format(layout)
	case FreqYearly:
		return t.AddDate(1, 0, 0).Format(layout)
	default:
		return t.AddDate(0, 1, 0).Format(layout)
	}
}
