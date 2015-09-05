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
	Init.Flags().BoolVarP(&do.Pull, "skip-pull", "p", false, "do not clone the default services and actions; use the flag when git is not installed")
	Init.Flags().BoolVarP(&do.Services, "services", "", false, "only update the default services (requires git to be installed)")
	Init.Flags().BoolVarP(&do.Actions, "actions", "", false, "only update the default actions (requires git to be installed)")
	Init.Flags().BoolVarP(&do.Yes, "yes", "", false, "over-ride command-line prompts (requires git to be installed)")
	Init.Flags().BoolVarP(&do.Tool, "tool", "", false, "only update the eris cli tool and nothing else (requires git and go to be installed)")
	Init.Flags().StringVarP(&do.Branch, "branch", "b", "master", "specify a branch to update from (mostly used for eris update) (requires git to be installed)")
	Init.Flags().BoolVarP(&do.All, "all", "", false, "update all the above and skip command-line prompts (requires git and go to be installed)")
}

func Router(cmd *cobra.Command, args []string) {
	switch {
	case do.Yes:
		do.Services = true
		do.Actions = true
	case do.All:
		do.Services = true
		do.Actions = true
		do.Tool = true
	}

	if do.Tool {
		util.UpdateEris(do.Branch)

		if !do.All {
			return
		}
	}

	ini.Initialize(do)
}
