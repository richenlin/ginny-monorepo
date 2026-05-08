package repo

import (
	"github.com/google/wire"
	"github.com/goriller/ginny-demo/internal/cache"
	gorm "github.com/goriller/ginny-gorm"
	// DATABASE_LIB 锚点请勿删除! Do not delete this line!
)

var ProviderSet = wire.NewSet(
	cache.CacheProvider,
	gorm.Provider,
	// mongo.Provider,
	// DATABASE_PROVIDER 锚点请勿删除! Do not delete this line!
	UserRepoProvider,
	// REPO_PROVIDER 锚点请勿删除! Do not delete this line!
)
