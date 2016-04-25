package commands

import (
	"github.com/eris-ltd/eris-cli/agent"

	. "github.com/eris-ltd/common/go/common"

	"github.com/spf13/cobra"
)

// Primary Agents Sub-Command
var Agents = &cobra.Command{
	Use:   "agent",
	Short: "Start an agent.",
	Long: `Start an agent.
An agent is local server that, when started,  exposes three endpoints:
  
  /chains	=> list running chains on the host (GET)
  /download	=> download a tar'ed contract bundle
  /install	=> download and deploy and tar'ed bundle

The command is used to support the Eris Contracts Library Marketplace.

Please see the pull request for more information about using the agent and its endpoints
 
  https://github.com/eris-ltd/eris-cli/pull/632
  
The agent is stopped with Ctrl+C`,
	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the agent subcommand
func buildAgentsCommand() {
	Agents.AddCommand(agentStart)
	//Agents.AddCommand(agentStop)
	//addAgentsFlags()
}

// start a agent
var agentStart = &cobra.Command{
	Use:   "start",
	Short: "Start the agent.",
	Long: `Start the agent requires to deploy contract bundles
	from the Eris Marketplace.`,
	Run: StartAgent,
}

/*var agentStop = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running agent.",
	Long:  `Stop a running agent.`,
	Run:   StopAgent,
}*/

//----------------------------------------------------------------------
// cli flags
func addAgentsFlags() {
}

//----------------------------------------------------------------------
// cli command wrappers

func StartAgent(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	IfExit(agent.StartAgent(do))
}

/*func StopAgent(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	IfExit(agent.StopAgent(do))
}*/
