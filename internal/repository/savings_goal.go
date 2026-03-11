package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"habitflow/internal/domain"
)

type SavingsGoalRepository struct {
	pool *pgxpool.Pool
}

func NewSavingsGoalRepository(pool *pgxpool.Pool) *SavingsGoalRepository {
	return &SavingsGoalRepository{pool: pool}
}

func (r *SavingsGoalRepository) Get(ctx context.Context, userID uuid.UUID) (*domain.SavingsGoal, error) {
	query := `
		SELECT id, user_id, monthly_target, created_at, updated_at
		FROM savings_goals
		WHERE user_id = $1
	`

	var goal domain.SavingsGoal
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&goal.ID,
		&goal.UserID,
		&goal.MonthlyTarget,
		&goal.CreatedAt,
		&goal.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &goal, nil
}

func (r *SavingsGoalRepository) Upsert(ctx context.Context, goal *domain.SavingsGoal) error {
	query := `
		INSERT INTO savings_goals (id, user_id, monthly_target, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			monthly_target = EXCLUDED.monthly_target,
			updated_at = NOW()
		RETURNING id, created_at, updated_at
	`

	if goal.ID == uuid.Nil {
		goal.ID = uuid.New()
	}
	if goal.CreatedAt.IsZero() {
		goal.CreatedAt = time.Now()
	}
	goal.UpdatedAt = time.Now()

	return r.pool.QueryRow(ctx, query,
		goal.ID,
		goal.UserID,
		goal.MonthlyTarget,
		goal.CreatedAt,
		goal.UpdatedAt,
	).Scan(&goal.ID, &goal.CreatedAt, &goal.UpdatedAt)
}

func (r *SavingsGoalRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM savings_goals WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *SavingsGoalRepository) GetCurrentSavings(ctx context.Context, userID uuid.UUID) (float64, float64, float64, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as expenses
		FROM transactions
		WHERE user_id = $1
			AND date_trunc('month', date) = date_trunc('month', CURRENT_DATE)
	`

	var income, expenses float64
	err := r.pool.QueryRow(ctx, query, userID).Scan(&income, &expenses)
	if err != nil {
		return 0, 0, 0, err
	}

	savings := income - expenses
	return income, expenses, savings, nil
}
