package commands

import (
	"github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Update = &cobra.Command{
	Use:     "update",
	Aliases: []string{"upgrade"},
	Short:   "Update the eris tool.",
	Long: `Fetch the latest version (master branch by default)
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
	Update.Flags().BoolVarP(&do.Pull, "pull-images", "", true, "by default, pulls and/or update latest primary images. use flag to skip pulling/updating of images.")
	Update.Flags().BoolVarP(&do.Yes, "yes", "", false, "over-ride command-line prompts")
	Update.Flags().StringVarP(&do.Source, "source", "", "rawgit", "source from which to download definition files for the eris platform. if toadserver fails, use: rawgit")
}

func UpdateTool(cmd *cobra.Command, args []string) {
	util.UpdateEris(do.Branch, true, true)
	initialize.Initialize(do)
}
