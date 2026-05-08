package COMPONENT_TYPE

import (
	"context"

	"github.com/google/wire"
	"github.com/goriller/ginny/logger"
	"go.uber.org/zap"
)

// COMPONENT_NAMECOMPONENT_UP_TYPEProviderSet
var COMPONENT_NAMECOMPONENT_UP_TYPEProviderSet = wire.NewSet(
	NewCOMPONENT_NAMECOMPONENT_UP_TYPE,
)

// COMPONENT_NAMECOMPONENT_UP_TYPE
type COMPONENT_NAMECOMPONENT_UP_TYPE struct {
}

// NewCOMPONENT_NAMECOMPONENT_UP_TYPE
func NewCOMPONENT_NAMECOMPONENT_UP_TYPE() (*COMPONENT_NAMECOMPONENT_UP_TYPE, error) {
	return &COMPONENT_NAMECOMPONENT_UP_TYPE{}, nil
}

func (p *COMPONENT_NAMECOMPONENT_UP_TYPE) Test(ctx context.Context) error {
	log := logger.WithContext(ctx).With(zap.String("action", "Test"))
	log.Debug("xx", zap.String("xx", "xx"))
	return nil
}
