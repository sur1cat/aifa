package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Blacklist stores revoked JWT IDs (jti) in Redis with TTL equal to the
// original token's remaining lifetime. Every authenticated request checks
// the blacklist — lookup is O(1) and ~1ms inside the same cluster.
type Blacklist struct {
	rdb *redis.Client
}

func NewBlacklist(rdb *redis.Client) *Blacklist {
	return &Blacklist{rdb: rdb}
}

const keyPrefix = "auth:blacklist:"

func (b *Blacklist) Revoke(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	return b.rdb.Set(ctx, keyPrefix+jti, "1", ttl).Err()
}

func (b *Blacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := b.rdb.Exists(ctx, keyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
