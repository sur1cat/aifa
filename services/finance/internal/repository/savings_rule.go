package repository

import (
	"context"
	"errors"

	"github.com/sur1cat/aifa/finance-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SavingsRuleRepository struct {
	pool *pgxpool.Pool
}

func NewSavingsRuleRepository(pool *pgxpool.Pool) *SavingsRuleRepository {
	return &SavingsRuleRepository{pool: pool}
}

const srCols = `id, user_id, kind, amount, period, goal_title, active, created_at, updated_at`

func scanRule(row pgx.Row, r *domain.SavingsRule) error {
	return row.Scan(&r.ID, &r.UserID, &r.Kind, &r.Amount, &r.Period, &r.GoalTitle, &r.Active, &r.CreatedAt, &r.UpdatedAt)
}

func (r *SavingsRuleRepository) List(ctx context.Context, userID uuid.UUID, activeOnly bool) ([]*domain.SavingsRule, error) {
	q := `SELECT ` + srCols + ` FROM savings_rules WHERE user_id = $1`
	if activeOnly {
		q += ` AND active = TRUE`
	}
	q += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*domain.SavingsRule
	for rows.Next() {
		rule := &domain.SavingsRule{}
		if err := rows.Scan(&rule.ID, &rule.UserID, &rule.Kind, &rule.Amount, &rule.Period,
			&rule.GoalTitle, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *SavingsRuleRepository) ListActiveByKind(ctx context.Context, userID uuid.UUID, kind domain.SavingsRuleKind) ([]*domain.SavingsRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+srCols+` FROM savings_rules WHERE user_id = $1 AND kind = $2 AND active = TRUE`,
		userID, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*domain.SavingsRule
	for rows.Next() {
		rule := &domain.SavingsRule{}
		if err := rows.Scan(&rule.ID, &rule.UserID, &rule.Kind, &rule.Amount, &rule.Period,
			&rule.GoalTitle, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, rows.Err()
}

func (r *SavingsRuleRepository) Create(ctx context.Context, rule *domain.SavingsRule) error {
	return scanRule(r.pool.QueryRow(ctx,
		`INSERT INTO savings_rules (user_id, kind, amount, period, goal_title)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+srCols,
		rule.UserID, rule.Kind, rule.Amount, rule.Period, rule.GoalTitle,
	), rule)
}

func (r *SavingsRuleRepository) Deactivate(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE savings_rules SET active = FALSE, updated_at = NOW() WHERE id = $1 AND user_id = $2`,
		id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SavingsRuleRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM savings_rules WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DailySpent возвращает сумму расходов за сегодня.
func (r *SavingsRuleRepository) DailySpent(ctx context.Context, userID uuid.UUID) (float64, error) {
	var total float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM transactions
		 WHERE user_id = $1 AND type = 'expense'
		   AND date::date = CURRENT_DATE`,
		userID,
	).Scan(&total)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return total, err
}
