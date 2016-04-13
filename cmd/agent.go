package commands

import (
	"github.com/eris-ltd/eris-cli/agent"

	. "github.com/eris-ltd/common/go/common"

	"github.com/spf13/cobra"
)

// Primary Agents Sub-Command
var Agents = &cobra.Command{
	Use:   "agents",
	Short: "Start, Stop, and Manage Agents.",
	Long: `Start, stop, and manage agents.
`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the agents subcommand
func buildAgentsCommand() {
	Agents.AddCommand(agentsStart)
	Agents.AddCommand(agentsStop)
	addAgentsFlags()
}

// start a agent
var agentsStart = &cobra.Command{
	Use:   "start",
	Short: "Start a agent registered with Eris.",
	Long: `Start a agent registered with Eris. If no is give Eris
will simply start the currently checked out agent. To stop a
agent use: [eris agents kill name].`,
	Run: StartAgent,
}

// stop a running agent
var agentsStop = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running agent.",
	Long: `Stop a running agent. If no is give Eris
will simply stop the currently checked out agent.`,
	Run: StopAgent,
}

//----------------------------------------------------------------------
// cli flags
func addAgentsFlags() {
	// buildFlag(actionsDo, do, "quiet", "action")
	// buildFlag(actionsDo, do, "chain", "action")
	// buildFlag(actionsDo, do, "services", "action")

	// buildFlag(actionsRemove, do, "file", "action")

	// actionsList.Flags().BoolVarP(&do.Quiet, "quiet", "", false, "machine readable output; also used in tests")
}

//----------------------------------------------------------------------
// cli command wrappers

func StartAgent(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(agents.StartAgents(do))
}

func StopAgent(cmd *cobra.Command, args []string) {
	// IfExit(ArgCheck(2, "eq", cmd, args))
	// do.Name = args[0]
	// do.Path = args[1]
	IfExit(agents.StopAgents(do))
}
