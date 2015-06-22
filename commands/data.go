package commands

import (
	"github.com/eris-ltd/eris-cli/data"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Data Sub-Command
var Data = &cobra.Command{
	Use:   "data",
	Short: "Manage Data containers for your Application.",
	Long: `The data subcommand is used to import, and export
data into containers for use by your application.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the data subcommand
func buildDataCommand() {
	Data.AddCommand(dataImport)
	Data.AddCommand(dataList)
	Data.AddCommand(dataRename)
	Data.AddCommand(dataExport)
	Data.AddCommand(dataRm)
}

var dataImport = &cobra.Command{
	Use:   "import [name] [folder]",
	Short: "Import a folder to a named data container",
	Long:  `Import a folder to a named data container`,
	Run: func(cmd *cobra.Command, args []string) {
		data.Import(args)
	},
}

var dataList = &cobra.Command{
	Use:   "ls",
	Short: "List the data containers",
	Long:  `List the data containers`,
	Run: func(cmd *cobra.Command, args []string) {
		data.ListKnown(args)
	},
}

var dataRename = &cobra.Command{
	Use:   "rename [oldName] [newName]",
	Short: "Rename a data container",
	Long:  `Rename a data container`,
	Run: func(cmd *cobra.Command, args []string) {
		data.Rename(cmd, args)
	},
}

var dataExport = &cobra.Command{
	Use:   "export [name] [folder]",
	Short: "Import a named data container to a folder",
	Long:  `Import a named data container to a folder`,
	Run: func(cmd *cobra.Command, args []string) {
		data.Export(args)
	},
}

var dataRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a data container",
	Long:  `Remove a data container`,
	Run: func(cmd *cobra.Command, args []string) {
		data.Rm(cmd, args)
	},
}
