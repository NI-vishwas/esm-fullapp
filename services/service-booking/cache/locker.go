package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrLockAcquisitionFailed = errors.New("could not acquire lock, resource is busy")

type RedisLocker struct {
	client *redis.Client
}

func NewRedisLocker(client *redis.Client) *RedisLocker {
	return &RedisLocker{client: client}
}

func (l *RedisLocker) Lock(ctx context.Context, resourceKey string, ttl time.Duration) (string, error) {
	token := uuid.New().String()
	lockKey := fmt.Sprintf("lock:%s", resourceKey)

	success, err := l.client.SetNX(ctx, lockKey, token, ttl).Result()
	if err != nil {
		return "", fmt.Errorf("redis error during lock: %w", err)
	}

	if !success {
		return "", ErrLockAcquisitionFailed
	}

	return token, nil
}

func (l *RedisLocker) Unlock(ctx context.Context, resourceKey string, token string) error {
	lockKey := fmt.Sprintf("lock:%s", resourceKey)

	var luaReleaseScript = `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, luaReleaseScript, []string{lockKey}, token).Int64()
	if err != nil {
		return fmt.Errorf("redis error during unlock: %w", err)
	}

	if result == 0 {
		return errors.New("lock was not released: token mismatch or lock expired")
	}

	return nil
}