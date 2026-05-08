//
package util

import (
	"fmt"
)

// Info should be used to describe the example commands that are about to run.
func Info(args ...interface{}) {
	fmt.Printf("\x1b[34;1m%s\x1b[0m\n", fmt.Sprintf("Info: %s", args...))
}

// Warning should be used to display a warning
func Warning(args ...interface{}) {
	fmt.Printf("\x1b[36;1m%s\x1b[0m\n", fmt.Sprintf("Warning: %s", args...))
}

// Error should be used to display a error
func Error(args ...interface{}) {
	fmt.Printf("\x1b[31;1m%s\x1b[0m\n", fmt.Sprintf("Error: %#v", args...))
}
