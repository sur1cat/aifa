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

type HabitRepository struct {
	pool *pgxpool.Pool
}

func NewHabitRepository(pool *pgxpool.Pool) *HabitRepository {
	return &HabitRepository{pool: pool}
}

func (r *HabitRepository) Create(ctx context.Context, habit *domain.Habit) error {
	query := `
		INSERT INTO habits (id, user_id, goal_id, title, icon, color, period, target_value, unit, created_at, archived_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	habit.ID = uuid.New()
	habit.CreatedAt = time.Now()
	habit.UpdatedAt = time.Now()

	_, err := r.pool.Exec(ctx, query,
		habit.ID, habit.UserID, habit.GoalID, habit.Title, habit.Icon, habit.Color, habit.Period,
		habit.TargetValue, habit.Unit,
		habit.CreatedAt, habit.ArchivedAt, habit.UpdatedAt,
	)
	return err
}

func (r *HabitRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Habit, error) {
	query := `
		SELECT id, user_id, goal_id, title, icon, color, period, target_value, unit, created_at, archived_at, updated_at
		FROM habits WHERE id = $1
	`
	habit := &domain.Habit{}
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&habit.ID, &habit.UserID, &habit.GoalID, &habit.Title, &habit.Icon, &habit.Color,
		&habit.Period, &habit.TargetValue, &habit.Unit, &habit.CreatedAt, &habit.ArchivedAt, &habit.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	habit.CompletedDates, err = r.getCompletions(ctx, habit.ID)
	if err != nil {
		return nil, err
	}

	habit.ProgressValues, err = r.getProgressValues(ctx, habit.ID)
	if err != nil {
		return nil, err
	}

	return habit, nil
}

func (r *HabitRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Habit, error) {
	query := `
		SELECT id, user_id, goal_id, title, icon, color, period, target_value, unit, created_at, archived_at, updated_at
		FROM habits WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []*domain.Habit
	var habitIDs []uuid.UUID

	for rows.Next() {
		habit := &domain.Habit{}
		if err := rows.Scan(
			&habit.ID, &habit.UserID, &habit.GoalID, &habit.Title, &habit.Icon, &habit.Color,
			&habit.Period, &habit.TargetValue, &habit.Unit, &habit.CreatedAt, &habit.ArchivedAt, &habit.UpdatedAt,
		); err != nil {
			return nil, err
		}
		habits = append(habits, habit)
		habitIDs = append(habitIDs, habit.ID)
	}

	if len(habitIDs) > 0 {
		completions, err := r.getCompletionsBatch(ctx, habitIDs)
		if err != nil {
			return nil, err
		}
		progressValues, err := r.getProgressValuesBatch(ctx, habitIDs)
		if err != nil {
			return nil, err
		}
		for _, habit := range habits {
			habit.CompletedDates = completions[habit.ID]
			habit.ProgressValues = progressValues[habit.ID]
		}
	}

	return habits, nil
}

func (r *HabitRepository) Update(ctx context.Context, habit *domain.Habit) error {
	query := `
		UPDATE habits SET goal_id = $2, title = $3, icon = $4, color = $5, period = $6, target_value = $7, unit = $8, archived_at = $9, updated_at = $10
		WHERE id = $1
	`
	habit.UpdatedAt = time.Now()
	_, err := r.pool.Exec(ctx, query,
		habit.ID, habit.GoalID, habit.Title, habit.Icon, habit.Color, habit.Period, habit.TargetValue, habit.Unit, habit.ArchivedAt, habit.UpdatedAt,
	)
	return err
}

func (r *HabitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM habits WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	return err
}

func (r *HabitRepository) AddCompletion(ctx context.Context, habitID uuid.UUID, date string) error {
	query := `
		INSERT INTO habit_completions (id, habit_id, completed_date, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (habit_id, completed_date) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), habitID, date, time.Now())
	return err
}

func (r *HabitRepository) RemoveCompletion(ctx context.Context, habitID uuid.UUID, date string) error {
	query := `DELETE FROM habit_completions WHERE habit_id = $1 AND completed_date = $2`
	_, err := r.pool.Exec(ctx, query, habitID, date)
	return err
}

func (r *HabitRepository) getCompletions(ctx context.Context, habitID uuid.UUID) ([]string, error) {
	query := `
		SELECT completed_date FROM habit_completions
		WHERE habit_id = $1
		ORDER BY completed_date DESC
	`
	rows, err := r.pool.Query(ctx, query, habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dates []string
	for rows.Next() {
		var date time.Time
		if err := rows.Scan(&date); err != nil {
			return nil, err
		}
		dates = append(dates, date.Format("2006-01-02"))
	}
	return dates, nil
}

func (r *HabitRepository) getCompletionsBatch(ctx context.Context, habitIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	query := `
		SELECT habit_id, completed_date FROM habit_completions
		WHERE habit_id = ANY($1)
		ORDER BY completed_date DESC
	`
	rows, err := r.pool.Query(ctx, query, habitIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var habitID uuid.UUID
		var date time.Time
		if err := rows.Scan(&habitID, &date); err != nil {
			return nil, err
		}
		result[habitID] = append(result[habitID], date.Format("2006-01-02"))
	}
	return result, nil
}

func (r *HabitRepository) VerifyOwnership(ctx context.Context, habitID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM habits WHERE id = $1 AND user_id = $2)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, habitID, userID).Scan(&exists)
	return exists, err
}

func (r *HabitRepository) getProgressValues(ctx context.Context, habitID uuid.UUID) (map[string]int, error) {
	query := `
		SELECT progress_date, progress_value FROM habit_progress
		WHERE habit_id = $1
	`
	rows, err := r.pool.Query(ctx, query, habitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var date time.Time
		var value int
		if err := rows.Scan(&date, &value); err != nil {
			return nil, err
		}
		result[date.Format("2006-01-02")] = value
	}
	return result, nil
}

func (r *HabitRepository) getProgressValuesBatch(ctx context.Context, habitIDs []uuid.UUID) (map[uuid.UUID]map[string]int, error) {
	query := `
		SELECT habit_id, progress_date, progress_value FROM habit_progress
		WHERE habit_id = ANY($1)
	`
	rows, err := r.pool.Query(ctx, query, habitIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[uuid.UUID]map[string]int)
	for rows.Next() {
		var habitID uuid.UUID
		var date time.Time
		var value int
		if err := rows.Scan(&habitID, &date, &value); err != nil {
			return nil, err
		}
		if result[habitID] == nil {
			result[habitID] = make(map[string]int)
		}
		result[habitID][date.Format("2006-01-02")] = value
	}
	return result, nil
}

func (r *HabitRepository) SetProgressValue(ctx context.Context, habitID uuid.UUID, date string, value int) error {
	query := `
		INSERT INTO habit_progress (id, habit_id, progress_date, progress_value, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (habit_id, progress_date) DO UPDATE SET progress_value = $4
	`
	_, err := r.pool.Exec(ctx, query, uuid.New(), habitID, date, value, time.Now())
	return err
}

func (r *HabitRepository) RemoveProgressValue(ctx context.Context, habitID uuid.UUID, date string) error {
	query := `DELETE FROM habit_progress WHERE habit_id = $1 AND progress_date = $2`
	_, err := r.pool.Exec(ctx, query, habitID, date)
	return err
}
