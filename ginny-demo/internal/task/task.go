package task

import (
	"context"

	"github.com/google/wire"
	broker "github.com/goriller/ginny-broker"
)

// TaskProvider
var TaskProvider = wire.NewSet(
	broker.BrokerProvider,
	wire.Struct(new(Task), "*"), wire.Bind(new(ITask), new(*Task)),
)

// ITask
type ITask interface {
	Subscribe(topic []string, h broker.Handler, queue ...string) error
	Publish(ctx context.Context, topic string, msg *broker.Message) error
}

// Task
type Task struct {
	Broker *broker.Broker
}

func NewTask(
	bk *broker.Broker,
) (*Task, error) {
	t := &Task{
		Broker: bk,
	}

	return t, nil
}

// Subscribe 订阅消息 queue 是订阅的队列。queue相同，则只会接收到一个消息，否则接到多个
func (p *Task) Subscribe(topic []string, h broker.Handler, queue ...string) error {
	var q string
	if len(queue) > 0 {
		q = queue[0]
	}
	if q == "" {
		q = "test-consumer-group"
	}
	return p.Broker.Subscribe(context.Background(), topic, q, h, false)
}

// Publish 发送消息
func (p *Task) Publish(ctx context.Context, topic string, msg *broker.Message) error {
	return p.Broker.Publish(ctx, topic, msg)
}
