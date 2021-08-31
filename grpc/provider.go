package grpc

import (
	"github.com/google/wire"
)

// Provider
var ProviderSet = wire.NewSet(NewOptions, NewServer, NewClientOptions, NewClient)
