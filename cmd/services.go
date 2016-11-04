package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/eris-ltd/eris-cli/config"
	"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/list"
	"github.com/eris-ltd/eris-cli/services"
	"github.com/eris-ltd/eris-cli/util"

	"github.com/spf13/cobra"
)

var Services = &cobra.Command{
	Use:   "services",
	Short: "start, stop, and manage services required for your application",
	Long: `start, stop, and manage services required for your application

Eris services are "things that you turn on or off". They are meant to be long
running microservices on which your application relies. They can be public
blockchains, services your application needs, workers, bridges to other data
or process management systems, or pretty much any process that has a docker
image.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildServicesCommand() {
	Services.AddCommand(servicesMake)
	Services.AddCommand(servicesList)
	Services.AddCommand(servicesEdit)
	Services.AddCommand(servicesStart)
	Services.AddCommand(servicesLogs)
	Services.AddCommand(servicesInspect)
	Services.AddCommand(servicesIP)
	Services.AddCommand(servicesPorts)
	Services.AddCommand(servicesExec)
	Services.AddCommand(servicesStop)
	Services.AddCommand(servicesRename)
	Services.AddCommand(servicesUpdate)
	Services.AddCommand(servicesRm)
	Services.AddCommand(servicesCat)
	addServicesFlags()
}

var servicesList = &cobra.Command{
	Use:   "ls",
	Short: "lists everything service related",
	Long: `list services or known service definition files

The -r flag limits the output to running services only.

The --json flag dumps the container or known files information
in the JSON format.

The -q flag is equivalent to the '{{.ShortName}}' format.

The -f flag specifies an alternate format for the list, using the syntax
of Go text templates. See the more detailed description in the help
output for the [eris ls] command. The struct passed to the Go template
for the -k flag is this

  type Definition struct {
    Name       string       // service name
    Definition string       // definition file name
  }

The -k flag displays the known definition files.`,
	Run: ListServices,
	Example: `$ eris services ls -f '{{.ShortName}}\t{{.Info.Config.Cmd}}\t{{.Info.Config.Entrypoint}}'
$ eris services ls -f '{{.ShortName}}\t{{.Info.Config.Image}}\t{{ports .Info}}'
$ eris services ls -f '{{.ShortName}}\t{{.Info.Config.Volumes}}\t{{.Info.Config.Mounts}}'
$ eris services ls -f '{{.Info.ID}}\t{{.Info.HostConfig.VolumesFrom}}'`,
}

var servicesMake = &cobra.Command{
	Use:   "make NAME IMAGE",
	Short: "create a new service",
	Long: `create a new service

Command must be given a NAME and a container IMAGE using the standard
docker format of [repository/organization/image].`,
	Example: "$ eris services make eth eris/eth\n" +
		"$ eris services make mint tutum.co/tendermint/tendermint",
	Run: MakeService,
}

var servicesEdit = &cobra.Command{
	Use:   "edit NAME",
	Short: "edit a service",
	Long: `edit a service definition file which is kept in ` + util.Tilde(config.ServicesPath) + `.
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
	Short: "start a service",
	Long: `start a service according to the service definition file which
eris stores in the ` + util.Tilde(config.ServicesPath) + `directory

The [eris services start NAME] command by default will put the
service into the background so its logs will not be viewable
from the command line.

To stop the service use:      [eris services stop NAME].
To view a service's logs use: [eris services logs NAME].

You can redefine service ports accessible over the network with
the --ports flag.
`,
	Run: StartService,

	Example: `$ eris services start ipfs --ports 17000 -- map the first port from the definition file to the host port 17000
$ eris services start ipfs --ports 17000,18000- -- redefine the first and the second port mappings and autoincrement the rest
$ eris services start ipfs --ports 50000:5001 -- redefine the specific port mapping (published host port:exposed container port)`,
}

var servicesInspect = &cobra.Command{
	Use:   "inspect NAME [KEY]",
	Short: "machine readable service operation details",
	Long: `display machine readable details about running containers

Information available to the inspect command is provided by the Docker API.
For more information about return values, see:
https://github.com/fsouza/go-dockerclient/blob/master/container.go#L235`,
	Example: `$ eris services inspect ipfs -- will display the entire information about ipfs containers
$ eris services inspect ipfs name -- will display the name in machine readable format
$ eris services inspect ipfs host_config.binds -- will display only that value`,
	Run: InspectService,
}

var servicesIP = &cobra.Command{
	Use:   "ip NAME",
	Short: "display service IP",
	Long:  `display service IP`,

	Run: IPService,
}

var servicesPorts = &cobra.Command{
	Use:   "ports NAME [PORT]...",
	Short: "print port mappings",
	Long: `print port mappings

The [eris services ports] command displays published service ports.`,
	Example: `$ eris services ports ipfs -- will display all IPFS ports
$ eris services ports ipfs 4001 5001 -- will display specific IPFS ports`,
	Run: PortsService,
}

var servicesLogs = &cobra.Command{
	Use:   "logs NAME",
	Short: "display the logs of a running service",
	Long:  `display the logs of a running service`,
	Run:   LogService,
}

var servicesExec = &cobra.Command{
	Use:   "exec NAME",
	Short: "run a command or interactive shell",
	Long:  "run a command or interactive shell in a container with volumes-from the data container",
	Run:   ExecService,
}

var servicesStop = &cobra.Command{
	Use:   "stop NAME",
	Short: "stop a running service",
	Long:  `stop a service which is currently running`,
	Run:   KillService,
}

var servicesRename = &cobra.Command{
	Use:   "rename OLD_NAME NEW_NAME",
	Short: "rename an installed service",
	Long:  `rename an installed service`,
	Run:   RenameService,
}

