package commands

import (
	"github.com/eris-ltd/eris-cli/clean"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//TODO better explanations of command
var Clean = &cobra.Command{
	Use:   "clean",
	Short: "Clean up your eris working environment.",
	Long: `By default, this command will stop and
	force remove all eris containers (chains, services, 
	datas, etc). Addtional flags can be used to remove
	the eris home directory and eris images. Useful
	for rapid development with docker containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		CleanItUp(cmd, args)
	},
}

func buildCleanCommand() {
	addCleanFlags()
}

func addCleanFlags() {
	Clean.Flags().BoolVarP(&do.All, "all", "a", false, "removes everything, stopping short of uninstalling eris")
	Clean.Flags().BoolVarP(&do.Containers, "containers", "c", true, "remove all eris containers")
	Clean.Flags().BoolVarP(&do.Scratch, "scratch", "s", true, "remove contents of: $HOME/.eris/scratch")
	Clean.Flags().BoolVarP(&do.RmD, "dir", "", false, "remove the eris home directory: $HOME/.eris")
	Clean.Flags().BoolVarP(&do.Images, "images", "i", false, "remove all eris docker images")
	//Clean.Flags().BoolVarP(&do.Volumes, "volumes", "", true, "remove orphaned volumes")
	//Clean.Flags().BoolVarP(&do.Uninstall, "uninstall", "", false, "removes everything; leaves no trace of marmot") //gofmt yourself
	Clean.Flags().BoolVarP(&do.Yes, "yes", "y", false, "overrides prompts prior to removing things")
}

func CleanItUp(cmd *cobra.Command, args []string) {

	if do.All {
		do.Containers = true
		do.Scratch = true
		do.RmD = true
		do.Images = true
	}

	if err := clean.Clean(do); err != nil {
		return
	}
}
