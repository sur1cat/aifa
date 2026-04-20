package domain

import (
	"time"

	"github.com/google/uuid"
)

type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderApple  AuthProvider = "apple"
)

type User struct {
	ID           uuid.UUID    `json:"id"`
	Email        *string      `json:"email,omitempty"`
	Phone        *string      `json:"phone,omitempty"`
	Name         *string      `json:"name,omitempty"`
	AvatarURL    *string      `json:"avatar_url,omitempty"`
	AuthProvider AuthProvider `json:"auth_provider"`
	ProviderID   string       `json:"-"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}
