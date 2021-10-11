package http

import (
	"github.com/google/wire"
)

// Provider
var ProviderSet = wire.NewSet(NewOptions, NewRouter, NewServer, NewClientOptions, NewClient)
