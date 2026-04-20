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
	TargetValue *int       `json:"target_value,omitempty"`
	Unit        *string    `json:"unit,omitempty"`
	Deadline    *time.Time `json:"deadline,omitempty"`
	ArchivedAt  *time.Time `json:"archived_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
