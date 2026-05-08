package repo

import (
	"MODULE_NAME/internal/repo/entity"

	"github.com/google/wire"
	gorm "github.com/goriller/ginny-gorm"
	// DATABASE_LIB 锚点请勿删除! Do not delete this line!
)

// Query
type Query struct {
	QueryStr string
	Attrs    []interface{}
}

var ProviderSet = wire.NewSet(
	gorm.Provider,
	// mongo.Provider,
	// DATABASE_PROVIDER 锚点请勿删除! Do not delete this line!
	entity.NewUserEntity,
	UserRepoProvider,
	// REPO_PROVIDER 锚点请勿删除! Do not delete this line!
)