var servicesUpdate = &cobra.Command{
	Use:     "update NAME",
	Aliases: []string{"restart"},
	Short:   "update an installed service",
	Long: `update an installed service, or install it if it has not been installed

Functionally this command will perform the following sequence of steps:

1. Stop the service (if it is running).
2. Remove the container which ran the service.
3. Pull the image the container uses from a hub.
4. Rebuild the container from the updated image.
5. Restart the service (if it was previously running).`,
	Run: UpdateService,
}

var servicesRm = &cobra.Command{
	Use:   "rm NAME",
	Short: "remove an installed service",
	Long: `remove an installed service

Command will remove the service's container but will not remove
the service definition file.`,
	Run: RmService,
}

var servicesCat = &cobra.Command{
	Use:   "cat NAME",
	Short: "display the service definition file",
	Long: `display the service definition file

Command will cat local service definition file.`,
	Run: CatService,
}

func addServicesFlags() {
	buildFlag(servicesLogs, do, "follow", "service")
	buildFlag(servicesLogs, do, "tail", "service")

	buildFlag(servicesExec, do, "env", "service")
	buildFlag(servicesExec, do, "links", "service")
	servicesExec.Flags().StringVarP(&do.Operations.Volume, "volume", "", "", fmt.Sprintf("mount a DIR or a VOLUME to a %v/DIR inside a container", util.Tilde(config.ErisRoot)))
	buildFlag(servicesExec, do, "publish", "service")
	buildFlag(servicesExec, do, "ports", "service")
	buildFlag(servicesExec, do, "interactive", "service")

	buildFlag(servicesUpdate, do, "pull", "service")
	buildFlag(servicesUpdate, do, "timeout", "service")
	buildFlag(servicesUpdate, do, "env", "service")
	buildFlag(servicesUpdate, do, "links", "service")

	buildFlag(servicesRm, do, "force", "service")
	buildFlag(servicesRm, do, "file", "service")
	buildFlag(servicesRm, do, "data", "service")
	buildFlag(servicesRm, do, "rm-volumes", "service")
	servicesRm.Flags().BoolVarP(&do.RmImage, "image", "", false, "remove the services' docker image")

	buildFlag(servicesStart, do, "publish", "service")
	buildFlag(servicesStart, do, "ports", "service")
	buildFlag(servicesStart, do, "env", "service")
	buildFlag(servicesStart, do, "links", "service")
	servicesStart.Flags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain the service depends on")

	buildFlag(servicesStop, do, "rm", "service")
	buildFlag(servicesStop, do, "volumes", "service")
	buildFlag(servicesStop, do, "data", "service")
	buildFlag(servicesStop, do, "force", "service")
	buildFlag(servicesStop, do, "timeout", "service")
	servicesStop.Flags().BoolVarP(&do.All, "all", "a", false, "stop the primary service and its dependent services")
	servicesStop.Flags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain the service should also stop")

	buildFlag(servicesList, do, "known", "service")
	servicesList.Flags().BoolVarP(&do.JSON, "json", "", false, "machine readable output")
	servicesList.Flags().BoolVarP(&do.All, "all", "a", false, "show extended output")
	servicesList.Flags().BoolVarP(&do.Running, "running", "r", false, "show running containers only")
	servicesList.Flags().BoolVarP(&do.Quiet, "quiet", "q", false, "show a list of service names")
	servicesList.Flags().StringVarP(&do.Format, "format", "f", "", "alternate format for columnized output")
}

func StartService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	util.IfExit(services.StartService(do))
}

func LogService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(services.LogsService(do))
}

func ExecService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	args = args[1:]
	if !do.Operations.Interactive {
		if len(args) == 0 {
			util.Exit(fmt.Errorf("Non-interactive exec sessions must provide arguments to execute"))
		}
	}
	if len(args) == 1 {
		args = strings.Split(args[0], " ")
	}
	do.Operations.Terminal = true
	do.Operations.Args = args
	config.Global.InteractiveWriter = os.Stdout
	config.Global.InteractiveErrorWriter = os.Stderr
	_, err := services.ExecService(do)
	util.IfExit(err)
}

func KillService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	util.IfExit(services.KillService(do))
}

func MakeService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = []string{args[1]}
	util.IfExit(services.MakeService(do))
}

func EditService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(services.EditService(do))
}

func RenameService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(2, "ge", cmd, args))
	do.Name = args[0]
	do.NewName = args[1]
	util.IfExit(services.RenameService(do))
}

func InspectService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))

	do.Name = args[0]
	if len(args) == 1 {
		do.Operations.Args = []string{"all"}
	} else {
		do.Operations.Args = []string{args[1]}
	}

	util.IfExit(services.InspectService(do))
}

func IPService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = []string{"NetworkSettings.IPAddress"}
	util.IfExit(services.InspectService(do))
}

func PortsService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = args[1:]
	util.IfExit(services.PortsService(do))
}

func UpdateService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	util.IfExit(services.UpdateService(do))
}

func ListServices(cmd *cobra.Command, args []string) {
	if do.All {
		do.Format = "extended"
	}
	if do.Quiet {
		do.Format = "{{.ShortName}}"
	}
	if do.JSON {
		do.Format = "json"
	}
	if do.Known {
		util.IfExit(list.Known("services", do.Format))
	} else {
		util.IfExit(list.Containers(definitions.TypeService, do.Format, do.Running))
	}
}

func RmService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	util.IfExit(services.RmService(do))
}

func CatService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	out, err := services.CatService(do)
	util.IfExit(err)
	fmt.Fprint(config.Global.Writer, out)
}
