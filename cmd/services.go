package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/list"
	srv "github.com/eris-ltd/eris-cli/services"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------
// cli definition

// Primary Services Sub-Command
var Services = &cobra.Command{
	Use:   "services",
	Short: "Start, stop, and manage services required for your application",
	Long: `Start, stop, and manage services required for your application.

Eris services are "things that you turn on or off". They are meant to be long
running microservices on which your application relies. They can be public
blockchains, services your application needs, workers, bridges to other data
or process management systems, or pretty much any process that has a docker
image.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// build the services subcommand
func buildServicesCommand() {
	Services.AddCommand(servicesNew)
	Services.AddCommand(servicesImport)
	Services.AddCommand(servicesListAll)
	Services.AddCommand(servicesEdit)
	Services.AddCommand(servicesStart)
	Services.AddCommand(servicesLogs)
	Services.AddCommand(servicesInspect)
	Services.AddCommand(servicesPorts)
	Services.AddCommand(servicesExec)
	Services.AddCommand(servicesStop)
	Services.AddCommand(servicesExport)
	Services.AddCommand(servicesRename)
	Services.AddCommand(servicesUpdate)
	Services.AddCommand(servicesRm)
	Services.AddCommand(servicesCat)
	addServicesFlags()
}

// Services Sub-sub-Commands

//lists all or specify flag
var servicesListAll = &cobra.Command{
	Use:   "ls",
	Short: "Lists everything service related.",
	Long: `Lists all: service definition files (--known), current existing containers
for each service (--existing), and current running containers
for each service (--running).

Known services can be started with the [eris services start NAME] command.
To install a new service, use [eris services import]. Services include
all executable services supported by the Eris platform which are
NOT blockchains or key managers.

Blockchains are handled using the [eris chains] command.`,
	Run: ListAllServices,
}

var servicesImport = &cobra.Command{
	Use:     "import NAME HASH",
	Short:   "Import a service definition file from IPFS.",
	Long:    `Import a service for your platform.`,
	Example: "$ eris services import eth QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2",
	Run:     ImportService,
}

var servicesNew = &cobra.Command{
	Use:   "new NAME IMAGE",
	Short: "Create a new service.",
	Long: `Create a new service.

Command must be given a NAME and a container IMAGE using the standard
docker format of [repository/organization/image].`,
	Example: "$ eris services new eth eris/eth\n" +
		"$ eris services new mint tutum.co/tendermint/tendermint",
	Run: NewService,
}

var servicesEdit = &cobra.Command{
	Use:   "edit NAME",
	Short: "Edit a service.",
	Long: `Edit a service definition file which is kept in ~/.eris/services.
Edit will utilize your default editor. (See also the ERIS environment variable.)

NOTE: Do not use this command for configuring a *specific* service. This
command will only operate on *service configuration file* which tell Eris
how to start and stop a specific service.

How that service is used for a specific project is handled from project
definition files.`,
	Run: EditService,
}

var servicesStart = &cobra.Command{
	Use:   "start NAME",
	Short: "Start a service.",
	Long: `Start a service according to the service definition file which
eris stores in the ~/.eris/services directory.

The [eris services start NAME] command by default will put the
service into the background so its logs will not be viewable
from the command line.

To stop the service use:      [eris services stop NAME].
To view a service's logs use: [eris services logs NAME].`,
	Run: StartService,
}

var servicesInspect = &cobra.Command{
	Use:   "inspect NAME [KEY]",
	Short: "Machine readable service operation details.",
	Long: `Display machine readable details about running containers.

Information available to the inspect command is provided by the Docker API.
For more information about return values, see:
https://github.com/fsouza/go-dockerclient/blob/master/container.go#L235`,
	Example: `$ eris services inspect ipfs -- will display the entire information about ipfs containers
$ eris services inspect ipfs name -- will display the name in machine readable format
$ eris services inspect ipfs host_config.binds -- will display only that value`,
	Run: InspectService,
}

var servicesPorts = &cobra.Command{
	Use:   "ports NAME [PORT]...",
	Short: "Print port mappings",
	Long: `Print port mappings.

The [eris services ports] command displays published service ports.`,
	Example: `$ eris services ports ipfs -- will display all IPFS ports
$ eris services ports ipfs 4001 5001 -- will display specific IPFS ports`,
	Run: PortsService,
}

var servicesExport = &cobra.Command{
	Use:   "export NAME",
	Short: "Export a service definition file to IPFS.",
	Long: `Export a service definition file to IPFS.

Command will return a machine readable version of the IPFS hash.`,
	Run: ExportService,
}

var servicesLogs = &cobra.Command{
	Use:   "logs NAME",
	Short: "Display the logs of a running service.",
	Long:  `Display the logs of a running service.`,
	Run:   LogService,
}

var servicesExec = &cobra.Command{
	Use:   "exec NAME",
	Short: "Run a command or interactive shell",
	Long:  "Run a command or interactive shell in a container with volumes-from the data container",
	Run:   ExecService,
}

// stop stops a running service
var servicesStop = &cobra.Command{
	Use:   "stop NAME",
	Short: "Stop a running service.",
	Long:  `Stop a service which is currently running.`,
	Run:   KillService,
}

