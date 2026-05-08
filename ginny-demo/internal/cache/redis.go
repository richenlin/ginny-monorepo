package cache

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	gredis "github.com/goriller/ginny-redis"
)

// RedisCacheProvider
var RedisCacheProvider = wire.NewSet(NewRedisCache,
	wire.Bind(new(IRedisCache), new(*RedisCache)))

// IRedisCache
type IRedisCache interface {
	Ping(ctx context.Context) (string, error)
}

// RedisCache
type RedisCache struct {
	redis redis.UniversalClient
}

// NewRedisCache
func NewRedisCache(
	redis *gredis.Redis,
) (*RedisCache, error) {
	return &RedisCache{
		redis: redis.Client(),
	}, nil
}

// Ping
func (p *RedisCache) Ping(ctx context.Context) (string, error) {
	return p.redis.Ping(ctx).Result()
}
