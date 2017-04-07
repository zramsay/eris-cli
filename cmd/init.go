package commands

import (
	"os"

	"github.com/monax/cli/config"
	"github.com/monax/cli/initialize"
	"github.com/monax/cli/log"
	"github.com/monax/cli/util"

	"github.com/spf13/cobra"
)

var Init = &cobra.Command{
	Use:   "init",
	Short: "initialize your work space for smart contract glory",
	Long:  `create the root ` + util.Tilde(config.MonaxRoot) + ` directory and subdirectories.`,
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
	err := initialize.Initialize(do)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
