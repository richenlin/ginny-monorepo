package asyncq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
)

var (
	defaultQueue = "default"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type queueType int

const (
	Low queueType = iota
	Default
	Critical
)

// Optional configures a task's properties.
type Optional func(*taskInfo)
type HandlerFunc = func(context.Context, interface{}) (interface{}, error)
type CallHandlerFunc = func(context.Context, interface{})
type CallErrHandlerFunc = func(context.Context, interface{}, error)

var (
	defaultCallHandlerFunc = func(ctx context.Context, param interface{}) {
		slog.DebugContext(ctx, "onSuccess", slog.Any("param", param))
	}
	defaultCallErrHandlerFunc = func(ctx context.Context, param interface{}, err error) {
		slog.ErrorContext(ctx, "onError",
			slog.Any("param", param),
			slog.String("error", err.Error()),
		)
	}
)

// taskInfo describes a registered async task.
type taskInfo struct {
	Step       []StepInfo // step handler functions
	RetryTimes int        // task retry count
	TimeOut    int        // task timeout in seconds
	ProcessIn  int        // task delay in seconds
	Queue      string
	OnSuccess  CallHandlerFunc
	OnError    CallErrHandlerFunc
}

// StepInfo describes a single processing step.
type StepInfo struct {
	Fn          HandlerFunc // first Fn input type is map[string]interface{}
	RetryTimes  int         // step retry count
	RetryPeriod int         // retry interval in milliseconds
	TimeOut     int         // step timeout in milliseconds
}

type taskArg struct {
	TaskType string
	Step     int
	Arg      interface{}
}

// WithRetryTimes sets the task-level retry count.
func WithRetryTimes(r int) Optional {
	return func(t *taskInfo) {
		if r > 0 {
			t.RetryTimes = r
		}
	}
}

// WithTimeOut sets the task-level timeout in seconds.
func WithTimeOut(r int) Optional {
	return func(t *taskInfo) {
		if r > 0 {
			t.TimeOut = r
		}
	}
}

// WithProcessIn delays task execution by r seconds.
func WithProcessIn(r int) Optional {
	return func(t *taskInfo) {
		if r > 0 {
			t.ProcessIn = r
		}
	}
}

// WithQueue sets the task queue.
func WithQueue(s queueType) Optional {
	return func(t *taskInfo) {
		switch s {
		case Critical:
			t.Queue = QueueCritical
		case Low:
			t.Queue = QueueLow
		default:
			t.Queue = QueueDefault
		}
	}
}

// WithOnSuccess sets the success callback.
func WithOnSuccess(f CallHandlerFunc) Optional {
	return func(t *taskInfo) {
		if f != nil {
			t.OnSuccess = f
		}
	}
}

// WithOnError sets the error callback.
func WithOnError(f CallErrHandlerFunc) Optional {
	return func(t *taskInfo) {
		if f != nil {
			t.OnError = f
		}
	}
}

// NewTask registers an async task.
func (a *Asyncq) NewTask(ctx context.Context, taskType string,
	step []StepInfo, opt ...Optional) {

	task := &taskInfo{
		RetryTimes: 2,
		TimeOut:    60 * 60 * 6,
		Step:       step,
		Queue:      defaultQueue,
	}
	for _, v := range opt {
		v(task)
	}
	if task.OnError == nil {
		task.OnError = defaultCallErrHandlerFunc
	}
	if task.OnSuccess == nil {
		task.OnSuccess = defaultCallHandlerFunc
	}

	a.Server.Dispatcher.SetTask(taskType, task)
}

// InvorkTask triggers execution of a registered task.
func (a *Asyncq) InvorkTask(ctx context.Context, taskType string, param interface{}) (string, error) {
	task := a.Server.Dispatcher.GetTask(taskType)
	if task == nil {
		return "", fmt.Errorf("an asynchronous task must be declared before triggering the execution of the task")
	}

	arg := &taskArg{
		TaskType: taskType,
		Arg:      param,
	}

	bt, err := json.Marshal(arg)
	if err != nil {
		return "", err
	}
	t := asynq.NewTask(dispatcherName, bt)

	if task.TimeOut == 0 {
		task.TimeOut = 60 * 60
	}
	if task.RetryTimes == 0 {
		task.RetryTimes = 2
	}

	opt := []asynq.Option{
		asynq.Retention(24 * 30 * 3 * time.Hour),
		asynq.MaxRetry(task.RetryTimes),
		asynq.Timeout(time.Second * time.Duration(task.TimeOut)),
	}

	if task.Queue != "" {
		opt = append(opt, asynq.Queue(task.Queue))
	}

	if task.ProcessIn > 0 {
		opt = append(opt, asynq.ProcessIn(time.Duration(task.ProcessIn)*time.Second))
	}
	info, err := a.Client.EnqueueContext(ctx, t, opt...)
	if err != nil {
		return "", err
	}
	return info.ID, nil
}

// QueryTask queries async task info by task ID.
func (a *Asyncq) QueryTask(ctx context.Context, taskId string) (*asynq.TaskInfo, error) {
	return a.Client.Inspector.GetTaskInfo(defaultQueue, taskId)
}

// StopJobError returns an error that stops the job and executes OnError callback.
// Step and job retries are disabled for this error.
func (a *Asyncq) StopJobError(err error) error {
	return fmt.Errorf("Stop job: [%w]", &stopJobError{msg: err.Error()})
}

// ConvertParams converts task params from arg to input.
func ConvertParams(arg interface{}, input interface{}) error {
	bt, err := json.Marshal(arg)
	if err != nil {
		return err
	}
	return json.Unmarshal(bt, input)
}
