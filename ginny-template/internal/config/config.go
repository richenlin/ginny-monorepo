package config

import (
	"sync/atomic"

	"github.com/google/wire"
	"github.com/spf13/viper"
)

var (
	// 为解决并发读写问题，使用atomic存储操作配置类的指针
	appConfig   atomic.Value
	ProviderSet = wire.NewSet(NewConfig)
)

// Config
type Config struct {
	Test string
}

// NewConfig
func NewConfig(v *viper.Viper) (*Config, error) {
	o := &Config{}
	if err := v.Unmarshal(o); err != nil {
		return nil, err
	}
	appConfig.Store(o)
	return o, nil
}

// Get 获取配置
func Get() *Config {
	res := appConfig.Load()
	if res == nil {
		panic("config is nil")
	}
	if data, ok := res.(*Config); ok {
		return data
	}
	return &Config{}
}
