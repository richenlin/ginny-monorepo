package command

import (
	"fmt"

	"github.com/goriller/ginny-cli/ginny/options"
	"github.com/goriller/ginny-cli/ginny/util"
)

// CheckArgs should be used to ensure the right command line arguments are
// passed before executing an example.
func CheckArgs(args []string) error {
	if len(args) < 1 {
		msg := "Missing required parameters"
		util.Error(msg)
		return fmt.Errorf(msg)
	}
	return nil
}

// ShowLogo
func ShowLogo() {
	fmt.Printf("\x1b[35;1m%s\x1b[0m\n", options.Logo)
}
