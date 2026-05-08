package command

import (
	"github.com/goriller/ginny-cli/ginny/handler"
	"github.com/goriller/ginny-cli/ginny/util"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serviceCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Create service file by proto file",
	Long:  "Create service file by proto file",
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		// if err := CheckArgs(args); err != nil {
		// 	return err
		// }

		if err := handler.CreateService(); err != nil {
			return err
		}

		util.Info("Create new service file success!")
		return nil
	},
}
