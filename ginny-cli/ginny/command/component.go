package command

import (
	"fmt"

	"github.com/goriller/ginny-cli/ginny/handler"
	"github.com/goriller/ginny-cli/ginny/util"
	"github.com/spf13/cobra"
)

func init() {
	repoCmd.Flags().StringP("type", "t", "repo", "Define the component type, exp: logic、repo...")
	rootCmd.AddCommand(repoCmd)
}

var repoCmd = &cobra.Command{
	Use:   "component",
	Short: "Create component file",
	Long:  "Create component file from template",
	PreRunE: func(cmd *cobra.Command, args []string) (err error) {
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		if err := CheckArgs(args); err != nil {
			return err
		}
		// 获取参数
		repoName := args[0]
		flags := cmd.Flags()
		t, err := flags.GetString("type")
		if err != nil {
			return err
		}
		if err := handler.CreateComponent(repoName, t); err != nil {
			return err
		}

		util.Info(fmt.Sprintf("Create new %s file success", t))
		return nil
	},
}
