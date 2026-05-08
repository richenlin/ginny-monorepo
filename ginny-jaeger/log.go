package jaeger

import (
	"fmt"

	"github.com/goriller/ginny/logger"
)

type stdLog struct{}

// Error
func (l *stdLog) Error(msg string) {
	logger.Action("Jaeger").Error(msg)
}

// Infof
func (l *stdLog) Infof(msg string, args ...interface{}) {
	logger.Action("Jaeger").Info(fmt.Sprintf(msg, args...))
}
