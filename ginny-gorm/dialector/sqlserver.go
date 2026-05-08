package dialector

import (
	"net/url"

	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

func NewSqlserver(u *url.URL) gorm.Dialector {
	return sqlserver.Open(u.String())
}
