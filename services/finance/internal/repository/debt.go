package repository

import (
	"context"
	"errors"

	"github.com/sur1cat/aifa/finance-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DebtRepository struct {
	pool *pgxpool.Pool
}

func NewDebtRepository(pool *pgxpool.Pool) *DebtRepository {
	return &DebtRepository{pool: pool}
}

const debtColumns = `id, user_id, counterparty, direction, amount, original_amount, note, settled, created_at, updated_at`

func scanDebt(row pgx.Row, d *domain.Debt) error {
	return row.Scan(&d.ID, &d.UserID, &d.Counterparty, &d.Direction,
		&d.Amount, &d.OriginalAmount, &d.Note, &d.Settled, &d.CreatedAt, &d.UpdatedAt)
}

func (r *DebtRepository) List(ctx context.Context, userID uuid.UUID, settledOnly *bool) ([]*domain.Debt, error) {
	q := `SELECT ` + debtColumns + ` FROM debts WHERE user_id = $1`
	args := []any{userID}
	if settledOnly != nil {
		q += ` AND settled = $2`
		args = append(args, *settledOnly)
	}
	q += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var debts []*domain.Debt
	for rows.Next() {
		d := &domain.Debt{}
		if err := rows.Scan(&d.ID, &d.UserID, &d.Counterparty, &d.Direction,
			&d.Amount, &d.OriginalAmount, &d.Note, &d.Settled, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		debts = append(debts, d)
	}
	return debts, rows.Err()
}

func (r *DebtRepository) Get(ctx context.Context, id, userID uuid.UUID) (*domain.Debt, error) {
	d := &domain.Debt{}
	err := scanDebt(r.pool.QueryRow(ctx,
		`SELECT `+debtColumns+` FROM debts WHERE id = $1 AND user_id = $2`, id, userID), d)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return d, err
}

func (r *DebtRepository) Create(ctx context.Context, d *domain.Debt) error {
	return scanDebt(r.pool.QueryRow(ctx,
		`INSERT INTO debts (user_id, counterparty, direction, amount, original_amount, note)
		 VALUES ($1, $2, $3, $4, $4, $5)
		 RETURNING `+debtColumns,
		d.UserID, d.Counterparty, d.Direction, d.Amount, d.Note,
	), d)
}

func (r *DebtRepository) Patch(ctx context.Context, id, userID uuid.UUID, reduceBy float64) (*domain.Debt, error) {
	d := &domain.Debt{}
	err := scanDebt(r.pool.QueryRow(ctx,
		`UPDATE debts
		 SET amount     = GREATEST(0, amount - $3),
		     settled    = (GREATEST(0, amount - $3) = 0),
		     updated_at = NOW()
		 WHERE id = $1 AND user_id = $2
		 RETURNING `+debtColumns,
		id, userID, reduceBy,
	), d)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return d, err
}

func (r *DebtRepository) Settle(ctx context.Context, id, userID uuid.UUID) (*domain.Debt, error) {
	d := &domain.Debt{}
	err := scanDebt(r.pool.QueryRow(ctx,
		`UPDATE debts SET amount = 0, settled = TRUE, updated_at = NOW()
		 WHERE id = $1 AND user_id = $2
		 RETURNING `+debtColumns,
		id, userID,
	), d)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return d, err
}

func (r *DebtRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM debts WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}
