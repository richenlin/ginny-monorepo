package entity

import "gorm.io/gorm"

// UserEntity
type UserEntity struct {
	Id        int64  `json:"id" bson:"_id"`
	Name      string `json:"name" bson:"name"`
	Status    int    `json:"status" bson:"status"`
	DeletedAt gorm.DeletedAt
}

// NewUserEntity
func NewUserEntity() (*UserEntity, error) {
	return &UserEntity{}, nil
}

// TableName
func (p *UserEntity) TableName() string {
	return "user"
}

// Validate
func (p *UserEntity) Validate() error {
	return nil
}
