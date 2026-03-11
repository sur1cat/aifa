package repository

import (
	"context"
	"errors"
	"time"

	"habitflow/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecurringTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewRecurringTransactionRepository(pool *pgxpool.Pool) *RecurringTransactionRepository {
	return &RecurringTransactionRepository{pool: pool}
}

func (r *RecurringTransactionRepository) Create(ctx context.Context, rt *domain.RecurringTransaction) error {
	query := `
		INSERT INTO recurring_transactions (id, user_id, title, amount, type, category, frequency, start_date, next_date, end_date, remaining_payments, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	rt.ID = uuid.New()
	rt.CreatedAt = time.Now()
	rt.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		rt.ID, rt.UserID, rt.Title, rt.Amount, rt.Type, rt.Category,
		rt.Frequency, rt.StartDate, rt.NextDate, rt.EndDate, rt.RemainingPayments,
		rt.IsActive, rt.CreatedAt, rt.UpdatedAt,
	)
	return err
}

func (r *RecurringTransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RecurringTransaction, error) {
	query := `
		SELECT id, user_id, title, amount, type, category, frequency, start_date, next_date, end_date, remaining_payments, is_active, created_at, updated_at
		FROM recurring_transactions WHERE id = $1
	`
	rt := &domain.RecurringTransaction{}
	var startDate, nextDate time.Time
	var endDate *time.Time
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&rt.ID, &rt.UserID, &rt.Title, &rt.Amount, &rt.Type, &rt.Category,
		&rt.Frequency, &startDate, &nextDate, &endDate, &rt.RemainingPayments,
		&rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	rt.StartDate = startDate.Format("2006-01-02")
	rt.NextDate = nextDate.Format("2006-01-02")
	if endDate != nil {
		ed := endDate.Format("2006-01-02")
		rt.EndDate = &ed
	}
	return rt, nil
}

func (r *RecurringTransactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RecurringTransaction, error) {
	query := `
		SELECT id, user_id, title, amount, type, category, frequency, start_date, next_date, end_date, remaining_payments, is_active, created_at, updated_at
		FROM recurring_transactions WHERE user_id = $1
		ORDER BY next_date ASC, created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.RecurringTransaction
	for rows.Next() {
		rt := &domain.RecurringTransaction{}
		var startDate, nextDate time.Time
		var endDate *time.Time
		if err := rows.Scan(
			&rt.ID, &rt.UserID, &rt.Title, &rt.Amount, &rt.Type, &rt.Category,
			&rt.Frequency, &startDate, &nextDate, &endDate, &rt.RemainingPayments,
			&rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rt.StartDate = startDate.Format("2006-01-02")
		rt.NextDate = nextDate.Format("2006-01-02")
		if endDate != nil {
			ed := endDate.Format("2006-01-02")
			rt.EndDate = &ed
		}
		transactions = append(transactions, rt)
	}

	return transactions, nil
}

func (r *RecurringTransactionRepository) GetActiveByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RecurringTransaction, error) {
	query := `
		SELECT id, user_id, title, amount, type, category, frequency, start_date, next_date, end_date, remaining_payments, is_active, created_at, updated_at
		FROM recurring_transactions WHERE user_id = $1 AND is_active = true
		ORDER BY next_date ASC, created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.RecurringTransaction
	for rows.Next() {
		rt := &domain.RecurringTransaction{}
		var startDate, nextDate time.Time
		var endDate *time.Time
		if err := rows.Scan(
			&rt.ID, &rt.UserID, &rt.Title, &rt.Amount, &rt.Type, &rt.Category,
			&rt.Frequency, &startDate, &nextDate, &endDate, &rt.RemainingPayments,
			&rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt,
		); err != nil {
			return nil, err
		}
		rt.StartDate = startDate.Format("2006-01-02")
		rt.NextDate = nextDate.Format("2006-01-02")
		if endDate != nil {
			ed := endDate.Format("2006-01-02")
			rt.EndDate = &ed
		}
		transactions = append(transactions, rt)
	}

	return transactions, nil
}

func (r *RecurringTransactionRepository) Update(ctx context.Context, rt *domain.RecurringTransaction) error {
	query := `
		UPDATE recurring_transactions
		SET title = $2, amount = $3, type = $4, category = $5, frequency = $6,
		    start_date = $7, next_date = $8, end_date = $9, remaining_payments = $10,
		    is_active = $11, updated_at = $12
		WHERE id = $1
	`
	rt.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		rt.ID, rt.Title, rt.Amount, rt.Type, rt.Category,
		rt.Frequency, rt.StartDate, rt.NextDate, rt.EndDate, rt.RemainingPayments,
		rt.IsActive, rt.UpdatedAt,
	)
	return err
}

func (r *RecurringTransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM recurring_transactions WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *RecurringTransactionRepository) VerifyOwnership(ctx context.Context, rtID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM recurring_transactions WHERE id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, rtID, userID).Scan(&exists)
	return exists, err
}

func (r *RecurringTransactionRepository) UpdateNextDateAtomic(ctx context.Context, id uuid.UUID, expectedDate, newDate string, remainingPayments *int, isActive bool) (bool, error) {
	query := `
		UPDATE recurring_transactions
		SET next_date = $3, remaining_payments = $4, is_active = $5, updated_at = $6
		WHERE id = $1 AND next_date = $2
	`
	result, err := r.pool.Exec(ctx, query, id, expectedDate, newDate, remainingPayments, isActive, time.Now())
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (r *RecurringTransactionRepository) GetMonthlyProjection(ctx context.Context, userID uuid.UUID) (income float64, expense float64, err error) {
	query := `
		SELECT
			COALESCE(SUM(CASE
				WHEN type = 'income' THEN
					CASE frequency
						WHEN 'weekly' THEN amount * 4.33
						WHEN 'biweekly' THEN amount * 2.17
						WHEN 'monthly' THEN amount
						WHEN 'quarterly' THEN amount / 3
						WHEN 'yearly' THEN amount / 12
						ELSE amount
					END
				ELSE 0
			END), 0) as income,
			COALESCE(SUM(CASE
				WHEN type = 'expense' THEN
					CASE frequency
						WHEN 'weekly' THEN amount * 4.33
						WHEN 'biweekly' THEN amount * 2.17
						WHEN 'monthly' THEN amount
						WHEN 'quarterly' THEN amount / 3
						WHEN 'yearly' THEN amount / 12
						ELSE amount
					END
				ELSE 0
			END), 0) as expense
		FROM recurring_transactions
		WHERE user_id = $1 AND is_active = true
	`
	err = r.pool.QueryRow(ctx, query, userID).Scan(&income, &expense)
	return
}
