package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sur1cat/aifa/finance-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SavingsRepository struct {
	pool *pgxpool.Pool
}

func NewSavingsRepository(pool *pgxpool.Pool) *SavingsRepository {
	return &SavingsRepository{pool: pool}
}

// Get returns the user's savings goal, or (nil, nil) if none is set.
// (The resource is optional, so "no goal" is a normal state — not an error.)
func (r *SavingsRepository) Get(ctx context.Context, userID uuid.UUID) (*domain.SavingsGoal, error) {
	const q = `
		SELECT id, user_id, monthly_target, created_at, updated_at
		FROM savings_goals WHERE user_id = $1
	`
	g := &domain.SavingsGoal{}
	err := r.pool.QueryRow(ctx, q, userID).Scan(
		&g.ID, &g.UserID, &g.MonthlyTarget, &g.CreatedAt, &g.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("select savings goal: %w", err)
	}
	return g, nil
}

func (r *SavingsRepository) Upsert(ctx context.Context, g *domain.SavingsGoal) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	now := time.Now()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	g.UpdatedAt = now

	const q = `
		INSERT INTO savings_goals (id, user_id, monthly_target, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			monthly_target = EXCLUDED.monthly_target,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`
	if err := r.pool.QueryRow(ctx, q,
		g.ID, g.UserID, g.MonthlyTarget, g.CreatedAt, g.UpdatedAt,
	).Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return fmt.Errorf("upsert savings goal: %w", err)
	}
	return nil
}

func (r *SavingsRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM savings_goals WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete savings goal: %w", err)
	}
	return nil
}

func (r *SavingsRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.Delete(ctx, userID)
}
