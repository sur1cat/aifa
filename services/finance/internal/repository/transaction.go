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

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

const txColumns = `id, user_id, title, amount, type, category, date, created_at, updated_at`

func scanTransaction(row pgx.Row, t *domain.Transaction) error {
	var date time.Time
	if err := row.Scan(
		&t.ID, &t.UserID, &t.Title, &t.Amount, &t.Type, &t.Category,
		&date, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return err
	}
	t.Date = date.Format("2006-01-02")
	return nil
}

func (r *TransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	const q = `
		INSERT INTO transactions (id, user_id, title, amount, type, category, date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	if _, err := r.pool.Exec(ctx, q,
		t.ID, t.UserID, t.Title, t.Amount, t.Type, t.Category, t.Date, t.CreatedAt, t.UpdatedAt,
	); err != nil {
		return fmt.Errorf("insert transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) GetOwnedByID(ctx context.Context, id, userID uuid.UUID) (*domain.Transaction, error) {
	q := `SELECT ` + txColumns + ` FROM transactions WHERE id = $1 AND user_id = $2`
	t := &domain.Transaction{}
	err := scanTransaction(r.pool.QueryRow(ctx, q, id, userID), t)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select transaction: %w", err)
	}
	return t, nil
}

func (r *TransactionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error) {
	q := `SELECT ` + txColumns + ` FROM transactions WHERE user_id = $1 ORDER BY date DESC, created_at DESC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()
	var out []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		if err := scanTransaction(rows, t); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TransactionRepository) ListByUserAndMonth(ctx context.Context, userID uuid.UUID, year, month, limit, offset int) ([]*domain.Transaction, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = $1
		   AND EXTRACT(YEAR FROM date) = $2 AND EXTRACT(MONTH FROM date) = $3`,
		userID, year, month,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count transactions: %w", err)
	}

	q := `SELECT ` + txColumns + ` FROM transactions WHERE user_id = $1
	      AND EXTRACT(YEAR FROM date) = $2 AND EXTRACT(MONTH FROM date) = $3
	      ORDER BY date DESC, created_at DESC LIMIT $4 OFFSET $5`
	rows, err := r.pool.Query(ctx, q, userID, year, month, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list transactions by month: %w", err)
	}
	defer rows.Close()
	var out []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		if err := scanTransaction(rows, t); err != nil {
			return nil, 0, err
		}
		out = append(out, t)
	}
	return out, total, rows.Err()
}

func (r *TransactionRepository) Update(ctx context.Context, t *domain.Transaction) error {
	t.UpdatedAt = time.Now()
	const q = `
		UPDATE transactions
		SET title = $2, amount = $3, type = $4, category = $5, date = $6, updated_at = $7
		WHERE id = $1
	`
	if _, err := r.pool.Exec(ctx, q, t.ID, t.Title, t.Amount, t.Type, t.Category, t.Date, t.UpdatedAt); err != nil {
		return fmt.Errorf("update transaction: %w", err)
	}
	return nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM transactions WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TransactionRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM transactions WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete transactions by user: %w", err)
	}
	return nil
}

// SumMonth returns (income, expense) for the given calendar month.
func (r *TransactionRepository) SumMonth(ctx context.Context, userID uuid.UUID, year, month int) (float64, float64, error) {
	const q = `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1
		  AND EXTRACT(YEAR FROM date) = $2
		  AND EXTRACT(MONTH FROM date) = $3
	`
	var income, expense float64
	if err := r.pool.QueryRow(ctx, q, userID, year, month).Scan(&income, &expense); err != nil {
		return 0, 0, fmt.Errorf("sum month: %w", err)
	}
	return income, expense, nil
}

// SumCurrentMonth is a specialization matching Postgres' date_trunc for the
// running calendar month — used by the savings-goal progress calculation.
func (r *TransactionRepository) SumCurrentMonth(ctx context.Context, userID uuid.UUID) (float64, float64, error) {
	const q = `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income'  THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = $1
		  AND date_trunc('month', date) = date_trunc('month', CURRENT_DATE)
	`
	var income, expense float64
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&income, &expense); err != nil {
		return 0, 0, fmt.Errorf("sum current month: %w", err)
	}
	return income, expense, nil
}
