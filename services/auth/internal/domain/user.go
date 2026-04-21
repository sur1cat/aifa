package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuthProvider string

const (
	ProviderGoogle AuthProvider = "google"
	ProviderApple  AuthProvider = "apple"
	ProviderPhone  AuthProvider = "phone"
)

// User is the minimal identity record owned by auth-service.
// Profile data (email, name, avatar) lives in user-service and is
// propagated there via the user.provisioned event.
type User struct {
	ID           uuid.UUID    `json:"id"`
	AuthProvider AuthProvider `json:"auth_provider"`
	ProviderID   string       `json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
}