var servicesRename = &cobra.Command{
	Use:   "rename OLD_NAME NEW_NAME",
	Short: "Rename an installed service.",
	Long:  `Rename an installed service.`,
	Run:   RenameService,
}

var servicesUpdate = &cobra.Command{
	Use:     "update NAME",
	Aliases: []string{"restart"},
	Short:   "Update an installed service.",
	Long: `Update an installed service, or install it if it has not been installed.

Functionally this command will perform the following sequence of steps:

1. Stop the service (if it is running).
2. Remove the container which ran the service.
3. Pull the image the container uses from a hub.
4. Rebuild the container from the updated image.
5. Restart the service (if it was previously running).

NOTE: If the service uses data containers, those will not be affected
by the [eris update] command.`,
	Run: UpdateService,
}

var servicesRm = &cobra.Command{
	Use:   "rm NAME",
	Short: "Remove an installed service.",
	Long: `Remove an installed service.

Command will remove the service's container but will not remove
the service definition file.`,
	Run: RmService,
}

var servicesCat = &cobra.Command{
	Use:   "cat NAME",
	Short: "Display the service definition file.",
	Long: `Display the service definition file.

Command will cat local service definition file.`,
	Run: CatService,
}

//----------------------------------------------------------------------
// cli flags

func addServicesFlags() {
	buildFlag(servicesLogs, do, "follow", "service")
	buildFlag(servicesLogs, do, "tail", "service")

	buildFlag(servicesExec, do, "env", "service")
	buildFlag(servicesExec, do, "links", "service")
	servicesExec.Flags().StringVarP(&do.Operations.Volume, "volume", "", "", fmt.Sprintf("mount a volume %v/VOLUME on a host machine to a %v/VOLUME on a container", ErisRoot, ErisContainerRoot))
	buildFlag(servicesExec, do, "publish", "service")
	buildFlag(servicesExec, do, "interactive", "service")

	buildFlag(servicesUpdate, do, "pull", "service")
	buildFlag(servicesUpdate, do, "timeout", "service")
	buildFlag(servicesUpdate, do, "env", "service")
	buildFlag(servicesUpdate, do, "links", "service")

	buildFlag(servicesRm, do, "force", "service")
	buildFlag(servicesRm, do, "file", "service")
	buildFlag(servicesRm, do, "data", "service")
	buildFlag(servicesRm, do, "rm-volumes", "service")

	buildFlag(servicesStart, do, "publish", "service")
	buildFlag(servicesStart, do, "env", "service")
	buildFlag(servicesStart, do, "links", "service")
	buildFlag(servicesStart, do, "chain", "service")

	buildFlag(servicesStop, do, "rm", "service")
	buildFlag(servicesStop, do, "volumes", "service")
	buildFlag(servicesStop, do, "data", "service")
	buildFlag(servicesStop, do, "force", "service")
	buildFlag(servicesStop, do, "timeout", "service")
	servicesStop.Flags().BoolVarP(&do.All, "all", "a", false, "stop the primary service and its dependent services")
	servicesStop.Flags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain the service should also stop")

	buildFlag(servicesListAll, do, "known", "service")
	buildFlag(servicesListAll, do, "existing", "service")
	buildFlag(servicesListAll, do, "running", "service")
	buildFlag(servicesListAll, do, "quiet", "service")

}

//----------------------------------------------------------------------
// cli command wrappers

func StartService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	IfExit(srv.StartService(do))
}

func LogService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(srv.LogsService(do))
}

func ExecService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	args = args[1:]
	if !do.Operations.Interactive {
		if len(args) == 0 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
	}
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	do.Operations.Args = args
	config.GlobalConfig.InteractiveWriter = os.Stdout
	config.GlobalConfig.InteractiveErrorWriter = os.Stderr
	_, err := srv.ExecService(do)
	IfExit(err)
}

func KillService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	IfExit(srv.KillService(do))
}

// install
func ImportService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.Hash = args[1]
	IfExit(srv.ImportService(do))
}

func NewService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = []string{args[1]}
	IfExit(srv.NewService(do))
}

func EditService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(srv.EditService(do))
}

func RenameService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.NewName = args[1]
	IfExit(srv.RenameService(do))
}

func InspectService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	if len(args) == 1 {
		do.Operations.Args = []string{"all"}
	} else {
		do.Operations.Args = []string{args[1]}
	}

	IfExit(srv.InspectService(do))
}

func PortsService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = args[1:]
	IfExit(srv.PortsService(do))
}

func ExportService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(srv.ExportService(do))
}

// Updates an installed service, or installs it if it has not been installed.
func UpdateService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(srv.UpdateService(do))
}

func ListAllServices(cmd *cobra.Command, args []string) {
	//if no flags are set, list all the things
	//otherwise, allow only a single flag
	if !(do.Known || do.Running || do.Existing) {
		do.All = true
	} else {
		flargs := []bool{do.Known, do.Running, do.Existing}
		flags := []string{}

		for _, f := range flargs {
			if f == true {
				flags = append(flags, "true")
			}
		}
		IfExit(FlagCheck(1, "eq", cmd, flags))
	}

	if err := list.ListAll(do, "services"); err != nil {
		return
	}
	if !do.All { //do.All will output a pretty table on its own
		fmt.Println(do.Result)
	}
}

func RmService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	IfExit(srv.RmService(do))
}

func CatService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(srv.CatService(do))
}
