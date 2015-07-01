package commands

import (
	"fmt"
	"strings"

	"github.com/eris-ltd/eris-cli/data"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------

func ListKnownData(cmd *cobra.Command, args []string) {
	dataCont, err := data.ListKnownRaw()
	IfExit(err)
	for _, s := range dataCont {
		fmt.Println(s)
	}
}

func RenameData(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) != 2 {
		fmt.Println("Please give me: eris data rename [oldName] [newName]")
		return
	}
	IfExit(data.RenameDataRaw(args[0], args[1], ContainerNumber))
}

func InspectData(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	if len(args) == 1 {
		args = append(args, "all")
	}
	IfExit(data.InspectDataRaw(args[0], args[1], ContainerNumber))
}

func RmData(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(data.RmDataRaw(args[0], ContainerNumber))
}

func ImportData(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(data.ImportDataRaw(args[0], ContainerNumber))
}

func ExportData(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	IfExit(data.ExportDataRaw(args[0], ContainerNumber))
}

func ExecData(cmd *cobra.Command, args []string) {
	IfExit(checkServiceGiven(args))
	srv := args[0]

	// if interactive, we ignore args. if not, run args as command
	if !Interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
		if len(args) == 1 {
			args = strings.Split(args[0], " ")
		}
	}

	IfExit(data.ExecDataRaw(srv, ContainerNumber, Interactive, args))
}

//----------------------------------------------------

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
	Data.AddCommand(dataInspect)
	Data.AddCommand(dataExport)
	dataExec.Flags().BoolVarP(&Interactive, "interactive", "i", false, "interactive shell")
	Data.AddCommand(dataExec)
	Data.AddCommand(dataRm)
}

var dataImport = &cobra.Command{
	Use:   "import [name] [folder]",
	Short: "Import a folder to a named data container",
	Long:  `Import a folder to a named data container`,
	Run: func(cmd *cobra.Command, args []string) {
		ImportData(cmd, args)
	},
}

var dataList = &cobra.Command{
	Use:   "ls",
	Short: "List the data containers",
	Long:  `List the data containers`,
	Run: func(cmd *cobra.Command, args []string) {
		ListKnownData(cmd, args)
	},
}

var dataExec = &cobra.Command{
	Use:   "exec",
	Short: "Run a command or interactive shell in in data container",
	Long:  "Run a command or interactive shell in a container with volumes-from the data container",
	Run: func(cmd *cobra.Command, args []string) {
		ExecData(cmd, args)
	},
}

var dataRename = &cobra.Command{
	Use:   "rename [oldName] [newName]",
	Short: "Rename a data container",
	Long:  `Rename a data container`,
	Run: func(cmd *cobra.Command, args []string) {
		RenameData(cmd, args)
	},
}

var dataInspect = &cobra.Command{
	Use:   "inspect [name] [key]",
	Short: "Machine readable details.",
	Long:  `Displays machine readable details about running containers.`,
	Run: func(cmd *cobra.Command, args []string) {
		InspectData(cmd, args)
	},
}

var dataExport = &cobra.Command{
	Use:   "export [name] [folder]",
	Short: "Import a named data container to a folder",
	Long:  `Import a named data container to a folder`,
	Run: func(cmd *cobra.Command, args []string) {
		ExportData(cmd, args)
	},
}

var dataRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Remove a data container",
	Long:  `Remove a data container`,
	Run: func(cmd *cobra.Command, args []string) {
		RmData(cmd, args)
	},
}
