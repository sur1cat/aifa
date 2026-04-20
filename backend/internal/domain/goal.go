package domain

import (
	"time"

	"github.com/google/uuid"
)

type Goal struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Title       string     `json:"title"`
	Icon        string     `json:"icon"`
	TargetValue *int       `json:"target_value"` // Optional target (e.g., "Complete 5 habits")
	Unit        *string    `json:"unit"`         // Optional unit (e.g., "habits", "days")
	Deadline    *time.Time `json:"deadline"`     // Optional deadline
	CreatedAt   time.Time  `json:"created_at"`
	ArchivedAt  *time.Time `json:"archived_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
