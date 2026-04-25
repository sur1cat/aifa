package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/sur1cat/aifa/auth-service/internal/domain"

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

// FindOrCreate returns (user, isNew, err).
func (r *UserRepository) FindOrCreate(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, bool, error) {
	if u, err := r.GetByProvider(ctx, provider, providerID); err == nil {
		return u, false, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return nil, false, err
	}

	u := &domain.User{ID: uuid.New(), AuthProvider: provider, ProviderID: providerID}
	const q = `
		INSERT INTO users (id, auth_provider, provider_id, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING created_at
	`
	if err := r.pool.QueryRow(ctx, q, u.ID, u.AuthProvider, u.ProviderID).Scan(&u.CreatedAt); err != nil {
		return nil, false, fmt.Errorf("insert user: %w", err)
	}
	return u, true, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const q = `SELECT id, auth_provider, provider_id, created_at FROM users WHERE id = $1`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, q, id).Scan(&u.ID, &u.AuthProvider, &u.ProviderID, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user: %w", err)
	}
	return u, nil
}

func (r *UserRepository) GetByProvider(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	const q = `SELECT id, auth_provider, provider_id, created_at FROM users WHERE auth_provider = $1 AND provider_id = $2`
	u := &domain.User{}
	err := r.pool.QueryRow(ctx, q, provider, providerID).Scan(&u.ID, &u.AuthProvider, &u.ProviderID, &u.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select user by provider: %w", err)
	}
	return u, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
