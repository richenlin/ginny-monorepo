package grpc

import (
	"context"
	"fmt"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	_ "github.com/mbobakov/grpc-consul-resolver" // It's important
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

// ClientOptions
type ClientOptions struct {
	Target          string // "consul://xxx" or "http://xxx"
	Wait            time.Duration
	Tag             string
	GrpcDialOptions []grpc.DialOption
}

// NewClientOptions
func NewClientOptions(v *viper.Viper) (*ClientOptions, error) {
	var (
		err error
		o   = new(ClientOptions)
	)
	if err = v.UnmarshalKey("grpc.client", o); err != nil {
		return nil, err
	}
	return o, nil
}

// ClientOptional
type ClientOptional func(o *ClientOptions)

// WithTarget
func WithTarget(t string) ClientOptional {
	return func(o *ClientOptions) {
		o.Target = t
	}
}

// WithTimeout
func WithTimeout(d time.Duration) ClientOptional {
	return func(o *ClientOptions) {
		o.Wait = d
	}
}

// WithTag
func WithTag(tag string) ClientOptional {
	return func(o *ClientOptions) {
		o.Tag = tag
	}
}

// WithGrpcDialOptions
func WithGrpcDialOptions(options ...grpc.DialOption) ClientOptional {
	return func(o *ClientOptions) {
		o.GrpcDialOptions = append(o.GrpcDialOptions, options...)
	}
}

// Client
type Client struct {
	options *ClientOptions
}

// NewClient
func NewClient(o *ClientOptions, tracer opentracing.Tracer) (*Client, error) {
	grpc_prometheus.EnableClientHandlingTimeHistogram()

	o.GrpcDialOptions = append(o.GrpcDialOptions,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_prometheus.UnaryClientInterceptor,
			otgrpc.OpenTracingClientInterceptor(tracer)),
		),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			grpc_prometheus.StreamClientInterceptor,
			otgrpc.OpenTracingStreamClientInterceptor(tracer)),
		),
	)
	return &Client{
		options: o,
	}, nil
}

// Dial
func (c *Client) Dial(service string, options ...ClientOptional) (*grpc.ClientConn, error) {
	o := &ClientOptions{
		Target:          c.options.Target,
		Wait:            c.options.Wait,
		Tag:             c.options.Tag,
		GrpcDialOptions: c.options.GrpcDialOptions,
	}
	for _, option := range options {
		option(o)
	}
	ctx, cancel := context.WithTimeout(context.Background(), o.Wait)
	defer cancel()

	target := fmt.Sprintf("%s/%s?wait=%s&tag=%s", o.Target, service, o.Wait, o.Tag)
	conn, err := grpc.DialContext(ctx, target, o.GrpcDialOptions...)
	if err != nil {
		return nil, errors.Wrap(err, "grpc dial error")
	}

	return conn, nil
}
