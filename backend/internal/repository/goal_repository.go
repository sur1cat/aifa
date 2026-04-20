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

type GoalRepository struct {
	pool *pgxpool.Pool
}

func NewGoalRepository(pool *pgxpool.Pool) *GoalRepository {
	return &GoalRepository{pool: pool}
}

// Create creates a new goal
func (r *GoalRepository) Create(ctx context.Context, goal *domain.Goal) error {
	query := `
		INSERT INTO goals (id, user_id, title, icon, target_value, unit, deadline, created_at, archived_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	goal.ID = uuid.New()
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		goal.ID, goal.UserID, goal.Title, goal.Icon,
		goal.TargetValue, goal.Unit, goal.Deadline,
		goal.CreatedAt, goal.ArchivedAt, goal.UpdatedAt,
	)
	return err
}

// GetByID returns a goal by ID
func (r *GoalRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Goal, error) {
	query := `
		SELECT id, user_id, title, icon, target_value, unit, deadline, created_at, archived_at, updated_at
		FROM goals WHERE id = $1
	`
	goal := &domain.Goal{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&goal.ID, &goal.UserID, &goal.Title, &goal.Icon,
		&goal.TargetValue, &goal.Unit, &goal.Deadline,
		&goal.CreatedAt, &goal.ArchivedAt, &goal.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return goal, nil
}

// GetByUserID returns all goals for a user
func (r *GoalRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Goal, error) {
	query := `
		SELECT id, user_id, title, icon, target_value, unit, deadline, created_at, archived_at, updated_at
		FROM goals WHERE user_id = $1 AND archived_at IS NULL
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var goals []*domain.Goal
	for rows.Next() {
		goal := &domain.Goal{}
		if err := rows.Scan(
			&goal.ID, &goal.UserID, &goal.Title, &goal.Icon,
			&goal.TargetValue, &goal.Unit, &goal.Deadline,
			&goal.CreatedAt, &goal.ArchivedAt, &goal.UpdatedAt,
		); err != nil {
			return nil, err
		}
		goals = append(goals, goal)
	}
	return goals, nil
}

// Update updates a goal
func (r *GoalRepository) Update(ctx context.Context, goal *domain.Goal) error {
	query := `
		UPDATE goals SET title = $2, icon = $3, target_value = $4, unit = $5, deadline = $6, archived_at = $7, updated_at = $8
		WHERE id = $1
	`
	goal.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		goal.ID, goal.Title, goal.Icon, goal.TargetValue, goal.Unit, goal.Deadline, goal.ArchivedAt, goal.UpdatedAt,
	)
	return err
}

// Delete deletes a goal
func (r *GoalRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE goals SET archived_at = $2, updated_at = $2 WHERE id = $1 AND archived_at IS NULL`
	_, err := r.pool.Exec(ctx, query, id, time.Now())
	return err
}

// VerifyOwnership checks if a goal belongs to a user
func (r *GoalRepository) VerifyOwnership(ctx context.Context, goalID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM goals WHERE id = $1 AND user_id = $2 AND archived_at IS NULL)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, goalID, userID).Scan(&exists)
	return exists, err
}
