package asyncq

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/goriller/ginny-util/graceful"
	"github.com/goriller/ginny/logger"
	"github.com/hibiken/asynq"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Server struct {
	server     *asynq.Server
	Dispatcher *taskDispatcher
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
			// Specify how many concurrent workers to use
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

	graceful.AddCloser(func(ctx context.Context) error {
		server.Shutdown()
		return nil
	})

	return &Server{
		server:     server,
		Dispatcher: dispatcher,
	}, nil
}

func (s *Server) Start() error {
	// mux maps a type to a handler
	mux := asynq.NewServeMux()
	mux.Use(loggingMiddleware)
	mux.HandleFunc(dispatcherName, s.Dispatcher.ProcessTask)
	if err := s.server.Run(mux); err != nil {
		return fmt.Errorf("could not run server: %v", err)
	}
	return nil
}

func HandleErrorFunc(ctx context.Context, task *asynq.Task, err error) {
	log := logger.WithContext(ctx)
	log.Error("TaskServer handler error", zap.Error(errors.WithStack(err)))
	// report error
	// ReportService.Notify(err)
}
