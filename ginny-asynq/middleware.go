package asyncq

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

// loggingMiddleware logs task processing.
func loggingMiddleware(h asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		start := time.Now()
		rw := t.ResultWriter()
		slog.InfoContext(ctx, "Start processing",
			slog.String("TaskID", rw.TaskID()),
		)
		err := h.ProcessTask(ctx, t)
		if err != nil {
			slog.InfoContext(ctx, "Failed processing",
				slog.String("TaskID", rw.TaskID()),
				slog.String("error", err.Error()),
			)
			return err
		}
		slog.InfoContext(ctx, fmt.Sprintf("Finished processing %q: Elapsed Time = %v",
			rw.TaskID(), time.Since(start)))
		return nil
	})
}
