package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/google/wire"
	"github.com/goriller/ginny-demo/internal/repo/entity"
	"github.com/goriller/ginny-util/validation"
	"github.com/goriller/gorm-plus/gplus"
	"gorm.io/gorm"
	// DATABASE_LIB 锚点请勿删除! Do not delete this line!
)

// UserRepoProvider
var UserRepoProvider = wire.NewSet(NewUserRepo,
	wire.Bind(new(IUserRepo), new(*UserRepo)))

// IUserRepo
type IUserRepo interface {
	Count(ctx context.Context, where entity.UserEntity) (total int64, err error)
	Find(ctx context.Context, where entity.UserEntity, order []string) (
		result *entity.UserEntity, err error)
	FindAll(ctx context.Context, where entity.UserEntity,
		order []string, opt ...int) (result []entity.UserEntity, err error)
	Insert(ctx context.Context,
		entity *entity.UserEntity) (int64, error)
	Update(ctx context.Context, where entity.UserEntity,
		update entity.UserEntity) (int64, error)
	Delete(ctx context.Context,
		where entity.UserEntity) (int64, error)
	PDelete(ctx context.Context,
		where entity.UserEntity) (int64, error)
	SelectPage(ctx context.Context, query *gplus.QueryCond[entity.UserEntity],
		limit, offset int) (*gplus.Page[entity.UserEntity], error)
}

// UserRepo
type UserRepo struct {
	orm *gorm.DB
	// mongo *mongo.Manager
	entity *entity.UserEntity

	gplus.Dao[entity.UserEntity]
	// STRUCT_ATTR 锚点请勿删除! Do not delete this line!
}

// NewUserRepo
func NewUserRepo(
	// redis *redis.Manager,
	orm *gorm.DB,
	// mongo *mongo.Manager,
	// FUNC_PARAM 锚点请勿删除! Do not delete this line!
) (*UserRepo, error) {
	return &UserRepo{
		orm:    orm,
		entity: &entity.UserEntity{},
		// FUNC_ATTR 锚点请勿删除! Do not delete this line!
	}, nil
}

// Count
func (p *UserRepo) Count(ctx context.Context, where entity.UserEntity) (total int64, err error) {
	err = p.orm.Table(p.entity.TableName()).Where(where).Count(&total).Error
	return
}

// Find
/** 注意:
 * where条件仅可以使用非零值
 * order的字段名需要同数据库字段名一致
 */
func (p *UserRepo) Find(ctx context.Context, where entity.UserEntity, order []string) (
	result *entity.UserEntity, err error) {
	if order == nil {
		order = []string{"id desc"}
	}
	err = p.orm.Table(p.entity.TableName()).Where(where).Order(strings.Join(order, ",")).First(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return
}

// FindAll
/** 注意:
 * where条件仅可以使用非零值
 * order的字段名需要同数据库字段名一致
 */
func (p *UserRepo) FindAll(ctx context.Context, where entity.UserEntity,
	order []string, opt ...int) (result []entity.UserEntity, err error) {
	if order == nil {
		order = []string{"id desc"}
	}

	db := p.orm.Table(p.entity.TableName()).Where(where).Order(strings.Join(order, ","))
	var (
		limit  = 1000
		offset = 0
	)
	if len(opt) > 0 {
		limit = opt[0]
	}
	db = db.Limit(limit)
	if len(opt) == 2 {
		offset = opt[1]
	}
	db = db.Offset(offset)
	err = db.Find(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return
}

func (p *UserRepo) SelectPage(ctx context.Context, query *gplus.QueryCond[entity.UserEntity],
	offset, limit int) (*gplus.Page[entity.UserEntity], error) {
	page := gplus.NewPage[entity.UserEntity](offset, limit)
	page, resultDb := gplus.SelectPage(page, query)
	if resultDb.Error != nil {
		if errors.Is(resultDb.Error, gorm.ErrRecordNotFound) {
			return page, nil
		}
		return nil, resultDb.Error
	}
	return page, nil
}

// Insert
func (p *UserRepo) Insert(ctx context.Context,
	entity *entity.UserEntity) (int64, error) {
	if err := validation.Validate(entity); err != nil {
		return 0, err
	}
	result := p.orm.Table(p.entity.TableName()).Create(entity)
	return entity.Id, result.Error
}

// Update
/** 注意:
 * where、update仅可以使用非零值
 */
func (p *UserRepo) Update(ctx context.Context, where entity.UserEntity,
	update entity.UserEntity) (int64, error) {
	if err := validation.Validate(update); err != nil {
		return 0, err
	}
	result := p.orm.Table(p.entity.TableName()).Where(where).Updates(update)
	return result.RowsAffected, result.Error
}

// Delete
/** 注意:
 * where条件仅可以使用非零值
 */
func (p *UserRepo) Delete(ctx context.Context,
	where entity.UserEntity) (int64, error) {
	var t *entity.UserEntity
	result := p.orm.Table(p.entity.TableName()).Where(where).Delete(&t)
	return result.RowsAffected, result.Error
}

// PDelete physical deletion
/** 注意:
 * where条件仅可以使用非零值
 */
func (p *UserRepo) PDelete(ctx context.Context,
	where entity.UserEntity) (int64, error) {
	var t *entity.UserEntity
	result := p.orm.Table(p.entity.TableName()).Unscoped().Where(where).Delete(&t)
	return result.RowsAffected, result.Error
}
