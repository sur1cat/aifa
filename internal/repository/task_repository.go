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

type TaskRepository struct {
	pool *pgxpool.Pool
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{pool: pool}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (id, user_id, title, is_completed, priority, due_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	task.ID = uuid.New()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		task.ID, task.UserID, task.Title, task.IsCompleted, task.Priority,
		task.DueDate, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *TaskRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		SELECT id, user_id, title, is_completed, priority, due_date, created_at, updated_at
		FROM tasks WHERE id = $1
	`
	task := &domain.Task{}
	var dueDate time.Time
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&task.ID, &task.UserID, &task.Title, &task.IsCompleted, &task.Priority,
		&dueDate, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	task.DueDate = dueDate.Format("2006-01-02")
	return task, nil
}

func (r *TaskRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
	query := `
		SELECT id, user_id, title, is_completed, priority, due_date, created_at, updated_at
		FROM tasks WHERE user_id = $1
		ORDER BY
			CASE priority
				WHEN 'urgent' THEN 0
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			due_date ASC,
			created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		var dueDate time.Time
		if err := rows.Scan(
			&task.ID, &task.UserID, &task.Title, &task.IsCompleted, &task.Priority,
			&dueDate, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, err
		}
		task.DueDate = dueDate.Format("2006-01-02")
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) GetByUserIDAndDate(ctx context.Context, userID uuid.UUID, date string, limit, offset int) ([]*domain.Task, int, error) {

	countQuery := `SELECT COUNT(*) FROM tasks WHERE user_id = $1 AND due_date = $2`
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, userID, date).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, user_id, title, is_completed, priority, due_date, created_at, updated_at
		FROM tasks WHERE user_id = $1 AND due_date = $2
		ORDER BY
			is_completed ASC,
			CASE priority
				WHEN 'urgent' THEN 0
				WHEN 'high' THEN 1
				WHEN 'medium' THEN 2
				WHEN 'low' THEN 3
			END,
			created_at DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.pool.Query(ctx, query, userID, date, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		var dueDate time.Time
		if err := rows.Scan(
			&task.ID, &task.UserID, &task.Title, &task.IsCompleted, &task.Priority,
			&dueDate, &task.CreatedAt, &task.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		task.DueDate = dueDate.Format("2006-01-02")
		tasks = append(tasks, task)
	}

	return tasks, total, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `
		UPDATE tasks SET title = $2, is_completed = $3, priority = $4, due_date = $5, updated_at = $6
		WHERE id = $1
	`
	task.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		task.ID, task.Title, task.IsCompleted, task.Priority, task.DueDate, task.UpdatedAt,
	)
	return err
}

func (r *TaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *TaskRepository) ToggleCompleted(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	query := `
		UPDATE tasks SET is_completed = NOT is_completed, updated_at = $2
		WHERE id = $1
		RETURNING id, user_id, title, is_completed, priority, due_date, created_at, updated_at
	`
	task := &domain.Task{}
	var dueDate time.Time
	err := r.pool.QueryRow(ctx, query, id, time.Now()).Scan(
		&task.ID, &task.UserID, &task.Title, &task.IsCompleted, &task.Priority,
		&dueDate, &task.CreatedAt, &task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	task.DueDate = dueDate.Format("2006-01-02")
	return task, nil
}

func (r *TaskRepository) VerifyOwnership(ctx context.Context, taskID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM tasks WHERE id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, taskID, userID).Scan(&exists)
	return exists, err
}
