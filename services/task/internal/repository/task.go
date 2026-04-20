package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sur1cat/aifa/task-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

const (
	taskColumns = `id, user_id, title, is_completed, priority, due_date, created_at, updated_at`
	// DATE columns come back as time.Time; we format back to ISO in scanTask.
	// ORDER BY priority uses a CASE so we don't rely on lexical order of the enum.
	priorityOrder = `
		CASE priority
			WHEN 'urgent' THEN 0
			WHEN 'high'   THEN 1
			WHEN 'medium' THEN 2
			WHEN 'low'    THEN 3
		END`
)

func scanTask(row pgx.Row, t *domain.Task) error {
	var due time.Time
	if err := row.Scan(
		&t.ID, &t.UserID, &t.Title, &t.IsCompleted, &t.Priority,
		&due, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		return err
	}
	t.DueDate = due.Format("2006-01-02")
	return nil
}

func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) error {
	t.ID = uuid.New()
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt

	const q = `
		INSERT INTO tasks (id, user_id, title, is_completed, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, q,
		t.ID, t.UserID, t.Title, t.IsCompleted, t.Priority, t.DueDate, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert task: %w", err)
	}
	return nil
}

func (r *TaskRepository) GetOwnedByID(ctx context.Context, id, userID uuid.UUID) (*domain.Task, error) {
	q := `SELECT ` + taskColumns + ` FROM tasks WHERE id = $1 AND user_id = $2`

	t := &domain.Task{}
	err := scanTask(r.pool.QueryRow(ctx, q, id, userID), t)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select task: %w", err)
	}
	return t, nil
}

func (r *TaskRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
	q := `SELECT ` + taskColumns + ` FROM tasks WHERE user_id = $1 ORDER BY ` + priorityOrder + `, due_date ASC, created_at DESC`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t := &domain.Task{}
		if err := scanTask(rows, t); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// ListByUserAndDate returns tasks for the given due date with pagination.
// Incomplete tasks first, then by priority, then newest.
func (r *TaskRepository) ListByUserAndDate(ctx context.Context, userID uuid.UUID, date string, limit, offset int) ([]*domain.Task, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date = $2`,
		userID, date,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count tasks: %w", err)
	}

	q := `SELECT ` + taskColumns + ` FROM tasks WHERE user_id = $1 AND due_date = $2 ` +
		`ORDER BY is_completed ASC, ` + priorityOrder + `, created_at DESC LIMIT $3 OFFSET $4`

	rows, err := r.pool.Query(ctx, q, userID, date, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list tasks for date: %w", err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		t := &domain.Task{}
		if err := scanTask(rows, t); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *TaskRepository) Update(ctx context.Context, t *domain.Task) error {
	t.UpdatedAt = time.Now()
	const q = `
		UPDATE tasks
		SET title = $2, is_completed = $3, priority = $4, due_date = $5, updated_at = $6
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, q, t.ID, t.Title, t.IsCompleted, t.Priority, t.DueDate, t.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TaskRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete tasks by user: %w", err)
	}
	return nil
}

// ToggleCompleted flips is_completed atomically for a user-owned task and
// returns the updated row. Ownership is enforced in the WHERE clause so we
// don't need a separate read + write.
func (r *TaskRepository) ToggleCompleted(ctx context.Context, id, userID uuid.UUID) (*domain.Task, error) {
	q := `
		UPDATE tasks SET is_completed = NOT is_completed, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
		RETURNING ` + taskColumns
	t := &domain.Task{}
	err := scanTask(r.pool.QueryRow(ctx, q, id, userID), t)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("toggle task: %w", err)
	}
	return t, nil
}
