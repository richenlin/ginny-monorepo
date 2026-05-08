package orm

import (
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
	"moul.io/zapgorm2"
)

func newLogger(log *zap.Logger) logger.Interface {
	l := &zapgorm2.Logger{
		ZapLogger:                 log,
		LogLevel:                  logger.Info,
		SlowThreshold:             100 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: false,
		Context:                   nil,
	}
	l.SetAsDefault()
	return l
}
