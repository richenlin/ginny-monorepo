package orm

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Config
type Config struct {
	Type string `json:"type" mapstructure:"type"`
	// mysql://localhost:3306/dbname[?username=value1&password=value2&paramN=valueN]
	Master   string   `json:"master" mapstructure:"master"`
	Replicas []string `json:"replicas" mapstructure:"replicas"`
	gorm.Config
	dialector gorm.Dialector
}

// NewConfig
func NewConfig(v *viper.Viper, logger *zap.Logger) (*Config, error) {
	var err error
	o := new(Config)
	if err = v.UnmarshalKey("gorm", o); err != nil {
		return nil, errors.Wrap(err, "unmarshal gorm option error")
	}
	o.QueryFields = true
	o.Logger = newLogger(logger)

	if o.Replicas == nil || len(o.Replicas) == 0 {
		o.Replicas = []string{
			o.Master,
		}
	}

	// parse dsn
	o.Master, err = o.parseUrl(o.Master)
	if err != nil {
		return nil, errors.Wrap(err, "parse dsn error")
	}

	rdb := []string{}
	for _, v := range o.Replicas {
		u, err := o.parseUrl(v)
		if err != nil {
			return nil, errors.Wrap(err, "parse dsn error")
		}
		rdb = append(rdb, u)
	}
	o.Replicas = rdb

	return o, nil
}

func (c *Config) parseUrl(dsn string) (string, error) {
	if dsn == "" {
		return "", fmt.Errorf("undefined wdb dsn")
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	q := u.Query()
	wdbUser := q.Get("username")
	wdbPass := q.Get("password")

	q.Del("username")
	q.Del("password")

	if c.Type == "" {
		c.Type = u.Scheme
	}

	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?%s", wdbUser, wdbPass, u.Host, strings.TrimPrefix(u.Path, "/"), q.Encode()), nil
}
