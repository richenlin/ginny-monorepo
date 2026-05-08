package client

import (
	"github.com/google/wire"

	consul "github.com/goriller/ginny-consul"
)

// ProviderSet
var ProviderSet = wire.NewSet(
	consul.ProviderSet,
	NewExampleClient,
)
