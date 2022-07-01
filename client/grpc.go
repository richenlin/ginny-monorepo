package client

import (
	"context"
	"fmt"
	"time"

	"github.com/google/wire"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/tracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	consulApi "github.com/hashicorp/consul/api"
	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

var GrpcClientProvider = wire.NewSet(NewGrpcClientOptions, NewGrpcClient)

// GrpcClientOptions
type GrpcClientOptions struct {
	Target          string // "consul://xxx" or ip+port "xx.xx.xx.xx:xxxx"
	Wait            time.Duration
	Tag             string
	LoadBalance     string
	tracer          opentracing.Tracer
	grpcDialOptions []grpc.DialOption
	consulOptions   *consulApi.Config
}

// NewGrpcClientOptions
func NewGrpcClientOptions(v *viper.Viper) (*GrpcClientOptions, error) {
	var (
		err error
		o   = new(GrpcClientOptions)
	)
	if err = v.UnmarshalKey("grpc.client", o); err != nil {
		return nil, err
	}
	return o, nil
}

// GrpcClientOptional
type GrpcClientOptional func(o *GrpcClientOptions)

// WithTarget
// "consul://xxx" or ip+port "xx.xx.xx.xx:xxxx"
func WithTarget(t string) GrpcClientOptional {
	return func(o *GrpcClientOptions) {
		o.Target = t
	}
}

// WithTimeout
func WithTimeout(d time.Duration) GrpcClientOptional {
	return func(o *GrpcClientOptions) {
		o.Wait = d
	}
}

// WithTag
func WithTag(tag string) GrpcClientOptional {
	return func(o *GrpcClientOptions) {
		o.Tag = tag
	}
}

// WithGrpcDialOptions
func WithGrpcDialOptions(options ...grpc.DialOption) GrpcClientOptional {
	return func(o *GrpcClientOptions) {
		o.grpcDialOptions = append(o.grpcDialOptions, options...)
	}
}

// WithConsulConf
func WithConsulConf(consul *consulApi.Config) GrpcClientOptional {
	return func(o *GrpcClientOptions) {
		o.consulOptions = consul
	}
}

// WithTracer
func WithTracer(tracer opentracing.Tracer) GrpcClientOptional {
	return func(o *GrpcClientOptions) {
		o.tracer = tracer
	}
}

// GrpcClient
type GrpcClient struct {
	options *GrpcClientOptions
}

// NewGrpcClient
func NewGrpcClient(o *GrpcClientOptions, tracer opentracing.Tracer) (*GrpcClient, error) {
	o.tracer = tracer
	if o.LoadBalance == "" {
		o.LoadBalance = roundrobin.Name
	}
	loadBalanceConfig := fmt.Sprintf(`{"LoadBalancingPolicy":"%s"}`, o.LoadBalance)
	grpc_prometheus.EnableClientHandlingTimeHistogram()
	// retry
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
		grpc_retry.WithCodes(codes.NotFound, codes.Aborted),
	}
	o.grpcDialOptions = append(o.grpcDialOptions,
		grpc.WithDefaultServiceConfig(loadBalanceConfig),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpc_prometheus.UnaryClientInterceptor,
			grpc_retry.UnaryClientInterceptor(retryOpts...),
		),
		grpc.WithChainStreamInterceptor(
			grpc_prometheus.StreamClientInterceptor,
			grpc_retry.StreamClientInterceptor(retryOpts...),
		),
	)

	if o.tracer != nil {
		o.grpcDialOptions = append(o.grpcDialOptions,
			grpc.WithChainUnaryInterceptor(
				tracing.UnaryClientInterceptor(tracing.WithTracer(o.tracer)),
			),
			grpc.WithChainStreamInterceptor(
				tracing.StreamClientInterceptor(tracing.WithTracer(o.tracer)),
			),
		)
	} else {
		o.grpcDialOptions = append(o.grpcDialOptions,
			grpc.WithChainUnaryInterceptor(
				tracing.UnaryClientInterceptor(),
			),
			grpc.WithChainStreamInterceptor(
				tracing.StreamClientInterceptor(),
			),
		)
	}

	return &GrpcClient{
		options: o,
	}, nil
}

// Dial
func (c *GrpcClient) Dial(service string, options ...GrpcClientOptional) (*grpc.ClientConn, error) {
	o := &GrpcClientOptions{
		Target:          c.options.Target,
		Wait:            c.options.Wait,
		Tag:             c.options.Tag,
		grpcDialOptions: c.options.grpcDialOptions,
		consulOptions:   c.options.consulOptions,
	}

	for _, option := range options {
		option(o)
	}
	if o.Wait == 0 {
		o.Wait = time.Second * 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), o.Wait)
	defer cancel()

	if o.consulOptions != nil && o.consulOptions.Address != "" {
		o.Target = fmt.Sprintf("consul://%s/%s?wait=%s&tag=%s", o.consulOptions.Address, service, o.Wait, o.Tag)
	}

	conn, err := grpc.DialContext(ctx, o.Target, o.grpcDialOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "grpc dial error")
	}

	return conn, nil
}
