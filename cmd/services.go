package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/monax/monax/config"
	"github.com/monax/monax/services"
	"github.com/monax/monax/util"

	"github.com/spf13/cobra"
)

var Services = &cobra.Command{
	Use:   "services",
	Short: "start, stop, and manage services required for your application",
	Long: `start, stop, and manage services required for your application

Monax services are "things that you turn on or off". They are meant to be long
running microservices on which your application relies. They can be public
blockchains, services your application needs, workers, bridges to other data
or process management systems, or pretty much any process that has a docker
image.`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

func buildServicesCommand() {
	Services.AddCommand(servicesStart)
	Services.AddCommand(servicesLogs)
	Services.AddCommand(servicesIP)
	Services.AddCommand(servicesExec)
	Services.AddCommand(servicesStop)
	Services.AddCommand(servicesRm)
	addServicesFlags()
}

var servicesStart = &cobra.Command{
	Use:   "start NAME",
	Short: "start a service",
	Long: `start a service according to the service definition file which
monax stores in the ` + util.Tilde(config.ServicesPath) + `directory

The [monax services start NAME] command by default will put the
service into the background so its logs will not be viewable
from the command line.

To stop the service use:      [monax services stop NAME].
To view a service's logs use: [monax services logs NAME].

You can redefine service ports accessible over the network with
the --ports flag.
`,
	Run: StartService,

	Example: `$ monax services start ipfs --ports 17000 -- map the first port from the definition file to the host port 17000
$ monax services start ipfs --ports 17000,18000- -- redefine the first and the second port mappings and autoincrement the rest
$ monax services start ipfs --ports 50000:5001 -- redefine the specific port mapping (published host port:exposed container port)`,
}

var servicesIP = &cobra.Command{
	Use:   "ip NAME",
	Short: "display service IP",
	Long:  `display service IP`,

	Run: IPService,
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

var servicesRm = &cobra.Command{
	Use:   "rm NAME",
	Short: "remove an installed service",
	Long: `remove an installed service

Command will remove the service's container but will not remove
the service definition file.`,
	Run: RmService,
}

func addServicesFlags() {
	buildFlag(servicesLogs, do, "follow", "service")
	buildFlag(servicesLogs, do, "tail", "service")

	buildFlag(servicesExec, do, "env", "service")
	buildFlag(servicesExec, do, "links", "service")
	servicesExec.Flags().StringVarP(&do.Operations.Volume, "volume", "", "", fmt.Sprintf("mount a DIR or a VOLUME to a %v/DIR inside a container", util.Tilde(config.MonaxRoot)))
	buildFlag(servicesExec, do, "publish", "service")
	buildFlag(servicesExec, do, "ports", "service")
	buildFlag(servicesExec, do, "interactive", "service")

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

func IPService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Name = args[0]
	do.Operations.Args = []string{"NetworkSettings.IPAddress"}
	util.IfExit(services.InspectService(do))
}

func RmService(cmd *cobra.Command, args []string) {
	util.IfExit(ArgCheck(1, "ge", cmd, args))
	do.Operations.Args = args
	util.IfExit(services.RmService(do))
}
