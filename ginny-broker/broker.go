// Package broker provider kafka or other mq interface
package broker

import (
	"context"
	"net/url"

	graceful "github.com/goriller/ginny-util/graceful"
	"go.uber.org/zap"
)

// Broker broker for message queue.
type Broker struct {
	mq
}

// mq is an interface used for CloudEvents asynchronous messaging.
type mq interface {
	// Init 连接
	Init(ctx context.Context, log *zap.Logger, uri *url.URL) error
	// Close 断开连接
	Close(ctx context.Context) error
	// Publish 发送消息
	Publish(ctx context.Context, topic string, msg *Message) error
	// Subscribe 订阅：queue 是订阅的队列。queue相同，则只会接收到一个消息，否则接到多个， autoAck 为自动ACK
	Subscribe(ctx context.Context, topic []string, queue string, h Handler, autoAck bool) error
	// Unsubscribe 取消订阅
	Unsubscribe(ctx context.Context, topic []string) error
}

// Handler is used to process messages via a subscription of a topic.
type Handler func(Publication) error

// Publication is given to a subscription handler for processing.
type Publication interface {
	Message() *Message
	Ack() error
	Topic() string
}

// Message a message of common msq.
type Message struct {
	Header map[string]string
	Body   []byte
}

var implements = make(map[string]mq)

// RegisterImplements 注册实现类
func RegisterImplements(scheme string, i mq) {
	implements[scheme] = i
}

// Init 初始化
func (b *Broker) Init(ctx context.Context, log *zap.Logger, u *url.URL) error {
	for k, v := range implements {
		if k == u.Scheme {
			b.mq = v
		}
	}
	if b.mq != nil {
		graceful.AddCloser(func(ctx context.Context) error {
			return b.mq.Close(ctx)
		})
		return b.mq.Init(ctx, log, u)
	}

	return nil
}
