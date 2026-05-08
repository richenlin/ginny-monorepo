package dialector

import (
	"net/url"

	"gorm.io/driver/clickhouse"
	"gorm.io/gorm"
)

func NewClickhouse(u *url.URL) gorm.Dialector {
	return clickhouse.Open(u.String())
}
