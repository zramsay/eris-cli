package commands

import (
	rem "github.com/eris-ltd/eris-cli/remotes"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

var Remotes = &cobra.Command{
	Use:   "remotes",
	Short: "Manage and Perform Remote Machines and Services.",
	Long: `Display and Manage remote machines which are operating
various services reachable by the Eris platform.

Actions, if configured as such, can utilize remote machines.
To register and manage remote machines for sending of actions
to those machines, use this command.`,
}

// build the services subcommand
func buildRemotesCommand() {
	Remotes.AddCommand(remotesAdd)
	Remotes.AddCommand(remotesList)
	Remotes.AddCommand(remotesDo)
	Remotes.AddCommand(remotesEdit)
	Remotes.AddCommand(remotesRename)
	Remotes.AddCommand(remotesRemove)
}

// add
var remotesAdd = &cobra.Command{
	Use:   "add [name] [remote-definition-file]",
	Short: "Adds a remote to Eris.",
	Long:  `Adds a remote to Eris in JSON, TOML, or YAML format.`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Add(args)
	},
}

// ls
// flags to add: --global --project
var remotesList = &cobra.Command{
	Use:   "ls",
	Short: "List all registered remotes.",
	Long:  `List all registered remotes`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.List()
	},
}

// do
var remotesDo = &cobra.Command{
	Use:   "do [name]",
	Short: "Perform an action on a remote.",
	Long:  `Perform an action on a remote according to the action definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Do(args)
	},
}

// edit
var remotesEdit = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a remote definition file.",
	Long:  `Edit a remote definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Edit(args)
	},
}

// rename
var remotesRename = &cobra.Command{
	Use:   "rename [old] [new]",
	Short: "Rename a remote.",
	Long:  `Rename a remote`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Rename(args)
	},
}

// remove
var remotesRemove = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a remote definition file.",
	Long:  `Remove a remote definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Remove(args)
	},
}
