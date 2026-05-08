package jaeger

import (
	"context"

	"github.com/google/wire"
	"github.com/goriller/ginny-util/graceful"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics"
)

// ProviderSet
var (
	ProviderSet           = wire.NewSet(NewConfiguration, NewJaegerTracer)
	ProviderSetWithOption = wire.NewSet(NewConfiguration, NewJaegerTracerWithOption)
)

// NewConfiguration
func NewConfiguration(v *viper.Viper) (*config.Configuration, error) {
	var (
		err error
		c   = new(config.Configuration)
	)

	if err = v.UnmarshalKey("jaeger", c); err != nil {
		return nil, errors.Wrap(err, "unmarshal jaeger configuration error")
	}

	return c, nil
}

// NewJaegerTracer NewJaegerTracer for current service
func NewJaegerTracer(ctx context.Context, cfg *config.Configuration) (opentracing.Tracer, error) {
	return NewJaegerTracerWithOption(ctx, cfg)
}

// NewJaegerTracerWithOption
func NewJaegerTracerWithOption(ctx context.Context, cfg *config.Configuration,
	opt ...config.Option) (opentracing.Tracer, error) {
	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := &stdLog{}
	jMetricsFactory := metrics.NullFactory

	opts := []config.Option{
		config.Logger(jLogger),
		config.Metrics(jMetricsFactory),
	}
	opts = append(opts, opt...)

	// Initialize tracer with a logger and a metrics factory
	tracer, closer, err := cfg.NewTracer(opts...)
	if err != nil {
		return nil, errors.Wrap(err, "create jaeger tracer error")
	}

	opentracing.SetGlobalTracer(tracer)
	graceful.AddCloser(func(ctx context.Context) error {
		return closer.Close()
	})
	return tracer, nil
}
