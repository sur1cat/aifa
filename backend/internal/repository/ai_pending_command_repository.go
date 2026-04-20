package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AIPendingCommandRepository struct {
	pool *pgxpool.Pool
}

func NewAIPendingCommandRepository(pool *pgxpool.Pool) *AIPendingCommandRepository {
	return &AIPendingCommandRepository{pool: pool}
}

func (r *AIPendingCommandRepository) GetPayloadByUserID(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	query := `SELECT payload FROM ai_pending_commands WHERE user_id = $1`

	var payload []byte
	err := r.pool.QueryRow(ctx, query, userID).Scan(&payload)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return payload, nil
}

func (r *AIPendingCommandRepository) UpsertPayload(ctx context.Context, userID uuid.UUID, payload []byte) error {
	query := `
		INSERT INTO ai_pending_commands (user_id, payload, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET
			payload = EXCLUDED.payload,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, query, userID, payload)
	return err
}

func (r *AIPendingCommandRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM ai_pending_commands WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}
