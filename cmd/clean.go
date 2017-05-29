package commands

import (
	"github.com/monax/monax/clean"
	"github.com/monax/monax/config"
	"github.com/monax/monax/util"

	"github.com/spf13/cobra"
)

var Clean = &cobra.Command{
	Use:   "clean",
	Short: "clean up your Monax working environment",
	Long: `by default, this command will stop and force remove all Monax containers
(chains, services, data, etc.) and clean the scratch path. Addtional flags can be used
to remove the Monax home directory and Monax images. Useful during rapid development.`,
	Run: func(cmd *cobra.Command, args []string) {
		CleanItUp(cmd, args)
	},
}

func buildCleanCommand() {
	addCleanFlags()
}

func addCleanFlags() {
	Clean.Flags().BoolVarP(&do.Yes, "yes", "y", false, "overrides prompts prior to removing things")
	Clean.Flags().BoolVarP(&do.All, "all", "a", false, "removes everything, stopping short of uninstalling monax")
	Clean.Flags().BoolVarP(&do.Containers, "containers", "c", true, "remove all monax containers")
	Clean.Flags().BoolVarP(&do.ChnDirs, "chains", "x", false, "remove chain data in "+util.Tilde(config.ChainsPath))
	Clean.Flags().BoolVarP(&do.Scratch, "scratch", "s", true, "remove contents of "+util.Tilde(config.ScratchPath))
	Clean.Flags().BoolVarP(&do.RmD, "dir", "", false, "remove the monax home directory in "+util.Tilde(config.MonaxRoot))
	Clean.Flags().BoolVarP(&do.Images, "images", "i", false, "remove all monax docker images")
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
