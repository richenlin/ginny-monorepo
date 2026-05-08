package entity

import (
	"time"

	"gorm.io/gorm"
)

// UserEntity
type UserEntity struct {
	Id        int64  `json:"id" bson:"_id"`
	Name      string `json:"name" bson:"name"`
	Deleted   gorm.DeletedAt
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TableName
func (p *UserEntity) TableName() string {
	return "lc_user"
}

// Validate
func (p *UserEntity) Validate() error {
	return nil
}
