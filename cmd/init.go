package commands

import (
	"os"

	ini "github.com/eris-ltd/eris-cli/initialize"

	log "github.com/eris-ltd/eris-logger"
	"github.com/spf13/cobra"
)

// flags to add: --no-clone
var Init = &cobra.Command{
	Use:   "init",
	Short: "initialize your work space for smart contract glory",
	Long: `create the ~/.eris directory with services subfolders
and clone github.com/eris-ltd/eris-services into them.`,
	Run: func(cmd *cobra.Command, args []string) {
		Router(cmd, args)
	},
}

func buildInitCommand() {
	addInitFlags()
}

func addInitFlags() {
	Init.Flags().BoolVarP(&do.Pull, "pull-images", "", true, "by default, pulls and/or update latest primary images. use flag to skip pulling/updating of images.")
	Init.Flags().BoolVarP(&do.Yes, "yes", "y", false, "over-ride command-line prompts")
	Init.Flags().BoolVarP(&do.Quiet, "testing", "", false, "DO NOT USE (for testing only)")
}

func Router(cmd *cobra.Command, args []string) {
	err := ini.Initialize(do)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
