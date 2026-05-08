package asyncq

import (
	"context"

	"github.com/goriller/ginny-util/graceful"
	"github.com/hibiken/asynq"
)

type Client struct {
	client    *asynq.Client
	Inspector *asynq.Inspector
}

func newClient(ctx context.Context, opt *Config) (*Client, error) {
	var redisConnOpt asynq.RedisConnOpt
	if opt.redisClusterClientOpt != nil {
		redisConnOpt = opt.redisClusterClientOpt
	} else if opt.redisFailoverClientOpt != nil {
		redisConnOpt = opt.redisFailoverClientOpt
	} else {
		redisConnOpt = opt.redisClientOpt
	}
	client := asynq.NewClient(redisConnOpt)
	graceful.AddCloser(func(ctx context.Context) error {
		return client.Close()
	})

	return &Client{
		client:    client,
		Inspector: asynq.NewInspector(redisConnOpt),
	}, nil
}

func (c *Client) EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.EnqueueContext(ctx, task, opts...)
}
