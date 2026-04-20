package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"habitflow/internal/domain"
)

type DeviceTokenRepository struct {
	pool *pgxpool.Pool
}

func NewDeviceTokenRepository(pool *pgxpool.Pool) *DeviceTokenRepository {
	return &DeviceTokenRepository{pool: pool}
}

func (r *DeviceTokenRepository) Register(ctx context.Context, userID uuid.UUID, token, platform string) error {
	query := `
		INSERT INTO device_tokens (user_id, token, platform)
		VALUES ($1, $2, $3)
		ON CONFLICT (token) DO UPDATE SET
			user_id = $1,
			platform = $3,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query, userID, token, platform)
	return err
}

func (r *DeviceTokenRepository) Unregister(ctx context.Context, token string) error {
	query := `DELETE FROM device_tokens WHERE token = $1`
	_, err := r.pool.Exec(ctx, query, token)
	return err
}

func (r *DeviceTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.DeviceToken, error) {
	query := `SELECT id, user_id, token, platform, created_at, updated_at FROM device_tokens WHERE user_id = $1`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []domain.DeviceToken
	for rows.Next() {
		var t domain.DeviceToken
		err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}

	if tokens == nil {
		tokens = []domain.DeviceToken{}
	}
	return tokens, rows.Err()
}

func (r *DeviceTokenRepository) GetAllTokens(ctx context.Context) ([]domain.DeviceToken, error) {
	query := `SELECT id, user_id, token, platform, created_at, updated_at FROM device_tokens`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		if err == pgx.ErrNoRows {
			return []domain.DeviceToken{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var tokens []domain.DeviceToken
	for rows.Next() {
		var t domain.DeviceToken
		err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.Platform, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}

	if tokens == nil {
		tokens = []domain.DeviceToken{}
	}
	return tokens, rows.Err()
}
