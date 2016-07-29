package commands

import (
	rem "github.com/eris-ltd/eris-cli/remotes"

	"github.com/spf13/cobra"
)

var Remotes = &cobra.Command{
	Use:   "remotes",
	Short: "manage and perform remote machines and services",
	Long: `display and Manage remote machines which are operating
various services reachable by the Eris platform

Actions, if configured as such, can utilize remote machines.
To register and manage remote machines for sending of actions
to those machines, use this command.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildRemotesCommand() {
	Remotes.AddCommand(remotesAdd)
	Remotes.AddCommand(remotesList)
	Remotes.AddCommand(remotesDo)
	Remotes.AddCommand(remotesEdit)
	Remotes.AddCommand(remotesRename)
	Remotes.AddCommand(remotesRemove)
}

var remotesAdd = &cobra.Command{
	Use:   "add NAME DEFINITION",
	Short: "adds a remote to Eris",
	Long:  `adds a remote to Eris in JSON, TOML, or YAML format`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Add(args)
	},
}

var remotesList = &cobra.Command{
	Use:   "ls",
	Short: "list all registered remotes",
	Long:  `list all registered remotes`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.List()
	},
}

var remotesDo = &cobra.Command{
	Use:   "do NAME",
	Short: "perform an action on a remote",
	Long:  `perform an action on a remote according to the action definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Do(args)
	},
}

var remotesEdit = &cobra.Command{
	Use:   "edit NAME",
	Short: "edit a remote definition file",
	Long:  `edit a remote definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Edit(args)
	},
}

var remotesRename = &cobra.Command{
	Use:   "rename OLD_NAME NEW_NAME",
	Short: "rename a remote",
	Long:  `rename a remote`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Rename(args)
	},
}

var remotesRemove = &cobra.Command{
	Use:   "remove NAME",
	Short: "remove a remote definition file",
	Long:  `remove a remote definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		rem.Remove(args)
	},
}
