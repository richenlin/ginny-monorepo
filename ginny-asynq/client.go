package asyncq

import (
	"context"

	"github.com/hibiken/asynq"
)

// Client wraps the asynq client for enqueuing tasks.
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

	return &Client{
		client:    client,
		Inspector: asynq.NewInspector(redisConnOpt),
	}, nil
}

// EnqueueContext enqueues a task with optional settings.
func (c *Client) EnqueueContext(ctx context.Context, task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.EnqueueContext(ctx, task, opts...)
}

// Close shuts down the asynq client.
func (c *Client) Close() error {
	return c.client.Close()
}
