package main

import (
	"context"
	"log/slog"

	"github.com/goriller/ginny/v2"
	"github.com/goriller/ginny/v2/config"
	"github.com/goriller/ginny/v2/interceptor/logging"
	"github.com/goriller/ginny/v2/interceptor/recovery"
	"github.com/goriller/ginny/v2/log"
	"github.com/goriller/ginny/v2/server"
)

func main() {
	cfg := config.MustLoad()
	logger := log.New(log.WithLevel(slog.LevelInfo), log.WithSource(true))

	srv := server.New(
		server.WithAppAddr(cfg.Server.Addr),
		server.WithAdminAddr(cfg.Admin.Addr),
		server.WithLogger(logger),
		server.WithInterceptor(recovery.NewInterceptor(logger)),
		server.WithInterceptor(logging.NewInterceptor(logger)),
		server.WithReflection(true),
		server.WithDebug(true),
	)

	// Register your services here:
	// path, handler := myappv1connect.NewMyServiceHandler(&myServiceImpl{})
	// srv.RegisterService(path, handler)

	app, _ := ginny.New(
		ginny.WithName(cfg.App.Name),
		ginny.WithConfig(cfg),
		ginny.WithLogger(logger),
		ginny.WithServer(srv),
	)
	app.Start(context.Background())
}
