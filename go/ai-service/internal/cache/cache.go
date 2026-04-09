package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a tiny key/value interface. Bytes in, bytes out.
// Callers handle their own serialization so the cache stays transport-agnostic.
type Cache interface {
	Get(ctx context.Context, key string) (value []byte, ok bool, err error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
}

// NopCache is a zero-cost no-op. Used when Redis is unavailable or disabled.
type NopCache struct{}

func (NopCache) Get(context.Context, string) ([]byte, bool, error)        { return nil, false, nil }
func (NopCache) Set(context.Context, string, []byte, time.Duration) error { return nil }

// RedisCache is a Redis-backed Cache. All keys are automatically prefixed.
type RedisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(client *redis.Client, prefix string) *RedisCache {
	return &RedisCache{client: client, prefix: prefix}
}

func (c *RedisCache) key(k string) string { return c.prefix + ":" + k }

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	val, err := c.client.Get(ctx, c.key(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, c.key(key), value, ttl).Err()
}
