package commands

import (
	"fmt"

	srv "github.com/eris-ltd/eris-cli/services"

	. "github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/eris-ltd/common"
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
	Services.AddCommand(servicesListKnown)
	Services.AddCommand(servicesListExisting)
	Services.AddCommand(servicesEdit)
	Services.AddCommand(servicesStart)
	Services.AddCommand(servicesLogs)
	Services.AddCommand(servicesListRunning)
	Services.AddCommand(servicesInspect)
	Services.AddCommand(servicesStop)
	Services.AddCommand(servicesExport)
	Services.AddCommand(servicesRename)
	Services.AddCommand(servicesUpdate)
	Services.AddCommand(servicesRm)
	Services.AddCommand(servicesCat)
	addServicesFlags()
}

// Services Sub-sub-Commands
var servicesListKnown = &cobra.Command{
	Use:   "known",
	Short: "List all the services Eris knows about.",
	Long: `Lists the services which Eris has installed for you.

To install a new service, use: [eris services import].

Services include all executable services supported by the Eris platform which are
NOT blockchains or key managers.

Blockchains are handled using the [eris chains] command.`,
	Run: func(cmd *cobra.Command, args []string) {
		ListKnownServices()
	},
}

var servicesImport = &cobra.Command{
	Use:   "import [name] [location]",
	Short: "Import a service definition file from Github or IPFS.",
	Long: `Import a service for your platform.

By default, Eris will import from ipfs.

To list known services use: [eris services known].`,
	Example: "  eris services import eth ipfs:QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2",
	Run: func(cmd *cobra.Command, args []string) {
		ImportService(cmd, args)
	},
}

var servicesNew = &cobra.Command{
	Use:   "new [name] [image]",
	Short: "Creates a new service.",
	Long: `Creates a new service.

Command must be given a name and a Container Image using standard
docker format of [repository/organization/image].`,
	Example: `  eris new eth eris/eth
  eris new mint tutum.co/tendermint/tendermint`,
	Run: func(cmd *cobra.Command, args []string) {
		NewService(cmd, args)
	},
}

var servicesListExisting = &cobra.Command{
	Use:   "ls",
	Short: "List the installed and built services.",
	Long: `Lists the installed and built services which Eris knows about.

To list the known services: [eris services known]
To list the running services: [eris services ps]
To start a service use: [eris services start serviceName].`,
	Run: func(cmd *cobra.Command, args []string) {
		ListExistingServices()
	},
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
	Run: func(cmd *cobra.Command, args []string) {
		EditService(cmd, args)
	},
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
	Run: func(cmd *cobra.Command, args []string) {
		StartService(cmd, args)
	},
}

var servicesListRunning = &cobra.Command{
	Use:   "ps",
	Short: "Lists the running services.",
	Long:  `Lists the services which are currently running.`,
	Run: func(cmd *cobra.Command, args []string) {
		ListRunningServices()
	},
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
	Run: func(cmd *cobra.Command, args []string) {
		InspectService(cmd, args)
	},
}

var servicesExport = &cobra.Command{
	Use:   "export [serviceName]",
	Short: "Export a service definition file to IPFS.",
	Long: `Export a service definition file to IPFS.

Command will return a machine readable version of the IPFS hash
`,
	Run: func(cmd *cobra.Command, args []string) {
		ExportService(cmd, args)
	},
}

var servicesLogs = &cobra.Command{
	Use:   "logs [name]",
	Short: "Displays the logs of a running service.",
	Long:  `Displays the logs of a running service.`,
	Run: func(cmd *cobra.Command, args []string) {
		LogService(cmd, args)
	},
}

// stop stops a running service
var servicesStop = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stops a running service.",
	Long:  `Stops a service which is currently running.`,
	Run: func(cmd *cobra.Command, args []string) {
		KillService(cmd, args)
	},
}

