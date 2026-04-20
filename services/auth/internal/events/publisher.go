package events

import (
	"encoding/json"
	"log/slog"

	"github.com/nats-io/nats.go"
)

const (
	SubjectUserProvisioned = "user.provisioned"
	SubjectUserDeleted     = "user.deleted"
)

type UserProvisioned struct {
	UserID    string `json:"user_id"`
	Provider  string `json:"provider"`
	Email     string `json:"email,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type UserDeleted struct {
	UserID string `json:"user_id"`
}

type Publisher struct {
	nc *nats.Conn
}

func NewPublisher(url string) (*Publisher, error) {
	nc, err := nats.Connect(url, nats.Name("auth-service"), nats.MaxReconnects(-1))
	if err != nil {
		return nil, err
	}
	return &Publisher{nc: nc}, nil
}

func (p *Publisher) Close() {
	if p.nc != nil {
		p.nc.Close()
	}
}

func (p *Publisher) PublishUserProvisioned(evt UserProvisioned) {
	p.publish(SubjectUserProvisioned, evt)
}

func (p *Publisher) PublishUserDeleted(evt UserDeleted) {
	p.publish(SubjectUserDeleted, evt)
}

func (p *Publisher) publish(subject string, payload any) {
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("marshal event", "subject", subject, "err", err)
		return
	}
	if err := p.nc.Publish(subject, data); err != nil {
		slog.Error("publish event", "subject", subject, "err", err)
	}
}
