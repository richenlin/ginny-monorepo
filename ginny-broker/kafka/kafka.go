// Package kafka provider kafka implement for broker
package kafka

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/google/uuid"
	broker "github.com/goriller/ginny-broker"
	"go.uber.org/zap"
)

type kafkaBroker struct {
	c           sarama.Client
	p           sarama.SyncProducer
	subscribers *sync.Map
	sc          []sarama.Client
	opts        kafkaOpts
	scMutex     sync.Mutex
	connected   bool
}

type publication struct {
	err   error
	cg    sarama.ConsumerGroup
	sess  sarama.ConsumerGroupSession
	km    *sarama.ConsumerMessage
	m     *broker.Message
	topic string
}

type kafkaOpts struct {
	defaultQueue string
	addr         []string
	version      sarama.KafkaVersion
	logger       *zap.Logger
}

func init() {
	broker.RegisterImplements("kafka", &kafkaBroker{})
}

// Init for init
func (k *kafkaBroker) Init(ctx context.Context, log *zap.Logger, u *url.URL) error {
	var cAddrs []string
	if u.Host != "" {
		cAddrs = strings.Split(u.Host, ",")
	}
	var err error
	version := u.Query().Get("v")
	var v sarama.KafkaVersion
	if version != "" {
		v, err = sarama.ParseKafkaVersion(version)
		if err != nil {
			return fmt.Errorf(" parse kafka version %s error for %w ", version, err)
		}
	} else {
		v = sarama.V2_0_0_0
	}

	var defaultQueue string
	if len(u.Path) > 1 {
		defaultQueue = u.Path[1:]
	}
	k.subscribers = &sync.Map{}

	k.opts = kafkaOpts{
		version:      v,
		defaultQueue: defaultQueue,
		addr:         cAddrs,
		logger:       log,
	}
	return k.connect(ctx)
}

// Message message
func (p *publication) Message() *broker.Message {
	return p.m
}

// Ack Ack
func (p *publication) Ack() error {
	p.sess.MarkMessage(p.km, "")
	return nil
}

// Topic Topic
func (p *publication) Topic() string {
	return p.topic
}

// Error Error
func (p *publication) Error() error {
	return p.err
}

// Close Close
func (k *kafkaBroker) Close(ctx context.Context) error {
	k.scMutex.Lock()
	defer k.scMutex.Unlock()
	for _, client := range k.sc {
		_ = client.Close()
	}
	k.sc = nil
	_ = k.p.Close()
	if err := k.c.Close(); err != nil {
		return err
	}
	k.connected = false
	return nil
}

func (k *kafkaBroker) connect(_ context.Context) error {
	if k.connected {
		return nil
	}
	k.scMutex.Lock()
	if k.c != nil {
		k.scMutex.Unlock()
		return nil
	}
	k.scMutex.Unlock()
	pconfig := sarama.NewConfig()
	pconfig.Version = k.opts.version
	// For implementation reasons, the SyncProducer requires
	// `Producer.Return.Errors` and `Producer.Return.Successes`
	// to be set to true in its configuration.
	pconfig.Producer.Return.Successes = true
	pconfig.Producer.Return.Errors = true
	pconfig.Producer.Retry.Max = 5
	pconfig.Producer.RequiredAcks = sarama.WaitForLocal

	c, err := sarama.NewClient(k.opts.addr, pconfig)
	if err != nil {
		return err
	}
	p, err := sarama.NewSyncProducerFromClient(c)
	if err != nil {
		return err
	}
	k.scMutex.Lock()
	k.c = c
	k.p = p
	k.sc = make([]sarama.Client, 0)
	k.connected = true
	defer k.scMutex.Unlock()
	return nil
}

// Unsubscribe Unsubscribe
func (k *kafkaBroker) Unsubscribe(ctx context.Context, topics []string) error {
	key := strings.Join(topics, ",")
	if v, okv := k.subscribers.Load(key); okv {
		if cg, ok := v.(sarama.ConsumerGroup); ok {
			return cg.Close()
		}
	}
	return nil
}

// Subscribe subscribe
func (k *kafkaBroker) Subscribe(ctx context.Context, topics []string,
	consumerGroup string, handler broker.Handler, autoAck bool,
) error {
	if consumerGroup == "" {
		consumerGroup = k.opts.defaultQueue
	}
	if len(topics) == 0 {
		topics = []string{k.opts.defaultQueue}
	}
	// we need to create a new client per consumer
	c, err := k.getSaramaClusterClient()
	if err != nil {
		return err
	}
	cg, err := sarama.NewConsumerGroupFromClient(consumerGroup, c)
	if err != nil {
		return err
	}
	key := strings.Join(topics, ",")
	k.subscribers.Store(key, cg)
	h := &consumerGroupHandler{
		handler: handler,
		cg:      cg,
		autoAck: autoAck,
		logger:  k.opts.logger,
	}

	go func() {
		for {
			select {
			case err := <-cg.Errors():
				if !errors.Is(err, nil) {
					k.opts.logger.With(zap.String("action", "kafka")).Error("consumer error", zap.Error(err))
				}
			default:
				err := cg.Consume(ctx, topics, h)
				if err == nil {
					continue
				}
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				k.opts.logger.With(zap.String("action", "kafka")).Error("err", zap.Error(err))
			}
		}
	}()
	return nil
}

// Publish publish implement
func (k *kafkaBroker) Publish(_ context.Context, topic string, msg *broker.Message) error {
	if topic == "" {
		topic = k.opts.defaultQueue
	}
	headers := make([]sarama.RecordHeader, 0)
	for k, v := range msg.Header {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}
	_, _, err := k.p.SendMessage(&sarama.ProducerMessage{
		Topic:   topic,
		Value:   sarama.ByteEncoder(msg.Body),
		Headers: headers,
		Key:     sarama.StringEncoder(uuid.New().String()),
	})
	return err
}

func (k *kafkaBroker) getSaramaClusterClient() (sarama.Client, error) {
	clusterConfig := sarama.NewConfig()
	clusterConfig.Version = k.opts.version
	clusterConfig.Consumer.Return.Errors = true
	clusterConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	clusterConfig.Consumer.Offsets.CommitInterval = 1 * time.Second
	clusterConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange

	cs, err := sarama.NewClient(k.opts.addr, clusterConfig)
	if err != nil {
		return nil, err
	}
	k.scMutex.Lock()
	defer k.scMutex.Unlock()
	k.sc = append(k.sc, cs)
	return cs, nil
}

// consumerGroupHandler is the implementation of sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler broker.Handler
	cg      sarama.ConsumerGroup
	autoAck bool
	logger  *zap.Logger
}

// Setup Setup
func (*consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error { return nil }

// Cleanup Cleanup
func (*consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim ConsumeClaim
func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		m := broker.Message{
			Header: make(map[string]string),
		}
		for _, v := range msg.Headers {
			m.Header[string(v.Key)] = string(v.Value)
		}
		m.Body = msg.Value
		p := &publication{topic: msg.Topic, m: &m, km: msg, cg: h.cg, sess: sess}
		err := h.handler(p)
		if err == nil && h.autoAck {
			sess.MarkMessage(msg, "")
		} else if err != nil {
			h.logger.With(zap.String("action", "kafka")).Error("subscriber error", zap.Error(p.err))
		}
	}
	return nil
}
