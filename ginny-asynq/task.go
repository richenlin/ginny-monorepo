package asyncq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/goriller/ginny/logger"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

var (
	// default
	defaultQueue = "default"
)

const (
	// queue name
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

// Optional the Options for this module
type Optional func(*taskInfo)
type HandlerFunc = func(context.Context, interface{}) (interface{}, error)
type CallHandlerFunc = func(context.Context, interface{})
type CallErrHandlerFunc = func(context.Context, interface{}, error)

var (
	defaultCallHandlerFunc = func(ctx context.Context, param interface{}) {
		log := logger.WithContext(ctx)
		log.Info("onSuccess: ",
			zap.Any("param", param))
	}
	defaultCallErrHandlerFunc = func(ctx context.Context, param interface{}, err error) {
		log := logger.WithContext(ctx)
		log.Error("onError: ",
			zap.Any("param", param), zap.Error(err))
	}
)

// taskInfo is used to dispatch tasks to registered handlers.
type taskInfo struct {
	Step       []StepInfo // 步骤处理函数
	RetryTimes int        // 任务重试次数
	TimeOut    int        // 任务超时时间 s
	ProcessIn  int        // 任务延迟执行 s
	Queue      string
	OnSuccess  CallHandlerFunc
	OnError    CallErrHandlerFunc
}

// StepInfo
type StepInfo struct {
	Fn          HandlerFunc // 注意第一个Fn入参类型是 map[string]interface{}
	RetryTimes  int         // 单步重试次数
	RetryPeriod int         // 单步重试间隔, Millisecond(毫秒)
	TimeOut     int         // 单步超时时间, Millisecond(毫秒)
}

type taskArg struct {
	TaskType string
	Step     int
	Arg      interface{}
}

// WithRetryTimes
func WithRetryTimes(r int) Optional {
	return func(t *taskInfo) {
		if r > 0 {
			t.RetryTimes = r
		}
	}
}

// WithTimeOut
func WithTimeOut(r int) Optional {
	return func(t *taskInfo) {
		if r > 0 {
			t.TimeOut = r
		}
	}
}

// WithProcessIn
func WithProcessIn(r int) Optional {
	return func(t *taskInfo) {
		if r > 0 {
			t.ProcessIn = r
		}
	}
}

// WithQueue
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

// WithOnSuccess
func WithOnSuccess(f CallHandlerFunc) Optional {
	return func(t *taskInfo) {
		if f != nil {
			t.OnSuccess = f
		}
	}
}

// WithOnError
func WithOnError(f CallErrHandlerFunc) Optional {
	return func(t *taskInfo) {
		if f != nil {
			t.OnError = f
		}
	}
}

// NewTask 声明异步任务
func (a *Asyncq) NewTask(ctx context.Context, taskType string,
	step []StepInfo, opt ...Optional) {

	task := &taskInfo{
		RetryTimes: 2,           // 任务队列重试次数
		TimeOut:    60 * 60 * 6, // 任务队列最大执行时间6小时
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

// InvorkTask 触发执行任务,传入异步任务参数
func (a *Asyncq) InvorkTask(ctx context.Context, taskType string, param interface{}) (string, error) {
	task := a.Server.Dispatcher.GetTask(taskType)
	if task == nil {
		return "", fmt.Errorf("An asynchronous task must be declared before triggering the execution of the task")
	}

	arg := &taskArg{
		TaskType: taskType,
		Arg:      param,
	}

	bt, err := json.Marshal(arg)
	if err != nil {
		return "", err
	}
	// log.Info("触发异步任务: %s, param: %v", taskType, string(bt))
	t := asynq.NewTask(dispatcherName, bt)

	if task.TimeOut == 0 {
		task.TimeOut = 60 * 60 // 默认最大超时1小时
	}
	if task.RetryTimes == 0 {
		task.RetryTimes = 2
	}

	opt := []asynq.Option{
		asynq.Retention(24 * 30 * 3 * time.Hour), // 任务完成后默认存储时间
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

// QueryTask 查询异步任务信息
func (a *Asyncq) QueryTask(ctx context.Context, taskId string) (*asynq.TaskInfo, error) {
	return a.Client.Inspector.GetTaskInfo(defaultQueue, taskId)
}

// StopJobError 抛出该错误,Job跳过剩余步骤逻辑,并执行onError回调. Step以及Job的retry不再生效
func (a *Asyncq) StopJobError(err error) error {
	return fmt.Errorf("Stop job: [%w]", &stopJobError{msg: err.Error()})
}

// ConvertParams
func ConvertParams(arg interface{}, input interface{}) error {
	bt, err := json.Marshal(arg)
	if err != nil {
		return err
	}
	return json.Unmarshal(bt, input)
}
