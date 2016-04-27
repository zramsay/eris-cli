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
	Long: `create the ~/.eris directory with actions and services subfolders
and clone eris-ltd/eris-actions eris-ltd/eris-services into them, respectively`,
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
	Init.Flags().StringVarP(&do.Source, "source", "", "rawgit", "source from which to download definition files for the eris platform. if toadserver fails, use: rawgit")
	Init.Flags().StringVarP(&do.Proxy, "proxy", "", "", "use a proxy with format: (http://proxyIp:proxyPort). respects $HTTP_PROXY but over-riden by this flag.")
	Init.Flags().BoolVarP(&do.Quiet, "testing", "", false, "DO NOT USE (for testing only)")
}

func Router(cmd *cobra.Command, args []string) {
	err := ini.Initialize(do)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
