package commands

import (
	//	"path/filepath"

	"github.com/eris-ltd/eris-cli/agent"

	. "github.com/eris-ltd/common/go/common"

	"github.com/spf13/cobra"
)

//TODO revisit helpers

// Primary Agents Sub-Command
var Agents = &cobra.Command{
	Use:   "agent",
	Short: "Start and Stop an agent.",
	Long: `Start and Stop an agent.
An agent is used to deploy contract bundles
from the eris library marketplace, The command
requires an account and a registered chain.
Please see (link) for more info.`,

	Run: func(cmd *cobra.Command, args []string) { cmd.Help() },
}

// Build the agent subcommand
func buildAgentsCommand() {
	Agents.AddCommand(agentStart)
	Agents.AddCommand(agentStop)
	addAgentsFlags()
}

// start a agent
var agentStart = &cobra.Command{
	Use:   "start",
	Short: "Start the agent.",
	Long: `Start the agent requires to deploy contract bundles
	from the Eris Marketplace.`,
	Run: StartAgent,
}

var agentStop = &cobra.Command{
	Use:   "stop",
	Short: "Stop a running agent.",
	Long:  `Stop a running agent.`,
	Run:   StopAgent,
}

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

func StopAgent(cmd *cobra.Command, args []string) {
	IfExit(ArgCheck(0, "eq", cmd, args))
	IfExit(agent.StopAgent(do))
}
