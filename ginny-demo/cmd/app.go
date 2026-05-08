//go:build wireinject
// +build wireinject

package main

import (
	"context"

	"github.com/goriller/ginny-demo/internal/config"
	"github.com/goriller/ginny-demo/internal/logic"
	"github.com/goriller/ginny-demo/internal/repo"
	"github.com/goriller/ginny-demo/internal/service"

	"github.com/google/wire"
	"github.com/goriller/ginny"
	consul "github.com/goriller/ginny-consul"
	jaeger "github.com/goriller/ginny-jaeger"
	"github.com/goriller/ginny/server"
	"github.com/opentracing/opentracing-go"
)

// NewApp
func NewApp(ctx context.Context) (*ginny.Application, error) {
	panic(wire.Build(wire.NewSet(
		consul.ProviderSet,
		jaeger.ProviderSet,
		config.ProviderSet,
		repo.ProviderSet,
		logic.ProviderSet,
		service.ProviderSet,
		serverOption,
		ginny.AppProviderSet,
	)))
}

func serverOption(
	config *config.Config,
	consul *consul.Client,
	tracer opentracing.Tracer,
) (opts []server.Option) {
	opts = append(opts, server.WithDiscover(consul, config.ServiceTags...))
	opts = append(opts, server.WithTracer(tracer))
	return
}
