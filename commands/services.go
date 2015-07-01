package commands

import (
	srv "github.com/eris-ltd/eris-cli/services"

	"github.com/eris-ltd/eris-cli/Godeps/_workspace/src/github.com/spf13/cobra"
)

// Primary Services Sub-Command
var Services = &cobra.Command{
	Use:   "services",
	Short: "Start, Stop, and Manage Services Required for your Application.",
	Long: `Start, Stop, and Manage Services Required for your Application.

Services are all services known and used by the Eris platform with the
exception of blockchain services. Blockchain services are managed and
operated via the [eris chain] command.`,
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
	Services.AddCommand(servicesExec)
	Services.AddCommand(servicesExport)
	Services.AddCommand(servicesStop)
	Services.AddCommand(servicesRename)
	Services.AddCommand(servicesUpdate)
	Services.AddCommand(servicesRm)
	addServicesFlags()
}

func addServicesFlags() {
	servicesLogs.Flags().BoolVarP(&Follow, "tail", "t", false, "follow logs, like tail -f")

	servicesExec.Flags().BoolVarP(&Interactive, "interactive", "i", false, "interactive shell")

	servicesUpdate.Flags().BoolVarP(&SkipPull, "skip-pull", "p", false, "skip the pulling feature and simply rebuild the service container")

	servicesStop.Flags().BoolVarP(&All, "all", "a", false, "stop the primary service and its dependent services")
	servicesStop.Flags().BoolVarP(&Rm, "rm", "r", false, "remove containers after stopping")

	servicesRm.Flags().BoolVarP(&Force, "force", "f", false, "remove service definition file from eris library")
}

// list-known lists the services which eris can automagically install
// flags to add: --list-versions
var servicesListKnown = &cobra.Command{
	Use:   "known",
	Short: "List all the services Eris knows about.",
	Long: `Lists the services which Eris has installed for you.

To install a new service, use: [eris services import].

Services include all executable services supported by the Eris platform which are
NOT blockchains or key managers.

Blockchains are handled using the [eris chains] command.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.ListKnown()
	},
}

// install a service
var servicesImport = &cobra.Command{
	Use:   "import [name] [location]",
	Short: "Import a service definition file from Github or IPFS.",
	Long: `Import a service for your platform.

By default, Eris will import from ipfs.

To list known services use: [eris services known].`,
	Example: "  eris services import eth ipfs:QmQ1LZYPNG4wSb9dojRicWCmM4gFLTPKFUhFnMTR3GKuA2",
	Run: func(cmd *cobra.Command, args []string) {
		srv.Import(cmd, args)
	},
}

// new
// flags to add: --type, --genesis, --config, --checkout, --force-name
var servicesNew = &cobra.Command{
	Use:   "new [name] [image]",
	Short: "Creates a new service.",
	Long: `Creates a new service.

Command must be given a name and a Container Image using standard
docker format of [repository/organization/image].`,
	Example: `  eris new eth eris/eth
  eris new mint tutum.co/tendermint/tendermint`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.New(cmd, args)
	},
}

// ls lists the services available locally
var servicesListExisting = &cobra.Command{
	Use:   "ls",
	Short: "List the installed and built services.",
	Long: `Lists the installed and built services which Eris knows about.

To list the known services: [eris services known]
To list the running services: [eris services ps]
To start a service use: [eris services start serviceName].`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.ListExisting()
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
		srv.Edit(cmd, args)
	},
}

// flags to add: --identity (alias for --key), --foreground
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
		srv.Start(cmd, args)
	},
}

// flags to add: --all (list installed)
var servicesListRunning = &cobra.Command{
	Use:   "ps",
	Short: "Lists the running services.",
	Long:  `Lists the services which are currently running.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.ListRunning()
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
		srv.Inspect(cmd, args)
	},
}

var servicesExec = &cobra.Command{
	Use:   "exec [serviceName]",
	Short: "Run a command or interactive shell",
	Long:  "Run a command or interactive shell in a container with volumes-from the data container",
	Run: func(cmd *cobra.Command, args []string) {
		srv.Exec(cmd, args)
	},
}

var servicesExport = &cobra.Command{
	Use:   "export [serviceName]",
	Short: "Export a service definition file to IPFS.",
	Long: `Export a service definition file to IPFS.

Command will return a machine readable version of the IPFS hash
`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.Export(cmd, args)
	},
}

// flags to add: --tail
var servicesLogs = &cobra.Command{
	Use:   "logs [name]",
	Short: "Displays the logs of a running service.",
	Long:  `Displays the logs of a running service.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.Logs(cmd, args)
	},
}

// stop stops a running service
var servicesStop = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stops a running service.",
	Long:  `Stops a service which is currently running.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.Kill(cmd, args)
	},
}

// renames an installed service
var servicesRename = &cobra.Command{
	Use:   "rename [oldName] [newName]",
	Short: "Renames an installed service.",
	Long:  `Renames an installed service.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.Rename(cmd, args)
	},
}

// updates an installed service
// falts to add: --no-pull
var servicesUpdate = &cobra.Command{
	Use:   "update [name]",
	Short: "Updates an installed service.",
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
		srv.Update(cmd, args)
	},
}

// rm [name] removes a service
// flags to add: --force (includes the service definition file)
var servicesRm = &cobra.Command{
	Use:   "rm [name]",
	Short: "Removes an installed service.",
	Long: `Removes an installed service.

Command will remove the service's container but will not
remove the service definition file.

Use the --force flag to also remove the service definition file.`,
	Run: func(cmd *cobra.Command, args []string) {
		srv.Rm(cmd, args)
	},
}
