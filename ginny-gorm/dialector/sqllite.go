package dialector

import (
	"net/url"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func NewSqllite(u *url.URL) gorm.Dialector {
	return sqlite.Open(u.String())
}
