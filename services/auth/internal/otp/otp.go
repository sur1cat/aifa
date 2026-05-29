package otp

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "auth:otp:"
	codeTTL   = 5 * time.Minute
	codeLen   = 6
	maxVerify = 5
)

type Store struct {
	rdb *redis.Client
}

func NewStore(rdb *redis.Client) *Store {
	return &Store{rdb: rdb}
}

func (s *Store) Generate(ctx context.Context, phone string) (string, time.Time, error) {
	code, err := generateCode()
	if err != nil {
		return "", time.Time{}, err
	}

	key := keyPrefix + phone
	attemptsKey := key + ":attempts"

	pipe := s.rdb.Pipeline()
	pipe.Set(ctx, key, code, codeTTL)
	pipe.Set(ctx, attemptsKey, 0, codeTTL)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", time.Time{}, fmt.Errorf("store otp: %w", err)
	}

	return code, time.Now().Add(codeTTL), nil
}

func (s *Store) Verify(ctx context.Context, phone, code string) (bool, error) {
	key := keyPrefix + phone
	attemptsKey := key + ":attempts"

	attempts, err := s.rdb.Incr(ctx, attemptsKey).Result()
	if err != nil {
		return false, fmt.Errorf("incr attempts: %w", err)
	}
	if attempts > maxVerify {
		s.rdb.Del(ctx, key, attemptsKey)
		return false, nil
	}

	stored, err := s.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("get otp: %w", err)
	}

	if stored != code {
		return false, nil
	}

	s.rdb.Del(ctx, key, attemptsKey)
	return true, nil
}

func generateCode() (string, error) {
	max := new(big.Int).SetInt64(1_000_000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
