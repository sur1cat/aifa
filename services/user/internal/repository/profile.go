package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/sur1cat/aifa/user-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileRepository struct {
	pool *pgxpool.Pool
}

func NewProfileRepository(pool *pgxpool.Pool) *ProfileRepository {
	return &ProfileRepository{pool: pool}
}

// Upsert creates a profile on first sign-in or refreshes name/email/avatar
// if the upstream OAuth provider sent newer values. Invoked from the
// user.provisioned NATS subscriber.
func (r *ProfileRepository) Upsert(ctx context.Context, p *domain.Profile) error {
	const q = `
		INSERT INTO profiles (id, email, name, avatar_url, locale, timezone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		ON CONFLICT (id) DO UPDATE SET
			email      = COALESCE(EXCLUDED.email, profiles.email),
			name       = COALESCE(EXCLUDED.name, profiles.name),
			avatar_url = COALESCE(EXCLUDED.avatar_url, profiles.avatar_url),
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, q, p.ID, p.Email, p.Name, p.AvatarURL, p.Locale, p.Timezone)
	if err != nil {
		return fmt.Errorf("upsert profile: %w", err)
	}
	return nil
}

func (r *ProfileRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	const q = `
		SELECT id, email, name, avatar_url, locale, timezone, created_at, updated_at
		FROM profiles
		WHERE id = $1
	`
	p := &domain.Profile{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.Email, &p.Name, &p.AvatarURL, &p.Locale, &p.Timezone, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select profile: %w", err)
	}
	return p, nil
}

type UpdateInput struct {
	Name      *string
	AvatarURL *string
	Locale    *string
	Timezone  *string
}

func (r *ProfileRepository) Update(ctx context.Context, id uuid.UUID, in UpdateInput) (*domain.Profile, error) {
	const q = `
		UPDATE profiles
		SET name       = COALESCE($2, name),
		    avatar_url = COALESCE($3, avatar_url),
		    locale     = COALESCE($4, locale),
		    timezone   = COALESCE($5, timezone),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, email, name, avatar_url, locale, timezone, created_at, updated_at
	`
	p := &domain.Profile{}
	err := r.pool.QueryRow(ctx, q, id, in.Name, in.AvatarURL, in.Locale, in.Timezone).Scan(
		&p.ID, &p.Email, &p.Name, &p.AvatarURL, &p.Locale, &p.Timezone, &p.CreatedAt, &p.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update profile: %w", err)
	}
	return p, nil
}

func (r *ProfileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM profiles WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}
	return nil
}
