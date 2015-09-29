package commands

import (
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// flags to add: --no-clone
var Init = &cobra.Command{
	Use:   "init",
	Short: "Initialize the ~/.eris directory with default files or update to latest version",
	Long: `Create the ~/.eris directory with actions and services subfolders
and clone eris-ltd/eris-actions eris-ltd/eris-services into them, respectively.
`,
	Run: func(cmd *cobra.Command, args []string) {
		Router(cmd, args)
	},
}

func buildInitCommand() {
	addInitFlags()
}

func addInitFlags() {
	Init.Flags().BoolVarP(&do.Pull, "skip-pull", "p", false, "do not clone the default services and actions; use the flag when git is not installed")
	Init.Flags().BoolVarP(&do.Services, "services", "", false, "only update the default services (requires git to be installed)")
	Init.Flags().BoolVarP(&do.Actions, "actions", "", false, "only update the default actions (requires git to be installed)")
	Init.Flags().BoolVarP(&do.Yes, "yes", "", false, "over-ride command-line prompts (requires git to be installed)")
}

func Router(cmd *cobra.Command, args []string) {
	if do.Yes {
		do.Services = true
		do.Actions = true
	}

	if !do.Pull {
		util.CheckGitAndGo(true, false)
	}

	ini.Initialize(do)
}
