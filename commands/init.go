package commands

import (
	ini "github.com/eris-ltd/eris-cli/initialize"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// flags to add: --no-clone
var Init = &cobra.Command{
	Use:   "init",
	Short: "Initialize the ~/.eris directory with some default services and actions, and pull important images",
	Long: `Create the ~/.eris directory with actions and services subfolders
and clone eris-ltd/eris-actions eris-ltd/eris-services into them, respectively.

Pull important images; --dev flags pulls images needed for development`,
	Run: func(cmd *cobra.Command, args []string) {
		ini.Initialize(do.Pull, do.Verbose, do.Dev)
	},
}
