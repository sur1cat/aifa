package domain

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	ID        uuid.UUID `json:"id"`
	Email     *string   `json:"email,omitempty"`
	Name      *string   `json:"name,omitempty"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	Locale    string    `json:"locale"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
