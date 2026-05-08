package dialector

import (
	"net/url"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewMysql(u *url.URL) gorm.Dialector {
	return mysql.Open(u.String())
}
