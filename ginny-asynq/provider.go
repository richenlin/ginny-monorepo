package asyncq

import (
	"context"

	"github.com/google/wire"
)

var (
	asynqClient *Client
	asynqServer *Server
)

var AsyncqProvider = wire.NewSet(
	NewConfig,
	NewAsyncq,
)

type Asyncq struct {
	Client *Client
	Server *Server
}

func NewAsyncq(ctx context.Context, opt *Config) (q *Asyncq, err error) {
	// log := logger.GetLogger(ctx, nil)
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

func (a *Asyncq) Start() error {
	err := asynqServer.Start()
	if err != nil {
		return err
	}

	return nil
}
