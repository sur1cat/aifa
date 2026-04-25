package repository

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// Blacklist checks whether a JWT ID (jti) has been revoked by auth-service.
// The key is owned by auth-service (auth:blacklist:<jti>); every other
// service is a read-only consumer.
type Blacklist struct {
	rdb *redis.Client
}

func NewBlacklist(rdb *redis.Client) *Blacklist {
	return &Blacklist{rdb: rdb}
}

const keyPrefix = "auth:blacklist:"

func (b *Blacklist) IsRevoked(ctx context.Context, jti string) (bool, error) {
	n, err := b.rdb.Exists(ctx, keyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
