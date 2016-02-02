package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/data"
	"github.com/eris-ltd/eris-cli/list"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------

// Primary Data Sub-Command
var Data = &cobra.Command{
	Use:   "data",
	Short: "Manage data containers for your application.",
	Long: `The data subcommand is used to import, and export
data into containers for use by your application.

The [eris data import] and [eris data export] commands should be
thought of from the point of view of the container.

The [eris data import] command sends a directory *as is* from
SRC on the host to an existing DEST inside of the data container.

The [eris data export] command performs this process in the reverse.
It sucks out whatever is in the SRC directory in the data container
and sticks it back into a DEST directory on the host.

Notes:
- container paths enter at /home/eris/.eris
- import host path must be absolute, export host path is indifferent

At Eris, we use this functionality to formulate little JSONs
and configs on the host and then "stick them back into the
containers"`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the data subcommand
func buildDataCommand() {
	Data.AddCommand(dataImport)
	Data.AddCommand(dataList)
	Data.AddCommand(dataRename)
	Data.AddCommand(dataInspect)
	Data.AddCommand(dataExport)
	Data.AddCommand(dataExec)
	Data.AddCommand(dataRm)
	addDataFlags()
}

var dataImport = &cobra.Command{
	Use:   "import NAME SRC DEST",
	Short: "Import from a host folder to a named data container's directory",
	Long: `Import from a host folder to a named data container's directory.
Requires src and dest for each host and container, respectively.
Container path enters at /home/eris/.eris
Source (host) path must be absolute and destination dir must exist.`,
	Run: ImportData,
}

var dataExport = &cobra.Command{
	Use:   "export NAME SRC DEST",
	Short: "Export a named data container's directory to a host directory",
	Long: `Export a named data container's directory to a host directory.
Requires src and dest for each container and host, respectively.
Container path enters at /home/eris/.eris
Destination (host) path can be relative.`,
	Run: ExportData,
}

var dataList = &cobra.Command{
	Use:   "ls",
	Short: "List the data containers eris manages for you",
	Long:  `List the data containers eris manages for you`,
	Run:   ListKnownData,
}

var dataExec = &cobra.Command{
	Use:   "exec",
	Short: "Run a command or interactive shell in a data container",
	Long: `Run a command or interactive shell in a container with
volumes-from the data container.

Exec can be used to run a single one off command to interact
with the data. Use it for things like ls.

If you want to pass flags into the command that is run in the
data container, please surround the command you want to pass
in with double quotes. Use it like this: "ls -la".

Exec instances run as the Eris user.

Exec can also be used as an interactive shell. When put in
this mode, you can "get inside of" your containers. You will
have root access to a throwaway container which has the volumes
of the data container mounted to it.`,
	Example: `$ eris data exec name ls /home/eris/.eris -- will list the eris dir
$ eris data exec name "ls -la /home/eris/.eris" -- will pass flags to the ls command
$ eris data exec --interactive name -- will start interactive console`,
	Run: ExecData,
}

var dataRename = &cobra.Command{
	Use:   "rename OLD_NAME NEW_NAME",
	Short: "Rename a data container",
	Long:  `Rename a data container`,
	Run:   RenameData,
}

var dataInspect = &cobra.Command{
	Use:   "inspect NAME [KEY]",
	Short: "Show machine readable details.",
	Long:  `Display machine readable details about running containers.`,
	Run:   InspectData,
}

var dataRm = &cobra.Command{
	Use:   "rm NAME",
	Short: "Remove a data container",
	Long:  `Remove a data container`,
	Run:   RmData,
}

//----------------------------------------------------

func addDataFlags() {
	dataRm.Flags().BoolVarP(&do.RmHF, "dir", "", false, "remove data folder from host")

	dataList.Flags().BoolVarP(&do.Quiet, "quiet", "", false, "machine readable output; also used in tests")

	buildFlag(dataRm, do, "rm-volumes", "data")

	buildFlag(dataExec, do, "interactive", "data")
}

//----------------------------------------------------
func ListKnownData(cmd *cobra.Command, args []string) {
	if err := list.ListDatas(do); err != nil {
		return
	}
}

func RenameData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.NewName = args[1]
	IfExit(data.RenameData(do))
}

func InspectData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	if len(args) == 1 {
		do.Operations.Args = []string{"all"}
	} else {
		do.Operations.Args = []string{args[1]}
	}

	IfExit(data.InspectData(do))
}

func RmData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	IfExit(data.RmData(do))
}

//src on host, dest in container
func ImportData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(3, "eq", cmd, args))
	do.Name = args[0]
	do.Source = args[1]
	do.Destination = args[2]
	IfExit(data.ImportData(do))
}

//src in container, dest on host
func ExportData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(3, "eq", cmd, args))
	do.Name = args[0]
	do.Source = args[1]
	do.Destination = args[2]
	IfExit(data.ExportData(do))
}

func ExecData(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]

	// if interactive, we ignore args. if not, run args as command
	if !do.Operations.Interactive {
		if len(args) < 2 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
		args = args[1:]
		if len(args) == 1 {
			args = strings.Split(args[0], " ")
		}
	}

	do.Operations.Args = args
	config.GlobalConfig.InteractiveWriter = os.Stdout
	config.GlobalConfig.InteractiveErrorWriter = os.Stderr
	_, err := data.ExecData(do)
	IfExit(err)
}
