package commands

import (
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Update = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade"},
	Short:   "Update the eris tool.",
	Long: `Fetch the latest version (master branch by default)
and re-install eris; requires git and go to be installed.`,
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
	util.UpdateEris(do.Branch, true, true)

}
