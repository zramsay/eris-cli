package commands

import (
	"github.com/eris-ltd/eris-cli/update"

	. "github.com/eris-ltd/common/go/common"
	"github.com/spf13/cobra"
)

var Update = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade"},
	Short:   "update the eris tool",
	Long: `fetch the latest version (master branch by default)
and re-install eris. Once eris is reinstalled, then the
eris init function will be called automatically for you
in order to update your definition files and images.

If you have made modifications to the default definition files
then you will want to make backups of those **before** upgrading
your eris installation.`,
	Run: func(cmd *cobra.Command, args []string) {
		UpdateTool(cmd, args)
	},
}

func buildUpdateCommand() {
	addUpdateFlags()
}

func addUpdateFlags() {
	Update.Flags().StringVarP(&do.Branch, "branch", "b", "master", "specify a branch to update from")
}

func UpdateTool(cmd *cobra.Command, args []string) {
	IfExit(update.UpdateEris(do))
}
