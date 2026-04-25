package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sur1cat/aifa/habit-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type HabitRepository struct {
	pool *pgxpool.Pool
}

func NewHabitRepository(pool *pgxpool.Pool) *HabitRepository {
	return &HabitRepository{pool: pool}
}

const habitColumns = `id, user_id, goal_id, title, icon, color, period, kind, currency, financial_category, expected_amount, target_value, unit, created_at, updated_at, archived_at`

func scanHabit(row pgx.Row, h *domain.Habit) error {
	return row.Scan(
		&h.ID, &h.UserID, &h.GoalID, &h.Title, &h.Icon, &h.Color, &h.Period,
		&h.Kind, &h.Currency, &h.FinancialCategory, &h.ExpectedAmount,
		&h.TargetValue, &h.Unit, &h.CreatedAt, &h.UpdatedAt, &h.ArchivedAt,
	)
}

func (r *HabitRepository) Create(ctx context.Context, h *domain.Habit) error {
	h.ID = uuid.New()
	h.CreatedAt = time.Now()
	h.UpdatedAt = h.CreatedAt

	const q = `
		INSERT INTO habits (
			id, user_id, goal_id, title, icon, color, period,
			kind, currency, financial_category, expected_amount,
			target_value, unit, created_at, updated_at, archived_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`
	_, err := r.pool.Exec(ctx, q,
		h.ID, h.UserID, h.GoalID, h.Title, h.Icon, h.Color, h.Period,
		h.Kind, h.Currency, h.FinancialCategory, h.ExpectedAmount,
		h.TargetValue, h.Unit, h.CreatedAt, h.UpdatedAt, h.ArchivedAt,
	)
	if err != nil {
		return fmt.Errorf("insert habit: %w", err)
	}
	return nil
}

// GetOwnedByID returns the habit iff it belongs to userID, else ErrNotFound.
// Merges ownership check and fetch into a single query.
func (r *HabitRepository) GetOwnedByID(ctx context.Context, id, userID uuid.UUID) (*domain.Habit, error) {
	q := `SELECT ` + habitColumns + ` FROM habits WHERE id = $1 AND user_id = $2`

	h := &domain.Habit{}
	err := scanHabit(r.pool.QueryRow(ctx, q, id, userID), h)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("select habit: %w", err)
	}

	if err := r.hydrate(ctx, h); err != nil {
		return nil, err
	}
	return h, nil
}

func (r *HabitRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Habit, error) {
	q := `SELECT ` + habitColumns + ` FROM habits WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list habits: %w", err)
	}
	defer rows.Close()

	var habits []*domain.Habit
	var ids []uuid.UUID
	for rows.Next() {
		h := &domain.Habit{}
		if err := scanHabit(rows, h); err != nil {
			return nil, err
		}
		habits = append(habits, h)
		ids = append(ids, h.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return habits, nil
	}

	completions, progress, err := r.loadRelated(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, h := range habits {
		h.CompletedDates = completions[h.ID]
		h.ProgressValues = progress[h.ID]
	}
	return habits, nil
}

func (r *HabitRepository) Update(ctx context.Context, h *domain.Habit) error {
	h.UpdatedAt = time.Now()
	const q = `
		UPDATE habits
		SET goal_id = $2, title = $3, icon = $4, color = $5, period = $6,
		    kind = $7, currency = $8, financial_category = $9, expected_amount = $10,
		    target_value = $11, unit = $12, archived_at = $13, updated_at = $14
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, q,
		h.ID, h.GoalID, h.Title, h.Icon, h.Color, h.Period,
		h.Kind, h.Currency, h.FinancialCategory, h.ExpectedAmount,
		h.TargetValue, h.Unit, h.ArchivedAt, h.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update habit: %w", err)
	}
	return nil
}

func (r *HabitRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	res, err := r.pool.Exec(ctx, `DELETE FROM habits WHERE id = $1 AND user_id = $2`, id, userID)
	if err != nil {
		return fmt.Errorf("delete habit: %w", err)
	}
	if res.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// DeleteByUser is called from the user.deleted NATS handler.
func (r *HabitRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM habits WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete habits by user: %w", err)
	}
	return nil
}

