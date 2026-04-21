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

func userSentinel(userID uuid.UUID) string {
	return "user_invalidated:" + userID.String()
}

func (r *TokenRepository) InvalidateAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	query := `
		INSERT INTO invalidated_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, NOW() + INTERVAL '365 days')
		ON CONFLICT (token_hash) DO NOTHING
	`
	_, err := r.pool.Exec(ctx, query, userSentinel(userID), userID)
	return err
}

func (r *TokenRepository) IsUserInvalidated(ctx context.Context, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM invalidated_tokens WHERE token_hash = $1 AND expires_at > NOW())`
	var exists bool
	err := r.pool.QueryRow(ctx, query, userSentinel(userID)).Scan(&exists)
	return exists, err
}
