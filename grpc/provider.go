package grpc

import (
	"github.com/google/wire"
)

// Provider
var ProviderSet = wire.NewSet(NewServerOptions, NewServer, NewClientOptions, NewClient)
