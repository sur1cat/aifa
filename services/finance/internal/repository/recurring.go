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

type RecurringRepository struct {
	pool *pgxpool.Pool
}

func NewRecurringRepository(pool *pgxpool.Pool) *RecurringRepository {
	return &RecurringRepository{pool: pool}
}

const recurringColumns = `id, user_id, title, amount, type, category, frequency, start_date, next_date, end_date, remaining_payments, is_active, created_at, updated_at`

func scanRecurring(row pgx.Row, r *domain.Recurring) error {
	var start, next time.Time
	var end *time.Time
	if err := row.Scan(
		&r.ID, &r.UserID, &r.Title, &r.Amount, &r.Type, &r.Category,
		&r.Frequency, &start, &next, &end, &r.RemainingPayments,
		&r.IsActive, &r.CreatedAt, &r.UpdatedAt,
	); err != nil {
		return err
	}
	r.StartDate = start.Format("2006-01-02")
	r.NextDate = next.Format("2006-01-02")
	if end != nil {
		s := end.Format("2006-01-02")
		r.EndDate = &s
	}
	return nil
}

func (r *RecurringRepository) Create(ctx context.Context, x *domain.Recurring) error {
	x.ID = uuid.New()
	x.CreatedAt = time.Now()
	x.UpdatedAt = x.CreatedAt
	const q = `
		INSERT INTO recurring_transactions
			(id, user_id, title, amount, type, category, frequency, start_date, next_date, end_date, remaining_payments, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	if _, err := r.pool.Exec(ctx, q,
		x.ID, x.UserID, x.Title, x.Amount, x.Type, x.Category, x.Frequency,
		x.StartDate, x.NextDate, x.EndDate, x.RemainingPayments, x.IsActive,
		x.CreatedAt, x.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert recurring: %w", err)
	}
	return nil
}

func (r *RecurringRepository) GetOwnedByID(ctx context.Context, id, userID uuid.UUID) (*domain.Recurring, error) {
	q := `SELECT ` + recurringColumns + ` FROM recurring_transactions WHERE id = $1 AND user_id = $2`
	x := &domain.Recurring{}
	err := scanRecurring(r.pool.QueryRow(ctx, q, id, userID), x)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select recurring: %w", err)
	}
	return x, nil
}

func (r *RecurringRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Recurring, error) {
	q := `SELECT ` + recurringColumns + ` FROM recurring_transactions WHERE user_id = $1 ORDER BY next_date ASC, created_at DESC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list recurring: %w", err)
	}
	defer rows.Close()
	var out []*domain.Recurring
	for rows.Next() {
		x := &domain.Recurring{}
		if err := scanRecurring(rows, x); err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, rows.Err()
}

func (r *RecurringRepository) Update(ctx context.Context, x *domain.Recurring) error {
	x.UpdatedAt = time.Now()
	const q = `
		UPDATE recurring_transactions
		SET title = $2, amount = $3, type = $4, category = $5, frequency = $6,
		    start_date = $7, next_date = $8, end_date = $9, remaining_payments = $10,
		    is_active = $11, updated_at = $12
		WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, q,
		x.ID, x.Title, x.Amount, x.Type, x.Category, x.Frequency,
		x.StartDate, x.NextDate, x.EndDate, x.RemainingPayments, x.IsActive, x.UpdatedAt,
	); err != nil {
		return fmt.Errorf("update recurring: %w", err)
	}
	return nil
}

func (r *RecurringRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM recurring_transactions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete recurring: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *RecurringRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM recurring_transactions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete recurring by user: %w", err)
	}
	return nil
}

// AdvanceIfDue atomically moves the recurring row forward by one occurrence
// when the stored next_date still equals expectedDate. Returns true only when
// the caller held the "race" — preventing double-processing under concurrent
// scheduler + client triggers.
func (r *RecurringRepository) AdvanceIfDue(ctx context.Context, id uuid.UUID, expectedDate, newDate string, remaining *int, isActive bool) (bool, error) {
	const q = `
		UPDATE recurring_transactions
		SET next_date = $3, remaining_payments = $4, is_active = $5, updated_at = NOW()
		WHERE id = $1 AND next_date = $2
	`
	res, err := r.pool.Exec(ctx, q, id, expectedDate, newDate, remaining, isActive)
	if err != nil {
		return false, fmt.Errorf("advance recurring: %w", err)
	}
	return res.RowsAffected() > 0, nil
}

// MonthlyProjection approximates recurring income/expense at a monthly cadence.
// 4.33 = avg weeks per month; 2.17 = avg biweekly cycles per month.
func (r *RecurringRepository) MonthlyProjection(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
	const q = `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN
				CASE frequency
					WHEN 'weekly'    THEN amount * 4.33
					WHEN 'biweekly'  THEN amount * 2.17
					WHEN 'monthly'   THEN amount
					WHEN 'quarterly' THEN amount / 3
					WHEN 'yearly'    THEN amount / 12
					ELSE amount
				END
			ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN
				CASE frequency
					WHEN 'weekly'    THEN amount * 4.33
					WHEN 'biweekly'  THEN amount * 2.17
					WHEN 'monthly'   THEN amount
					WHEN 'quarterly' THEN amount / 3
					WHEN 'yearly'    THEN amount / 12
					ELSE amount
				END
			ELSE 0 END), 0)
		FROM recurring_transactions
		WHERE user_id = $1 AND is_active = true
	`
	var income, expense float64
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&income, &expense); err != nil {
		return 0, 0, fmt.Errorf("monthly projection: %w", err)
	}
	return income, expense, nil
}
