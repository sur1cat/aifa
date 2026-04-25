package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/sur1cat/aifa/user-service/internal/domain"
	"github.com/sur1cat/aifa/user-service/internal/repository"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

const (
	subjectUserProvisioned = "user.provisioned"
	subjectUserDeleted     = "user.deleted"
	handlerTimeout         = 5 * time.Second
)

type Subscriber struct {
	profiles *repository.ProfileRepository
	subs     []*nats.Subscription
}

func NewSubscriber(nc *nats.Conn, profiles *repository.ProfileRepository) (*Subscriber, error) {
	s := &Subscriber{profiles: profiles}

	prov, err := nc.Subscribe(subjectUserProvisioned, s.onProvisioned)
	if err != nil {
		return nil, err
	}
	s.subs = append(s.subs, prov)

	del, err := nc.Subscribe(subjectUserDeleted, s.onDeleted)
	if err != nil {
		prov.Unsubscribe()
		return nil, err
	}
	s.subs = append(s.subs, del)

	return s, nil
}

func (s *Subscriber) Unsubscribe() {
	for _, sub := range s.subs {
		_ = sub.Unsubscribe()
	}
}

type provisionedPayload struct {
	UserID    string `json:"user_id"`
	Provider  string `json:"provider"`
	Email     string `json:"email,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type deletedPayload struct {
	UserID string `json:"user_id"`
}

func (s *Subscriber) onProvisioned(msg *nats.Msg) {
	var p provisionedPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		slog.Error("decode user.provisioned", "err", err)
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		slog.Error("parse user id", "err", err, "raw", p.UserID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	err = s.profiles.Upsert(ctx, &domain.Profile{
		ID:        uid,
		Email:     strPtr(p.Email),
		Name:      strPtr(p.Name),
		AvatarURL: strPtr(p.AvatarURL),
		Locale:    "en",
		Timezone:  "UTC",
	})
	if err != nil {
		slog.Error("upsert profile from event", "err", err, "user_id", uid)
		return
	}
	slog.Info("profile provisioned", "user_id", uid, "provider", p.Provider)
}

func (s *Subscriber) onDeleted(msg *nats.Msg) {
	var p deletedPayload
	if err := json.Unmarshal(msg.Data, &p); err != nil {
		slog.Error("decode user.deleted", "err", err)
		return
	}
	uid, err := uuid.Parse(p.UserID)
	if err != nil {
		slog.Error("parse user id", "err", err, "raw", p.UserID)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), handlerTimeout)
	defer cancel()

	if err := s.profiles.Delete(ctx, uid); err != nil {
		slog.Error("delete profile from event", "err", err, "user_id", uid)
		return
	}
	slog.Info("profile deleted", "user_id", uid)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
