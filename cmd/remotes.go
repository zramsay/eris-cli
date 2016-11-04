package commands

import (
	"github.com/eris-ltd/eris-cli/remotes"

	"github.com/spf13/cobra"
)

var Remotes = &cobra.Command{
	Use:   "remotes",
	Short: "manage and perform remote machines and services",
	Long: `display and manage remote machines which are operating
various services reachable by Eris`,
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
	Long:  `adds a remote to Eris`,
	Run: func(cmd *cobra.Command, args []string) {
		remotes.Add(args)
	},
}

var remotesList = &cobra.Command{
	Use:   "ls",
	Short: "list all registered remotes",
	Long:  `list all registered remotes`,
	Run: func(cmd *cobra.Command, args []string) {
		remotes.List()
	},
}

var remotesDo = &cobra.Command{
	Use:   "do NAME",
	Short: "perform an action on a remote",
	Long:  `perform an action on a remote according to the action definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		remotes.Do(args)
	},
}

var remotesEdit = &cobra.Command{
	Use:   "edit NAME",
	Short: "edit a remote definition file",
	Long:  `edit a remote definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		remotes.Edit(args)
	},
}

var remotesRename = &cobra.Command{
	Use:   "rename OLD_NAME NEW_NAME",
	Short: "rename a remote",
	Long:  `rename a remote`,
	Run: func(cmd *cobra.Command, args []string) {
		remotes.Rename(args)
	},
}

var remotesRemove = &cobra.Command{
	Use:   "remove NAME",
	Short: "remove a remote definition file",
	Long:  `remove a remote definition file`,
	Run: func(cmd *cobra.Command, args []string) {
		remotes.Remove(args)
	},
}
