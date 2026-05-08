package cache

import (
	"github.com/google/wire"
	redis "github.com/goriller/ginny-redis"
)

var ProviderSet = wire.NewSet(
	redis.Provider,
	RedisCacheProvider,
)
