package commands

import (
	"os"

	"github.com/monax/monax/config"
	"github.com/monax/monax/initialize"
	"github.com/monax/monax/log"
	"github.com/monax/monax/util"

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
	log.Warn(`
	WARNING: after version 0.17, [monax] will no longer support docker-machine
	and will be providing official support for Linux only.`)

}
