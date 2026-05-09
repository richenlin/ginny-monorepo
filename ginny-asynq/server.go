package asyncq

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
)

// Server wraps the asynq server for processing tasks.
type Server struct {
	server     *asynq.Server
	Dispatcher *taskDispatcher
	logger     *slog.Logger
}

func newServer(ctx context.Context, opt *Config) (*Server, error) {
	var redisConnOpt asynq.RedisConnOpt
	if opt.redisClusterClientOpt != nil {
		redisConnOpt = opt.redisClusterClientOpt
	} else if opt.redisFailoverClientOpt != nil {
		redisConnOpt = opt.redisFailoverClientOpt
	} else {
		redisConnOpt = opt.redisClientOpt
	}
	server := asynq.NewServer(
		redisConnOpt,
		asynq.Config{
			Concurrency:  10,
			Logger:       opt.Logger,
			ErrorHandler: asynq.ErrorHandlerFunc(HandleErrorFunc),
			Queues: map[string]int{
				QueueCritical: 5,
				QueueDefault:  2,
				QueueLow:      1,
			},
		},
	)

	d, ok := redisConnOpt.MakeRedisClient().(redis.UniversalClient)
	if !ok {
		return nil, fmt.Errorf("invalid RedisConnOpt type %T", redisConnOpt)
	}

	dispatcher := newDispatcher(d)

	return &Server{
		server:     server,
		Dispatcher: dispatcher,
		logger:     opt.Logger,
	}, nil
}

// Run starts the asynq server. This blocks until the server is stopped.
func (s *Server) Run() error {
	mux := asynq.NewServeMux()
	mux.Use(loggingMiddleware)
	mux.HandleFunc(dispatcherName, s.Dispatcher.ProcessTask)
	if err := s.server.Run(mux); err != nil {
		return fmt.Errorf("could not run server: %v", err)
	}
	return nil
}

// Shutdown gracefully stops the asynq server.
func (s *Server) Shutdown() {
	s.server.Shutdown()
}

// HandleErrorFunc is the default error handler for asynq tasks.
func HandleErrorFunc(ctx context.Context, task *asynq.Task, err error) {
	slog.ErrorContext(ctx, "TaskServer handler error",
		slog.String("error", err.Error()),
		slog.String("task_type", task.Type()),
	)
}
