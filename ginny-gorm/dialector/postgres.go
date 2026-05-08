package dialector

import (
	"net/url"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewPostgres(u *url.URL) gorm.Dialector {
	return postgres.Open(u.String())
}
