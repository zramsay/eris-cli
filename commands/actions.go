package commands

import (
	act "github.com/eris-ltd/eris-cli/actions"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Actions = &cobra.Command{
	Use:   "actions",
	Short: "Manage and Perform Structured Actions.",
	Long: `Display and Manage actions for various components of the
Eris platform and for the platform itself.

Actions are bundles of commands which rely upon a project
which is currently in scope or a any set of installed
services. Actions are held in yaml, toml, or json
action-definition files within the action folder in the
eris tree (globally scoped actions) or in a directory
pointed to by the actions field of the currently checked
out project (project scoped actions). Actions are a
sequence of commands which operate in a similar fashion
to how a circle.yml file may operate.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the services subcommand
func buildActionsCommand() {
	Actions.AddCommand(actionsGet)
	Actions.AddCommand(actionsNew)
	Actions.AddCommand(actionsList)
	Actions.AddCommand(actionsDo)
	Actions.AddCommand(actionsEdit)
	Actions.AddCommand(actionsRename)
	Actions.AddCommand(actionsRemove)
	addActionsFlags()
}

func addActionsFlags() {
	actionsRemove.Flags().BoolVarP(&Force, "force", "f", false, "force action")
}

// get an actions definition file from a remote (currently limited to github.com and ipfs)
var actionsGet = &cobra.Command{
	Use:   "get [name] [github.com/USER/REPO] || [name] [ipfs hash]",
	Short: "Get a set of actions from Github or IPFS.",
	Long: `Retrieve an action file from the internet (utilizes git clone
or ipfs).

NOTE: This functionality is currently limited to github.com and IPFS.`,
	Run: func(cmd *cobra.Command, args []string) {
		act.Get(cmd, args)
	},
}

// new builds a new action definition file
// flags to add: --template
var actionsNew = &cobra.Command{
	Use:   "new [name]",
	Short: "Create a new action definition file.",
	Long:  `Create a new action definition file optionally from a template.`,
	Run: func(cmd *cobra.Command, args []string) {
		act.New(cmd, args)
	},
}

// ls
// flags to add: --global --project
var actionsList = &cobra.Command{
	Use:   "ls",
	Short: "List all registered action definition files.",
	Long:  `List all registered action definition files.`,
	Run: func(cmd *cobra.Command, args []string) {
		act.ListKnown()
	},
}

// do
var actionsDo = &cobra.Command{
	Use:   "do [name]",
	Short: "Perform an action.",
	Long:  `Perform an action according to the action definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		act.Do(cmd, args)
	},
}

// edit
var actionsEdit = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit an action definition file.",
	Long:  `Edit an action definition file in the default editor.`,
	Run: func(cmd *cobra.Command, args []string) {
		act.Edit(args)
	},
}

// rename
var actionsRename = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename an action.",
	Long:  `Rename an action.`,
	Run: func(cmd *cobra.Command, args []string) {
		act.Rename(cmd, args)
	},
}

// remove
var actionsRemove = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove an action definition file.",
	Long:  `Remove an action definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		act.Rm(cmd, args)
	},
}
