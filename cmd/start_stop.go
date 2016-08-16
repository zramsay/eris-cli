package commands

import (
	//"github.com/eris-ltd/eris-cli/definitions"
	"github.com/eris-ltd/eris-cli/kaihei"

	"github.com/spf13/cobra"

	. "github.com/eris-ltd/common/go/common"
)

var Start = &cobra.Command{
	Use:   "startup",
	Short: "start up your Eris dev environment for the day",
	Long: `start up your Eris dev environment for the day

The startup command starts the services for which you have started
in the past. The startup command also gives you the option to 
specify a '--chain' flag that specifies the chain you would 
like to start with the services.`,
	Run: StartEris,
}

var Stop = &cobra.Command{	
	Use:   "shutdown",
	Short: "shutdown your Eris dev environment for the day",
	Long: `shutdown your Eris dev environment for the day

The shutdown command finds all of the services and chains that 
are currently running and shuts them down for the day. In effect,
this command wraps:
[eris services stop $(eris services ls -q)] and
[eris chains stop $(eris chains ls -q)]`,
	Run: StopEris,
}

func buildStartCommand() {
	addStartFlags()
}

func addStartFlags() {
	Start.PersistentFlags().StringVarP(&do.ChainName, "chain", "c", "", "specify a chain to start")
}

func buildStopCommand() {
	addStopFlags()
}

func addStopFlags() {
}

func StartEris(cmd *cobra.Command, args []string) {
	IfExit(kaihei.StartUpEris(do))
}

func StopEris(cmd *cobra.Command, args []string) {
	IfExit(kaihei.ShutUpEris(do))
}
