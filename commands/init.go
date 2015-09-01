package commands

import (
	ini "github.com/eris-ltd/eris-cli/initialize"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// flags to add: --no-clone
var Init = &cobra.Command{
	Use:     "init",
	Aliases: []string{"update"},
	Short:   "Initialize the ~/.eris directory with default files or update to latest version",
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
	Init.Flags().BoolVarP(&do.Pull, "pull", "p", false, "git clone the default services and actions; use the flag when git is not installed")
	Init.Flags().StringVarP(&do.Branch, "branch", "b", "master", "specify a branch to update from")
	Init.Flags().BoolVarP(&do.Tool, "tool", "", false, "only update the tool and nothing else")
	Init.Flags().BoolVarP(&do.Services, "services", "", false, "only update the default services")
	Init.Flags().BoolVarP(&do.Actions, "actions", "", false, "only update the default actions")
	Init.Flags().BoolVarP(&do.All, "all", "", false, "update all the above")
}

func Router(cmd *cobra.Command, args []string) {

	if do.Services && do.Actions && !do.All {
		ini.Initialize(do)
	} else if do.All {
		do.Services = true
		do.Actions = true
		ini.Initialize(do)
		do.Tool = true
	} else if !do.Tool {
		do.Services = false
		do.Actions = false
		ini.Initialize(do)
	}

	if do.Tool {
		util.UpdateEris(do.Branch)
	}
}
