package cache

import (
	"github.com/google/wire"
	redis "github.com/goriller/ginny-redis"
)

var CacheProvider = wire.NewSet(redis.Provider, RedisCacheProvider)
