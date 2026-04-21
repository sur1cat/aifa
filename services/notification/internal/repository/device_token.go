package repository

import (
	"context"
	"fmt"

	"github.com/sur1cat/aifa/notification-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DeviceTokenRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceTokenRepository(pool *pgxpool.Pool) *DeviceTokenRepository {
	return &DeviceTokenRepository{pool: pool}
}

// Register upserts the (user_id, token, platform) tuple. The token column is
// unique, so a re-register from a different user (rare — phone reset) rebinds
// it to the current owner.
func (r *DeviceTokenRepository) Register(ctx context.Context, userID uuid.UUID, token, platform string) error {
	const q = `
		INSERT INTO device_tokens (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE SET
			user_id    = EXCLUDED.user_id,
			platform   = EXCLUDED.platform,
			updated_at = NOW()
	`
	if _, err := r.pool.Exec(ctx, q, userID, token, platform); err != nil {
		return fmt.Errorf("register device token: %w", err)
	}
	return nil
}

func (r *DeviceTokenRepository) Unregister(ctx context.Context, token string) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM device_tokens WHERE token = $1`, token); err != nil {
		return fmt.Errorf("unregister device token: %w", err)
	}
	return nil
}

func (r *DeviceTokenRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.DeviceToken, error) {
	const q = `
		SELECT id, user_id, token, platform, created_at, updated_at
		FROM device_tokens WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list device tokens: %w", err)
	}
	defer rows.Close()

	out := make([]domain.DeviceToken, 0)
	for rows.Next() {
		var t domain.DeviceToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *DeviceTokenRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	if _, err := r.pool.Exec(ctx, `DELETE FROM device_tokens WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete device tokens by user: %w", err)
	}
	return nil
}
