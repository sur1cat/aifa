package repository

import (
	"context"
	"errors"
	"time"

	"github.com/sur1cat/aifa/finance-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BudgetRepository struct {
	pool *pgxpool.Pool
}

func NewBudgetRepository(pool *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{pool: pool}
}

func scanBudget(row pgx.Row, b *domain.Budget) error {
	return row.Scan(&b.ID, &b.UserID, &b.Category, &b.MonthlyLimit, &b.CreatedAt, &b.UpdatedAt)
}

func (r *BudgetRepository) List(ctx context.Context, userID uuid.UUID) ([]*domain.Budget, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, category, monthly_limit, created_at, updated_at
		 FROM budgets WHERE user_id = $1 ORDER BY category`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []*domain.Budget
	for rows.Next() {
		b := &domain.Budget{}
		if err := rows.Scan(&b.ID, &b.UserID, &b.Category, &b.MonthlyLimit, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}
	return budgets, rows.Err()
}

func (r *BudgetRepository) Get(ctx context.Context, id, userID uuid.UUID) (*domain.Budget, error) {
	b := &domain.Budget{}
	err := scanBudget(r.pool.QueryRow(ctx,
		`SELECT id, user_id, category, monthly_limit, created_at, updated_at
		 FROM budgets WHERE id = $1 AND user_id = $2`, id, userID), b)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return b, err
}

func (r *BudgetRepository) GetByCategory(ctx context.Context, userID uuid.UUID, category string) (*domain.Budget, error) {
	b := &domain.Budget{}
	err := scanBudget(r.pool.QueryRow(ctx,
		`SELECT id, user_id, category, monthly_limit, created_at, updated_at
		 FROM budgets WHERE user_id = $1 AND category = $2`, userID, category), b)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return b, err
}

func (r *BudgetRepository) Create(ctx context.Context, b *domain.Budget) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO budgets (user_id, category, monthly_limit)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, category) DO UPDATE
		   SET monthly_limit = EXCLUDED.monthly_limit, updated_at = NOW()
		 RETURNING id, user_id, category, monthly_limit, created_at, updated_at`,
		b.UserID, b.Category, b.MonthlyLimit,
	).Scan(&b.ID, &b.UserID, &b.Category, &b.MonthlyLimit, &b.CreatedAt, &b.UpdatedAt)
}

func (r *BudgetRepository) Update(ctx context.Context, b *domain.Budget) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE budgets SET monthly_limit = $1, updated_at = NOW()
		 WHERE id = $2 AND user_id = $3`,
		b.MonthlyLimit, b.ID, b.UserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM budgets WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// MonthlySpent возвращает сумму расходов пользователя по категории за текущий месяц.
func (r *BudgetRepository) MonthlySpent(ctx context.Context, userID uuid.UUID, category string) (float64, error) {
	now := time.Now()
	var total float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0)
		 FROM transactions
		 WHERE user_id = $1
		   AND category = $2
		   AND type = 'expense'
		   AND date_trunc('month', date::timestamptz) = date_trunc('month', $3::timestamptz)`,
		userID, category, now,
	).Scan(&total)
	return total, err
}
