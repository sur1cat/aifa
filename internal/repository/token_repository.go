package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TokenRepository struct {
	pool *pgxpool.Pool
}

func NewTokenRepository(pool *pgxpool.Pool) *TokenRepository {
	return &TokenRepository{pool: pool}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (r *TokenRepository) Invalidate(ctx context.Context, token string, userID uuid.UUID, expiresAt time.Time) error {
	query := `
		INSERT INTO invalidated_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (token_hash) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, hashToken(token), userID, expiresAt)
	return err
}

func (r *TokenRepository) IsInvalidated(ctx context.Context, token string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM invalidated_tokens WHERE token_hash = $1)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, hashToken(token)).Scan(&exists)
	return exists, err
}

func (r *TokenRepository) CleanupExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM invalidated_tokens WHERE expires_at < NOW()`)
	return err
}

func (r *TokenRepository) InvalidateAllUserTokens(ctx context.Context, userID uuid.UUID) error {

	return nil
}
