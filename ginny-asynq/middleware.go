package asyncq

import (
	"context"
	"fmt"
	"time"

	"github.com/goriller/ginny/logger"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

// loggingMiddleware 记录任务日志中间件
func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		log := logger.WithContext(ctx)
		start := time.Now()
		rw := t.ResultWriter()
		log.Info("Start processing ", zap.String("TaskID", rw.TaskID()))
		err := h.ProcessTask(ctx, t)
		if err != nil {
			log.Info("Faild processing ", zap.String("TaskID", rw.TaskID()), zap.Error(err))
			return err
		}
		log.Info(fmt.Sprintf("Finished processing %q: Elapsed Time = %v", rw.TaskID(), time.Since(start)))
		return nil
	})
}