var servicesRename = &cobra.Command{
	Use:   "rename [oldName] [newName]",
	Short: "Renames an installed service.",
	Long:  `Renames an installed service.`,
	Run: func(cmd *cobra.Command, args []string) {
		RenameService(cmd, args)
	},
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
	Run: func(cmd *cobra.Command, args []string) {
		UpdateService(cmd, args)
	},
}

var servicesRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Removes an installed service.",
	Long: `Removes an installed service.

Command will remove the service's container but will not
remove the service definition file.

Use the --force flag to also remove the service definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		RmService(cmd, args)
	},
}

var servicesCat = &cobra.Command{
	Use:   "cat [name]",
	Short: "Displays service definition file.",
	Long: `Displays service definition file.

Command will cat local service definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		CatService(cmd, args)
	},
}

//----------------------------------------------------------------------
// cli flags

func addServicesFlags() {
	servicesLogs.Flags().BoolVarP(&do.Follow, "follow", "f", false, "follow logs")
	servicesLogs.Flags().StringVarP(&do.Tail, "tail", "t", "all", "number of lines to show from end of logs")

	servicesUpdate.Flags().BoolVarP(&do.Pull, "pull", "p", false, "skip the pulling feature and simply rebuild the service container")
	servicesUpdate.Flags().UintVarP(&do.Timeout, "timeout", "t", 10, "manually set the timeout; overridden by --force")

	servicesStart.Flags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain the service depends on")

	servicesStop.Flags().BoolVarP(&do.All, "all", "a", false, "stop the primary service and its dependent services")
	servicesStop.Flags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain the service should also stop")
	servicesStop.Flags().BoolVarP(&do.Rm, "rm", "r", false, "remove containers after stopping")
	servicesStop.Flags().BoolVarP(&do.RmD, "data", "x", false, "remove data containers after stopping")
	servicesStop.Flags().BoolVarP(&do.Force, "force", "f", false, "kill the container instantly without waiting to exit")
	servicesStop.Flags().UintVarP(&do.Timeout, "timeout", "t", 10, "manually set the timeout; overridden by --force")

	servicesRm.Flags().BoolVarP(&do.File, "file", "f", false, "remove service definition file as well as service container")
	servicesRm.Flags().BoolVarP(&do.RmD, "data", "x", false, "remove data containers as well")

	servicesListExisting.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "machine parsable output")
	servicesListRunning.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "machine parsable output")
}

//----------------------------------------------------------------------
// cli command wrappers

func StartService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Args = args
	IfExit(srv.StartService(do))
}

func LogService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(srv.LogsService(do))
}

func KillService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}

	do.Args = args
	IfExit(srv.KillService(do))
}

// install
func ImportService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	do.Path = args[1]
	IfExit(srv.ImportService(do))
}

func NewService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	do.Args = []string{args[1]}
	IfExit(srv.NewService(do))
}

func EditService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(srv.EditService(do))
}

func RenameService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) != 2 {
		cmd.Help()
		return
	}
	do.Name = args[0]
	do.NewName = args[1]
	IfExit(srv.RenameService(do))
}

func InspectService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	if len(args) == 1 {
		args = append(args, "all")
	}
	do.Name = args[0]
	do.Args = []string{args[1]}
	IfExit(srv.InspectService(do))
}

func ExportService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(srv.ExportService(do))
}

// Updates an installed service, or installs it if it has not been installed.
func UpdateService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(srv.UpdateService(do))
}

// list known
func ListKnownServices() {
	if err := srv.ListKnown(do); err != nil {
		return
	}

	fmt.Println(do.Result)
}

func ListRunningServices() {
	if err := srv.ListRunning(do); err != nil {
		return
	}
}

func ListExistingServices() {
	if err := srv.ListExisting(do); err != nil {
		return
	}
}

func RmService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Args = args
	IfExit(srv.RmService(do))
}

func CatService(cmd *cobra.Command, args []string) {
	if err := checkServiceGiven(args); err != nil {
		cmd.Help()
		return
	}
	do.Name = args[0]
	IfExit(srv.CatService(do))
}

func checkServiceGiven(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("No Service Given. Please rerun command with a known service.")
	}
	return nil
}
