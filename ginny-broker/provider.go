package broker

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/wire"
	"go.uber.org/zap"
)

// BrokerProvider
var BrokerProvider = wire.NewSet(NewConfiguration, NewBroker)

// NewBroker
func NewBroker(ctx context.Context, log *zap.Logger, conf *Config) (*Broker, error) {
	if conf.Dsn == "" {
		return nil, errors.New("broker dsn is empty")
	}
	uri, errParse := url.Parse(conf.Dsn)
	if errParse != nil {
		return nil, fmt.Errorf("dependency %s parse uri %s error for %w", "broker", conf.Dsn, errParse)
	}
	broker := &Broker{}
	err := broker.Init(ctx, log, uri)
	if err != nil {
		return nil, err
	}
	return broker, nil
}
