package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sur1cat/aifa/goal-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type GoalRepository struct {
	pool *pgxpool.Pool
}

func NewGoalRepository(pool *pgxpool.Pool) *GoalRepository {
	return &GoalRepository{pool: pool}
}

const goalColumns = `id, user_id, title, icon, goal_type, target_amount, current_amount, currency, deadline, archived_at, created_at, updated_at`

func scanGoal(row pgx.Row, g *domain.Goal) error {
	return row.Scan(
		&g.ID, &g.UserID, &g.Title, &g.Icon, &g.GoalType,
		&g.TargetAmount, &g.CurrentAmount, &g.Currency,
		&g.Deadline, &g.ArchivedAt, &g.CreatedAt, &g.UpdatedAt,
	)
}

func (r *GoalRepository) Create(ctx context.Context, g *domain.Goal) error {
	g.ID = uuid.New()
	g.CreatedAt = time.Now()
	g.UpdatedAt = g.CreatedAt

	const q = `
		INSERT INTO goals (
			id, user_id, title, icon, goal_type,
			target_amount, current_amount, currency,
			deadline, archived_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, q,
		g.ID, g.UserID, g.Title, g.Icon, g.GoalType,
		g.TargetAmount, g.CurrentAmount, g.Currency,
		g.Deadline, g.ArchivedAt, g.CreatedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert goal: %w", err)
	}
	return nil
}

func (r *GoalRepository) GetOwnedByID(ctx context.Context, id, userID uuid.UUID) (*domain.Goal, error) {
	q := `SELECT ` + goalColumns + ` FROM goals WHERE id = $1 AND user_id = $2`

	g := &domain.Goal{}
	err := scanGoal(r.pool.QueryRow(ctx, q, id, userID), g)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select goal: %w", err)
	}
	return g, nil
}

func (r *GoalRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Goal, error) {
	q := `SELECT ` + goalColumns + ` FROM goals WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list goals: %w", err)
	}
	defer rows.Close()

	var goals []*domain.Goal
	for rows.Next() {
		g := &domain.Goal{}
		if err := scanGoal(rows, g); err != nil {
			return nil, err
		}
		goals = append(goals, g)
	}
	return goals, rows.Err()
}

func (r *GoalRepository) Update(ctx context.Context, g *domain.Goal) error {
	g.UpdatedAt = time.Now()
	const q = `
		UPDATE goals
		SET title = $2, icon = $3, goal_type = $4,
		    target_amount = $5, current_amount = $6, currency = $7,
		    deadline = $8, archived_at = $9, updated_at = $10
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, q,
		g.ID, g.Title, g.Icon, g.GoalType,
		g.TargetAmount, g.CurrentAmount, g.Currency,
		g.Deadline, g.ArchivedAt, g.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update goal: %w", err)
	}
	return nil
}

func (r *GoalRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM goals WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete goal: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *GoalRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM goals WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete goals by user: %w", err)
	}
	return nil
}
