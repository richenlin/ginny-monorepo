package asyncq

import (
	"context"

	"github.com/google/wire"
)

var (
	asynqClient *Client
	asynqServer *Server
)

// AsyncqProvider is the wire provider set for asynq.
var AsyncqProvider = wire.NewSet(
	NewConfig,
	NewAsyncq,
)

// Asyncq manages asynq client and server lifecycle.
type Asyncq struct {
	Client *Client
	Server *Server
}

// NewAsyncq creates a new Asyncq instance.
func NewAsyncq(ctx context.Context, opt *Config) (q *Asyncq, err error) {
	asynqClient, err = newClient(ctx, opt)
	if err != nil {
		return
	}

	asynqServer, err = newServer(ctx, opt)
	if err != nil {
		return
	}

	q = &Asyncq{
		Client: asynqClient,
		Server: asynqServer,
	}
	return
}

// Start runs the asynq server in a background goroutine.
func (a *Asyncq) Start() {
	go func() {
		_ = asynqServer.Run()
	}()
}

// Stop gracefully shuts down the asynq server and client.
func (a *Asyncq) Stop() {
	asynqServer.Shutdown()
	_ = asynqClient.Close()
}
