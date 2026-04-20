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

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

// Create creates a new transaction
func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, title, amount, type, category, date, created_at, updated_at, deleted_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NULL)
	`
	tx.ID = uuid.New()
	tx.CreatedAt = time.Now()
	tx.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.UserID, tx.Title, tx.Amount, tx.Type, tx.Category,
		tx.Date, tx.CreatedAt, tx.UpdatedAt,
	)
	return err
}

// GetByID returns a transaction by ID
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Transaction, error) {
	query := `
		SELECT id, user_id, title, amount, type, category, date, created_at, updated_at
		FROM transactions WHERE id = $1 AND deleted_at IS NULL
	`
	tx := &domain.Transaction{}
	var date time.Time
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID, &tx.UserID, &tx.Title, &tx.Amount, &tx.Type, &tx.Category,
		&date, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	tx.Date = date.Format("2006-01-02")
	return tx, nil
}

// GetByUserID returns all transactions for a user
func (r *TransactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Transaction, error) {
	query := `
		SELECT id, user_id, title, amount, type, category, date, created_at, updated_at
		FROM transactions WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY date DESC, created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		var date time.Time
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.Title, &tx.Amount, &tx.Type, &tx.Category,
			&date, &tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tx.Date = date.Format("2006-01-02")
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// GetByUserIDAndMonth returns transactions for a user in a specific month with pagination
func (r *TransactionRepository) GetByUserIDAndMonth(ctx context.Context, userID uuid.UUID, year int, month int, limit, offset int) ([]*domain.Transaction, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*) FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL AND EXTRACT(YEAR FROM date) = $2 AND EXTRACT(MONTH FROM date) = $3
	`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID, year, month).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, title, amount, type, category, date, created_at, updated_at
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL AND EXTRACT(YEAR FROM date) = $2 AND EXTRACT(MONTH FROM date) = $3
		ORDER BY date DESC, created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.pool.Query(ctx, query, userID, year, month, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []*domain.Transaction
	for rows.Next() {
		tx := &domain.Transaction{}
		var date time.Time
		if err := rows.Scan(
			&tx.ID, &tx.UserID, &tx.Title, &tx.Amount, &tx.Type, &tx.Category,
			&date, &tx.CreatedAt, &tx.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		tx.Date = date.Format("2006-01-02")
		transactions = append(transactions, tx)
	}

	return transactions, total, nil
}

// Update updates a transaction
func (r *TransactionRepository) Update(ctx context.Context, tx *domain.Transaction) error {
	query := `
		UPDATE transactions SET title = $2, amount = $3, type = $4, category = $5, date = $6, updated_at = $7
		WHERE id = $1
	`
	tx.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.Title, tx.Amount, tx.Type, tx.Category, tx.Date, tx.UpdatedAt,
	)
	return err
}

// Delete deletes a transaction
func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE transactions SET deleted_at = $2, updated_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id, time.Now())
	return err
}

// VerifyOwnership checks if a transaction belongs to a user
func (r *TransactionRepository) VerifyOwnership(ctx context.Context, txID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, txID, userID).Scan(&exists)
	return exists, err
}

// GetSummary returns income and expense totals for a user in a specific month
func (r *TransactionRepository) GetSummary(ctx context.Context, userID uuid.UUID, year int, month int) (income float64, expense float64, err error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0) as expense
		FROM transactions
		WHERE user_id = $1 AND deleted_at IS NULL AND EXTRACT(YEAR FROM date) = $2 AND EXTRACT(MONTH FROM date) = $3
	`
	err = r.pool.QueryRow(ctx, query, userID, year, month).Scan(&income, &expense)
	return
}