// ListActiveUserIDs returns distinct user IDs that have at least one non-archived habit.
// Used by the cron.reminder.tick handler to fan out per-user reminder.due events.
func (r *HabitRepository) ListActiveUserIDs(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.pool.Query(ctx, `SELECT DISTINCT user_id FROM habits WHERE archived_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("list active user ids: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// ClearGoalRef is called from the goal.deleted NATS handler.
func (r *HabitRepository) ClearGoalRef(ctx context.Context, goalID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE habits SET goal_id = NULL WHERE goal_id = $1`, goalID)
	if err != nil {
		return fmt.Errorf("clear goal ref: %w", err)
	}
	return nil
}

func (r *HabitRepository) AddCompletion(ctx context.Context, habitID uuid.UUID, date string) error {
	const q = `
		INSERT INTO habit_completions (habit_id, completed_date)
		VALUES ($1, $2)
		ON CONFLICT (habit_id, completed_date) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, q, habitID, date)
	return err
}

func (r *HabitRepository) RemoveCompletion(ctx context.Context, habitID uuid.UUID, date string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM habit_completions WHERE habit_id = $1 AND completed_date = $2`, habitID, date)
	return err
}

func (r *HabitRepository) SetProgress(ctx context.Context, habitID uuid.UUID, date string, value int) error {
	const q = `
		INSERT INTO habit_progress (habit_id, progress_date, progress_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (habit_id, progress_date) DO UPDATE SET progress_value = EXCLUDED.progress_value
	`
	_, err := r.pool.Exec(ctx, q, habitID, date, value)
	return err
}

func (r *HabitRepository) RemoveProgress(ctx context.Context, habitID uuid.UUID, date string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM habit_progress WHERE habit_id = $1 AND progress_date = $2`, habitID, date)
	return err
}

func (r *HabitRepository) hydrate(ctx context.Context, h *domain.Habit) error {
	completions, progress, err := r.loadRelated(ctx, []uuid.UUID{h.ID})
	if err != nil {
		return err
	}
	h.CompletedDates = completions[h.ID]
	h.ProgressValues = progress[h.ID]
	return nil
}

// loadRelated fetches completions and progress in parallel — they are
// independent queries and dominate list-habit latency on cold caches.
func (r *HabitRepository) loadRelated(ctx context.Context, habitIDs []uuid.UUID) (map[uuid.UUID][]string, map[uuid.UUID]map[string]int, error) {
	var (
		completions map[uuid.UUID][]string
		progress    map[uuid.UUID]map[string]int
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		c, err := r.completionsBatch(gctx, habitIDs)
		if err != nil {
			return err
		}
		completions = c
		return nil
	})
	g.Go(func() error {
		p, err := r.progressBatch(gctx, habitIDs)
		if err != nil {
			return err
		}
		progress = p
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, nil, err
	}
	return completions, progress, nil
}

func (r *HabitRepository) completionsBatch(ctx context.Context, habitIDs []uuid.UUID) (map[uuid.UUID][]string, error) {
	const q = `
		SELECT habit_id, completed_date FROM habit_completions
		WHERE habit_id = ANY($1)
		ORDER BY completed_date DESC
	`
	rows, err := r.pool.Query(ctx, q, habitIDs)
	if err != nil {
		return nil, fmt.Errorf("load completions: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var id uuid.UUID
		var date time.Time
		if err := rows.Scan(&id, &date); err != nil {
			return nil, err
		}
		result[id] = append(result[id], date.Format("2006-01-02"))
	}
	return result, rows.Err()
}

func (r *HabitRepository) progressBatch(ctx context.Context, habitIDs []uuid.UUID) (map[uuid.UUID]map[string]int, error) {
	const q = `
		SELECT habit_id, progress_date, progress_value FROM habit_progress
		WHERE habit_id = ANY($1)
	`
	rows, err := r.pool.Query(ctx, q, habitIDs)
	if err != nil {
		return nil, fmt.Errorf("load progress: %w", err)
	}
	defer rows.Close()

	result := make(map[uuid.UUID]map[string]int)
	for rows.Next() {
		var id uuid.UUID
		var date time.Time
		var value int
		if err := rows.Scan(&id, &date, &value); err != nil {
			return nil, err
		}
		if result[id] == nil {
			result[id] = make(map[string]int)
		}
		result[id][date.Format("2006-01-02")] = value
	}
	return result, rows.Err()
}
