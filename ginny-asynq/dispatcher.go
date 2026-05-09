package asyncq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
)

var (
	dispatcherName = "DefaultTask"
)

type stopJobError struct {
	msg string
}

func (e *stopJobError) Error() string {
	return e.msg
}

type ExeFunc = func(ctx context.Context, param interface{}) (interface{}, error)

// taskDispatcher dispatches tasks to registered handlers.
type taskDispatcher struct {
	mapping sync.Map
	Redis   redis.UniversalClient
}

// newDispatcher creates a new task dispatcher.
func newDispatcher(r redis.UniversalClient) *taskDispatcher {
	return &taskDispatcher{
		mapping: sync.Map{},
		Redis:   r,
	}
}

// SetTask registers a task.
func (d *taskDispatcher) SetTask(taskType string, task *taskInfo) {
	d.mapping.Store(taskType, task)
}

// GetTask retrieves a registered task.
func (d *taskDispatcher) GetTask(taskType string) *taskInfo {
	if val, ok := d.mapping.Load(taskType); ok {
		if data, ok := val.(*taskInfo); ok {
			return data
		}
	}
	return nil
}

// ProcessTask processes a task. Implements asynq.Handler.
func (d *taskDispatcher) ProcessTask(ctx context.Context, task *asynq.Task) error {
	taskId := task.ResultWriter().TaskID()
	arg, err := d.GetPayload(ctx, taskId, task)
	if err != nil {
		slog.ErrorContext(ctx, "GetPayload error", slog.String("error", err.Error()))
		return err
	}

	var (
		res   interface{}
		param interface{}
	)
	param = arg.Arg
	t := d.GetTask(arg.TaskType)
	if t == nil {
		return fmt.Errorf("an asynchronous task must be declared before triggering the execution of the task")
	}

	defer func() {
		if rec := recover(); rec != nil {
			e := fmt.Errorf("panic: %+v", rec)
			slog.ErrorContext(ctx, "panic",
				slog.String("stack", string(debug.Stack())),
			)
			t.OnError(ctx, param, e)
		}
	}()

	// process steps
	for k, v := range t.Step {
		step := k + 1
		if arg.Step > 0 && step <= arg.Step {
			continue
		}
		arg, err = d.GetPayload(ctx, taskId, task)
		if err != nil {
			slog.ErrorContext(ctx, "Step error",
				slog.String("type", task.Type()),
				slog.Int("step", step),
				slog.String("error", err.Error()),
			)
			break
		}
		param = arg.Arg
		res, err = d.RetryCallFunc(ctx, v.Fn, param,
			v.RetryTimes, v.RetryPeriod, v.TimeOut)
		if err != nil {
			slog.ErrorContext(ctx, "Handler error",
				slog.String("type", task.Type()),
				slog.Int("step", step),
				slog.String("error", err.Error()),
			)
			break
		}
		if res != nil {
			arg.Arg = res
		}
		arg.Step = step
		err = d.SetPayload(ctx, taskId, arg)
		if err != nil {
			slog.ErrorContext(ctx, "Step setPayload error",
				slog.String("type", task.Type()),
				slog.Int("step", step),
				slog.String("error", err.Error()),
			)
			break
		}
	}

	if err != nil {
		t.OnError(ctx, param, err)
		return err
	}

	t.OnSuccess(ctx, param)
	return nil
}

// GetPayload retrieves task payload from Redis or task body.
func (d *taskDispatcher) GetPayload(ctx context.Context, taskId string, task *asynq.Task) (*taskArg, error) {
	arg := &taskArg{}
	key := fmt.Sprintf("asynq:{%s}:p:%s", defaultQueue, taskId)
	s, err := d.Redis.Get(ctx, key).Result()
	if err != nil || s == "" {
		s = string(task.Payload())
	}
	err = json.Unmarshal([]byte(s), arg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal task arg failed: %v", err.Error())
	}
	return arg, nil
}

// SetPayload stores task payload in Redis.
func (d *taskDispatcher) SetPayload(ctx context.Context, taskId string, arg *taskArg) error {
	key := fmt.Sprintf("asynq:{%s}:p:%s", defaultQueue, taskId)
	bt, err := json.Marshal(arg)
	if err != nil {
		return err
	}
	return d.Redis.Set(ctx, key, string(bt), 0).Err()
}

// RetryCallFunc calls a function with retry logic.
func (d *taskDispatcher) RetryCallFunc(ctx context.Context, fn ExeFunc,
	params interface{}, retry ...int) (interface{}, error) {
	var (
		err         error
		result      interface{}
		retryTimes  int = 1
		retryPeriod     = 100 * time.Millisecond
		timeout         = time.Millisecond * 600000
	)
	l := len(retry)
	if l == 1 {
		if retry[0] > 0 {
			retryTimes = retry[0]
		}
	} else if l == 2 {
		if retry[0] > 0 {
			retryTimes = retry[0]
		}
		if retry[1] > 0 {
			retryPeriod = time.Duration(retry[1]) * time.Millisecond
		}
	} else if l == 3 {
		if retry[0] > 0 {
			retryTimes = retry[0]
		}
		if retry[1] > 0 {
			retryPeriod = time.Duration(retry[1]) * time.Millisecond
		}
		if retry[2] > 0 {
			timeout = time.Duration(retry[2]) * time.Millisecond
		}
	}

	num := 0
	o := time.After(timeout)
	t := time.NewTicker(retryPeriod)
	defer t.Stop()
	for range t.C {
		select {
		case <-o:
			return nil, fmt.Errorf("call function context deadline exceeded, params: [%+v]", params)
		default:
		}
		num++
		result, err = fn(ctx, params)
		if err != nil {
			var targetErr *stopJobError
			if errors.As(err, &targetErr) {
				return nil, fmt.Errorf("err:[%v]", err.Error())
			}
			if num < retryTimes {
				time.Sleep(retryPeriod)
				continue
			}
			return nil, fmt.Errorf("err:[%v]", err.Error())
		}
		break
	}
	return result, nil
}
