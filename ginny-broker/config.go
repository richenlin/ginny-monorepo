package broker

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Dsn string
}

// NewConfiguration
func NewConfiguration(v *viper.Viper) (*Config, error) {
	var (
		err error
		c   = new(Config)
	)

	if err = v.UnmarshalKey("broker", c); err != nil {
		return nil, errors.Wrap(err, "unmarshal broker configuration error")
	}

	return c, nil
}
