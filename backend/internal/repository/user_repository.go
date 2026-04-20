package repository

import (
	"context"
	"errors"
	"fmt"

	"habitflow/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) FindOrCreateByProvider(ctx context.Context, provider domain.AuthProvider, providerID string, email, name, avatarURL *string) (*domain.User, bool, error) {
	// Try to find existing user
	user, err := r.GetByProvider(ctx, provider, providerID)
	if err == nil {
		// Update user info if changed
		if name != nil || avatarURL != nil {
			user.Name = name
			user.AvatarURL = avatarURL
			_ = r.Update(ctx, user)
		}
		return user, false, nil
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, false, err
	}

	// Create new user
	user = &domain.User{
		ID:           uuid.New(),
		Email:        email,
		Name:         name,
		AvatarURL:    avatarURL,
		AuthProvider: provider,
		ProviderID:   providerID,
	}

	query := `
		INSERT INTO users (id, email, name, avatar_url, auth_provider, provider_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err = r.pool.QueryRow(ctx, query, user.ID, user.Email, user.Name, user.AvatarURL, user.AuthProvider, user.ProviderID).
		Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create user: %w", err)
	}

	return user, true, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user := &domain.User{}

	query := `
		SELECT id, email, phone, name, avatar_url, auth_provider, provider_id, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.AvatarURL,
		&user.AuthProvider,
		&user.ProviderID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	user := &domain.User{}

	query := `
		SELECT id, email, phone, name, avatar_url, auth_provider, provider_id, created_at, updated_at
		FROM users
		WHERE auth_provider = $1 AND provider_id = $2
	`

	err := r.pool.QueryRow(ctx, query, provider, providerID).Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.AvatarURL,
		&user.AuthProvider,
		&user.ProviderID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by provider: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, phone = $3, name = $4, avatar_url = $5, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query, user.ID, user.Email, user.Phone, user.Name, user.AvatarURL)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
