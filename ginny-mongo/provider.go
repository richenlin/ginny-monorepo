package mongo

import "github.com/google/wire"

// Provider
var Provider = wire.NewSet(NewConfig, NewMongo)
