package commands

import (
	"fmt"
	"strings"

	srv "github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common/go/common"
	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

//----------------------------------------------------------------------
// cli definition

// Primary Services Sub-Command
var Services = &cobra.Command{
	Use:   "services",
	Short: "Start, Stop, and Manage Services Required for your Application.",
	Long: `Start, Stop, and Manage Services Required for your Application.

Services are all services known and used by the Eris platform with the
exception of blockchain services.`,
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
	Long: `Lists all:

	- service definition files (--known)
	- current existing containers for each service (--existing)
	- current running containers for each service (--running)

Known services can be started with:

	$ eris services start serviceName

To install a new service, use: 

	$ eris services import

Services include all executable services supported by the 
Eris platform which are NOT blockchains or key managers.

Blockchains are handled using the [eris chains] command.`,
	Run: ListAllServices,
}

var servicesImport = &cobra.Command{
	Use:   "import [name] [hash]",
	Short: "Import a service definition file from IPFS.",
	Long: `Import a service for your platform.


To list known services use: [eris services known].`,
	Example: "  eris services import eth QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2",
	Run:     ImportService,
}

var servicesNew = &cobra.Command{
	Use:   "new [name] [image]",
	Short: "Creates a new service.",
	Long: `Creates a new service.

Command must be given a name and a Container Image using standard
docker format of [repository/organization/image].`,
	Example: `  eris services new eth eris/eth
  eris services new mint tutum.co/tendermint/tendermint`,
	Run: NewService,
}

var servicesEdit = &cobra.Command{
	Use:   "edit [name]",
	Short: "Edit a service.",
	Long: `Edit a service definition file which is kept in ~/.eris/services.

Edit will utilize your default editor.

NOTE: Do not use this command for configuring a *specific* service. This
command will only operate on *service configuration file* which tell Eris
how to start and stop a specific service.

How that service is used for a specific project is handled from project
definition files.

For more information on project definition files please see: [eris help projects].`,
	Run: EditService,
}

var servicesStart = &cobra.Command{
	Use:   "start [name]",
	Short: "Start a service.",
	Long: `Starts a service according to the service definition file which
eris stores in the ~/.eris/services directory.

[eris services start name] by default will put the service into the
background so its logs will not be viewable from the command line.

To stop the service use:      [eris services stop serviceName].
To view a service's logs use: [eris services logs serviceName].`,
	Run: StartService,
}

var servicesInspect = &cobra.Command{
	Use:   "inspect [serviceName] [key]",
	Short: "Machine readable service operation details.",
	Long: `Displays machine readable details about running containers.

Information available to the inspect command is provided by the
Docker API. For more information about return values,
see: https://github.com/fsouza/go-dockerclient/blob/master/container.go#L235`,
	Example: `  eris services inspect ipfs -> will display the entire information about ipfs containers
  eris services inspect ipfs name -> will display the name in machine readable format
  eris services inspect ipfs host_config.binds -> will display only that value`,
	Run: InspectService,
}

var servicesPorts = &cobra.Command{
	Use:   "ports [port]",
	Short: "Print port mapping",
	Long:  "Print port mapping",
	Run:   PortsService,
}

var servicesExport = &cobra.Command{
	Use:   "export [serviceName]",
	Short: "Export a service definition file to IPFS.",
	Long: `Export a service definition file to IPFS.

Command will return a machine readable version of the IPFS hash
`,
	Run: ExportService,
}

var servicesLogs = &cobra.Command{
	Use:   "logs [name]",
	Short: "Displays the logs of a running service.",
	Long:  `Displays the logs of a running service.`,
	Run:   LogService,
}

var servicesExec = &cobra.Command{
	Use:   "exec [serviceName]",
	Short: "Run a command or interactive shell",
	Long:  "Run a command or interactive shell in a container with volumes-from the data container",
	Run:   ExecService,
}

// stop stops a running service
var servicesStop = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stops a running service.",
	Long:  `Stops a service which is currently running.`,
	Run:   KillService,
}

var servicesRename = &cobra.Command{
	Use:   "rename [oldName] [newName]",
	Short: "Renames an installed service.",
	Long:  `Renames an installed service.`,
	Run:   RenameService,
}

var servicesUpdate = &cobra.Command{
	Use:     "update [name]",
	Aliases: []string{"restart"},
	Short:   "Updates an installed service.",
	Long: `Updates an installed service, or installs it if it has not been installed.

Functionally this command will perform the following sequence:

1. Stop the service (if it is running)
2. Remove the container which ran the service
3. Pull the image the container uses from a hub
4. Rebuild the container from the updated image
5. Restart the service (if it was previously running)

**NOTE**: If the service uses data containers those will not be affected
by the update command.`,
	Run: UpdateService,
}

var servicesRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Removes an installed service.",
	Long: `Removes an installed service.

Command will remove the service's container but will not
remove the service definition file.

Use the --force flag to also remove the service definition file.`,
	Run: RmService,
}

var servicesCat = &cobra.Command{
	Use:   "cat [name]",
	Short: "Displays service definition file.",
	Long: `Displays service definition file.

Command will cat local service definition file.`,
	Run: CatService,
}

//----------------------------------------------------------------------
// cli flags

func addServicesFlags() {
	servicesLogs.Flags().BoolVarP(&do.Follow, "follow", "f", false, "follow logs")
	servicesLogs.Flags().StringVarP(&do.Tail, "tail", "t", "150", "number of lines to show from end of logs")

	servicesExec.Flags().BoolVarP(&do.Interactive, "interactive", "i", false, "interactive shell")
	servicesExec.Flags().StringVarP(&do.Volume, "volume", "m", "", "mount a volume $HOME/.eris/VOLUME on a host machine to a /home/eris/.eris/VOLUME on a container")

	servicesUpdate.Flags().BoolVarP(&do.Pull, "pull", "p", false, "skip the pulling feature and simply rebuild the service container")
	servicesUpdate.Flags().UintVarP(&do.Timeout, "timeout", "t", 10, "manually set the timeout; overridden by --force")
	servicesUpdate.PersistentFlags().StringSliceVarP(&do.Env, "env", "e", nil, "multiple env vars can be passed using the KEY1=val1,KEY2=val1 syntax")
	servicesUpdate.PersistentFlags().StringSliceVarP(&do.Links, "links", "l", nil, "multiple containers can be linked can be passed using the KEY1:val1,KEY2:val1 syntax")

	servicesStart.PersistentFlags().BoolVarP(&do.Operations.PublishAllPorts, "publish", "p", false, "publish random ports")
	servicesStart.Flags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain the service depends on")
	servicesStart.PersistentFlags().StringSliceVarP(&do.Env, "env", "e", nil, "multiple env vars can be passed using the KEY1=val1,KEY2=val1 syntax")
	servicesStart.PersistentFlags().StringSliceVarP(&do.Links, "links", "l", nil, "multiple containers can be linked can be passed using the KEY1:val1,KEY2:val1 syntax")

	servicesStop.Flags().BoolVarP(&do.All, "all", "a", false, "stop the primary service and its dependent services")
	servicesStop.Flags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain the service should also stop")
	servicesStop.Flags().BoolVarP(&do.Rm, "rm", "r", false, "remove containers after stopping")
	servicesStop.Flags().BoolVarP(&do.RmD, "data", "x", false, "remove data containers after stopping")
	servicesStop.Flags().BoolVarP(&do.Force, "force", "f", false, "kill the container instantly without waiting to exit")
	servicesStop.Flags().UintVarP(&do.Timeout, "timeout", "t", 10, "manually set the timeout; overridden by --force")

	servicesRm.Flags().BoolVarP(&do.File, "file", "f", false, "remove service definition file as well as service container")
	servicesRm.Flags().BoolVarP(&do.RmD, "data", "x", false, "remove data containers as well")

	servicesListAll.Flags().BoolVarP(&do.Known, "known", "k", false, "list all the service definition files that exist")
	servicesListAll.Flags().BoolVarP(&do.Running, "running", "r", false, "list all the current containers which are running for a service")
	servicesListAll.Flags().BoolVarP(&do.Existing, "existing", "e", false, "list all the all current containers which exist for a service")
	servicesListAll.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "machine parsable output")

}

//----------------------------------------------------------------------
// cli command wrappers

func StartService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Args = args
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
	// if interactive, we ignore args. if not, run args as command
	args = args[1:]
	if !do.Interactive {
		if len(args) == 0 {
			Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
	}
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	do.Args = args
	IfExit(srv.ExecService(do))
}

func KillService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Args = args
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
	do.Args = []string{args[1]}
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
		do.Args = []string{"all"}
	} else {
		do.Args = []string{args[1]}
	}

	IfExit(srv.InspectService(do))
}

func PortsService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.Args = args[1:]
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

	if err := util.ListAll(do, "services"); err != nil {
		return
	}
	if !do.All { //do.All will output a pretty table on its own
		fmt.Println(do.Result)
	}
}

func RmService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Args = args
	IfExit(srv.RmService(do))
}

func CatService(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	IfExit(srv.CatService(do))
}
