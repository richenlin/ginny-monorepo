// Package main is the entry point for the Ginny v2 Demo application.
//
// It demonstrates:
//   - Configuration loading via koanf
//   - Structured logging with slog
//   - Dual-port server (App :8080, Admin :8081)
//   - ConnectRPC handler registration
//   - Interceptor chain (recovery + logging)
//   - gRPC Reflection
//   - Lifecycle hooks
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/goriller/ginny/v2"
	"github.com/goriller/ginny/v2/config"
	"github.com/goriller/ginny/v2/interceptor/logging"
	"github.com/goriller/ginny/v2/interceptor/recovery"
	"github.com/goriller/ginny/v2/log"
	"github.com/goriller/ginny/v2/server"

	// Generated ConnectRPC handler
	helloworldv1connect "github.com/goriller/ginny-demo/v2/gen/helloworld/v1/helloworldv1connect"

	// Service implementation
	"github.com/goriller/ginny-demo/v2/internal/service"
)

func main() {
	// Load configuration from config.yaml (or defaults)
	cfg := config.MustLoad()

	// Create a structured JSON logger
	logger := log.New(
		log.WithLevel(slog.LevelInfo),
		log.WithSource(true),
	)

	// Create the app+admin dual-port server with interceptors
	srv := server.New(
		server.WithAppAddr(cfg.Server.Addr),
		server.WithAdminAddr(cfg.Admin.Addr),
		server.WithLogger(logger),
		// Layer 2: Connect interceptors (applied to RPC handlers only)
		server.WithInterceptor(recovery.NewInterceptor(logger)),
		server.WithInterceptor(logging.NewInterceptor(logger,
			logging.WithLevel(slog.LevelInfo),
		)),
		// Enable gRPC reflection for debugging (dev only)
		server.WithReflection(true),
		server.WithReflectionServiceNames(
			helloworldv1connect.GreeterServiceName,
		),
		// Enable pprof on admin port in dev
		server.WithDebug(true),
	)

	// Register ConnectRPC service handler
	greeterSvc := service.NewGreeterService()
	path, handler := helloworldv1connect.NewGreeterServiceHandler(greeterSvc)
	srv.RegisterService(path, handler)

	// Create the application with lifecycle management
	app, err := ginny.New(
		ginny.WithName(cfg.App.Name),
		ginny.WithConfig(cfg),
		ginny.WithLogger(logger),
		ginny.WithServer(srv),
	)
	if err != nil {
		logger.Error("failed to create app", slog.Any("error", err))
		os.Exit(1)
	}

	// Start the application (blocks until SIGINT/SIGTERM)
	ctx := context.Background()
	logger.Info("starting application",
		slog.String("name", app.Name()),
		slog.String("version", cfg.App.Version),
		slog.String("app_addr", cfg.Server.Addr),
		slog.String("admin_addr", cfg.Admin.Addr),
	)
	if err := app.Start(ctx); err != nil {
		logger.Error("application stopped with error", slog.Any("error", err))
		os.Exit(1)
	}
	logger.Info("application stopped gracefully")
}
