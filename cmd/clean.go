package commands

import (
	"github.com/eris-ltd/eris-cli/clean"
	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/spf13/cobra"
)

var Clean = &cobra.Command{
	Use:   "clean",
	Short: "clean up your Eris working environment",
	Long: `by default, this command will stop and force remove all Eris containers
(chains, services, data, etc.) and clean the scratch path, as well as latent directories
and files in the ` + util.Tilde(config.ChainsPath) + ` directory. Addtional flags can be used to remove
the Eris home directory and Eris images. Useful for rapid development
with Docker containers`,
	Run: func(cmd *cobra.Command, args []string) {
		CleanItUp(cmd, args)
	},
}

func buildCleanCommand() {
	addCleanFlags()
}

func addCleanFlags() {
	Clean.Flags().BoolVarP(&do.Yes, "yes", "y", false, "overrides prompts prior to removing things")
	Clean.Flags().BoolVarP(&do.All, "all", "a", false, "removes everything, stopping short of uninstalling eris")
	Clean.Flags().BoolVarP(&do.Containers, "containers", "c", true, "remove all eris containers")
	Clean.Flags().BoolVarP(&do.ChnDirs, "chains", "", false, "remove latent chain data in "+util.Tilde(config.ChainsPath))
	Clean.Flags().BoolVarP(&do.Scratch, "scratch", "s", true, "remove contents of "+util.Tilde(config.ScratchPath))
	Clean.Flags().BoolVarP(&do.RmD, "dir", "", false, "remove the eris home directory in "+util.Tilde(config.ErisRoot))
	Clean.Flags().BoolVarP(&do.Images, "images", "i", false, "remove all eris docker images")
}

func CleanItUp(cmd *cobra.Command, args []string) {
	if do.All {
		do.Containers = true
		do.Scratch = true
		do.ChnDirs = true
		do.RmD = true
		do.Images = true
	}

	util.IfExit(clean.Clean(do))
}
