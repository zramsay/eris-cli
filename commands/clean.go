package commands

import (
	"github.com/eris-ltd/eris-cli/clean"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//TODO better explanations of command
var Clean = &cobra.Command{
	Use:   "clean",
	Short: "Cleans up your eris working environment",
	Long: `Stops and removes all eris containers (chains, services, datas, etc); 
removes the ~/.eris directory. If you're really sick of marmots, this command can
also remove all images and completely uninstall eris :(
`,
	Run: func(cmd *cobra.Command, args []string) {
		CleanItUp(cmd, args)
	},
}

func buildCleanCommand() {
	addCleanFlags()
}

func addCleanFlags() {
	Clean.Flags().BoolVarP(&do.All, "all", "", false, "removes everything, stopping short of uninstalling eris")
	Clean.Flags().BoolVarP(&do.RmD, "dir", "", false, "remove the eris home directory ~/.eris")
	Clean.Flags().BoolVarP(&do.Images, "images", "", false, "remove all eris docker images")
	//Clean.Flags().BoolVarP(&do.Volumes, "volumes", "", true, "remove orphaned volumes")
	//Clean.Flags().BoolVarP(&do.Uninstall, "uninstall", "", false, "removes everything; leaves no trace of marmot") //gofmt yourself
	Clean.Flags().BoolVarP(&do.Yes, "yes", "", false, "overrides prompts prior to removing things")
}

func CleanItUp(cmd *cobra.Command, args []string) {
	//flag logic handled in Clean
	if err := clean.Clean(do); err != nil {
		return
	}
}
